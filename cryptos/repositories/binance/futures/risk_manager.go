package futures

import (
  "context"
  "fmt"
  "strconv"
  "strings"
  "sync"
  "time"

  "github.com/redis/go-redis/v9"
  "gorm.io/gorm"
)

// PredictionResult 风险预测结果
type PredictionResult struct {
  Direction  int             // 1=LONG, 2=SHORT, 0=NEUTRAL
  Confidence float64         // -1.0 ~ 1.0
  Quantity   float64         // 建议仓位
  Reasons    []string        // 预测理由
  Factors    FactorBreakdown // 各因子得分详情
}

// FactorBreakdown 因子分解
type FactorBreakdown struct {
  TrendScore    float64 // 趋势因子
  FundingScore  float64 // 资金费率因子
  VolumeScore   float64 // 成交量因子
  StrategyScore float64 // 策略共振因子
}

// StrategySignal 策略信号
type StrategySignal struct {
  Name       string
  Direction  int     // 1=LONG, 2=SHORT, 0=NEUTRAL
  Weight     float64 // 权重
  Confidence float64 // 策略自身置信度 0-1
}

// RiskManagerRepository 风险预测管理仓库
type RiskManagerRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context

  // 策略配置
  strategyWeights map[string]float64
  mu              sync.RWMutex

  // 日内风控
  dailyLossMap map[string]float64 // symbol -> 当日亏损
  dailyLossMu  sync.RWMutex

  // 参数配置
  MinConfidence  float64 // 最小置信度阈值
  MaxDailyLoss   float64 // 最大日亏损 (%)
  HighConfidence float64 // 高置信度阈值
  LowConfidence  float64 // 低置信度阈值
}

// NewRiskManagerRepository 创建风险预测仓库
func NewRiskManagerRepository(db *gorm.DB, rdb *redis.Client, ctx context.Context) *RiskManagerRepository {
  return &RiskManagerRepository{
    Db:              db,
    Rdb:             rdb,
    Ctx:             ctx,
    strategyWeights: make(map[string]float64),
    dailyLossMap:    make(map[string]float64),
    MinConfidence:   0.3,   // 30%置信度才交易
    MaxDailyLoss:    -0.05, // 日亏5%停止
    HighConfidence:  0.7,   // >70%加仓
    LowConfidence:   0.4,   // <40%减仓
  }
}

// SetStrategyWeight 设置策略权重
func (r *RiskManagerRepository) SetStrategyWeight(strategy string, weight float64) {
  r.mu.Lock()
  defer r.mu.Unlock()
  r.strategyWeights[strategy] = weight
}

// GetStrategyWeight 获取策略权重
func (r *RiskManagerRepository) GetStrategyWeight(strategy string) float64 {
  r.mu.RLock()
  defer r.mu.RUnlock()
  if w, ok := r.strategyWeights[strategy]; ok {
    return w
  }
  return 1.0 // 默认权重
}

// PredictDirection 预测方向 - 多因子聚合
func (r *RiskManagerRepository) PredictDirection(symbol string, interval string, signals []StrategySignal) (*PredictionResult, error) {
  result := &PredictionResult{
    Direction:  0,
    Confidence: 0,
    Reasons:    make([]string, 0),
    Factors:    FactorBreakdown{},
  }

  // 1. 策略共振因子
  strategyScore, _, strategyReasons := r.calculateStrategyFactor(signals)
  result.Factors.StrategyScore = strategyScore

  // 2. 趋势因子 (基于ZLEMA多周期)
  trendScore, _, trendReasons := r.calculateTrendFactor(symbol, interval)
  result.Factors.TrendScore = trendScore

  // 3. 资金费率因子 (反转信号)
  fundingScore, fundingReason := r.calculateFundingFactor(symbol)
  result.Factors.FundingScore = fundingScore

  // 4. 成交量因子
  volumeScore, volumeReason := r.calculateVolumeFactor(symbol, interval)
  result.Factors.VolumeScore = volumeScore

  // 计算综合得分 (带权重)
  finalScore := strategyScore*0.4 + trendScore*0.3 + fundingScore*0.2 + volumeScore*0.1

  // 确定方向
  if finalScore > 0.2 {
    result.Direction = 1 // LONG
  } else if finalScore < -0.2 {
    result.Direction = 2 // SHORT
  } else {
    result.Direction = 0 // NEUTRAL
  }

  result.Confidence = finalScore
  if result.Confidence < 0 {
    result.Confidence = -result.Confidence
  }

  // 生成理由
  result.Reasons = append(result.Reasons, strategyReasons...)
  result.Reasons = append(result.Reasons, trendReasons...)
  result.Reasons = append(result.Reasons, fundingReason)
  result.Reasons = append(result.Reasons, volumeReason)

  // 缓存预测结果
  r.cachePrediction(symbol, result)

  return result, nil
}

// calculateStrategyFactor 策略共振计算
func (r *RiskManagerRepository) calculateStrategyFactor(signals []StrategySignal) (score float64, direction int, reasons []string) {
  if len(signals) == 0 {
    return 0, 0, []string{"无策略信号"}
  }

  var longScore, shortScore, totalWeight float64

  for _, sig := range signals {
    weight := sig.Weight
    if weight == 0 {
      weight = r.GetStrategyWeight(sig.Name)
    }

    totalWeight += weight

    switch sig.Direction {
    case 1: // LONG
      longScore += weight * sig.Confidence
    case 2: // SHORT
      shortScore += weight * sig.Confidence
    }
  }

  if totalWeight == 0 {
    return 0, 0, []string{"策略权重为0"}
  }

  longRatio := longScore / totalWeight
  shortRatio := shortScore / totalWeight

  // 计算净得分
  score = (longRatio - shortRatio)

  if score > 0.2 {
    direction = 1
    reasons = append(reasons, fmt.Sprintf("策略共振做多 (得分: %.2f)", score))
  } else if score < -0.2 {
    direction = 2
    score = -score
    reasons = append(reasons, fmt.Sprintf("策略共振做空 (得分: %.2f)", score))
  } else {
    direction = 0
    reasons = append(reasons, fmt.Sprintf("策略分歧 (得分: %.2f)", score))
  }

  return score, direction, reasons
}

// calculateTrendFactor 趋势因子计算
func (r *RiskManagerRepository) calculateTrendFactor(symbol string, interval string) (score float64, direction int, reasons []string) {
  // 读取多周期趋势
  shortTrend := r.getTrendDirection(symbol, interval, 20)  // 短周期
  mediumTrend := r.getTrendDirection(symbol, interval, 50) // 中周期
  longTrend := r.getTrendDirection(symbol, interval, 100)  // 长周期

  // 趋势一致性评分
  consensus := 0
  trendValue := shortTrend*0.5 + mediumTrend*0.3 + longTrend*0.2

  if shortTrend > 0 && mediumTrend > 0 {
    consensus++
  }
  if mediumTrend > 0 && longTrend > 0 {
    consensus++
  }
  if shortTrend > 0 && longTrend > 0 {
    consensus++
  }

  score = trendValue * (1 + float64(consensus)*0.2)

  if trendValue > 0.1 {
    direction = 1
    reasons = append(reasons, fmt.Sprintf("趋势向上 (短/中/长: %.2f/%.2f/%.2f)", shortTrend, mediumTrend, longTrend))
  } else if trendValue < -0.1 {
    direction = 2
    score = -score
    reasons = append(reasons, fmt.Sprintf("趋势向下 (短/中/长: %.2f/%.2f/%.2f)", shortTrend, mediumTrend, longTrend))
  } else {
    direction = 0
    reasons = append(reasons, fmt.Sprintf("趋势震荡 (短/中/长: %.2f/%.2f/%.2f)", shortTrend, mediumTrend, longTrend))
  }

  return score, direction, reasons
}

// getTrendDirection 获取趋势方向 (基于ZLEMA)
func (r *RiskManagerRepository) getTrendDirection(symbol string, interval string, period int) float64 {
  key := fmt.Sprintf("binance:futures:indicators:%s:%s:zlema%d", interval, symbol, period)
  val, err := r.Rdb.Get(r.Ctx, key).Result()
  if err != nil || val == "" {
    return 0
  }

  current, err := strconv.ParseFloat(val, 64)
  if err != nil {
    return 0
  }

  // 获取前一期值计算斜率
  prevKey := fmt.Sprintf("binance:futures:indicators:%s:%s:zlema%d:prev", interval, symbol, period)
  prevVal, _ := r.Rdb.Get(r.Ctx, prevKey).Result()
  if prevVal == "" {
    return 0
  }

  prev, err := strconv.ParseFloat(prevVal, 64)
  if err != nil || prev == 0 {
    return 0
  }

  // 归一化趋势强度 (-1 ~ 1)
  return (current - prev) / prev * 10
}

// calculateFundingFactor 资金费率因子
func (r *RiskManagerRepository) calculateFundingFactor(symbol string) (score float64, reason string) {
  // 极端资金费率产生反向信号
  key := fmt.Sprintf("binance:futures:funding:%s", symbol)
  val, err := r.Rdb.Get(r.Ctx, key).Result()
  if err != nil {
    return 0, "资金费率数据不可用"
  }

  fundingRate, err := strconv.ParseFloat(val, 64)
  if err != nil {
    return 0, "资金费率解析失败"
  }

  // 极高正费率 -> 做空机会 (很多人做多，会爆仓)
  if fundingRate > 0.01 { // >1%
    score = -0.8
    return score, fmt.Sprintf("资金费率过高 %.4f，反向做空", fundingRate)
  }
  // 极高负费率 -> 做多机会
  if fundingRate < -0.01 { // <-1%
    score = 0.8
    return score, fmt.Sprintf("资金费率过低 %.4f%，反向做多", fundingRate)
  }

  // 温和费率，轻微反向
  score = -fundingRate * 10
  return score, fmt.Sprintf("资金费率 %.4f，轻微反向信号", fundingRate)
}

// calculateVolumeFactor 成交量因子
func (r *RiskManagerRepository) calculateVolumeFactor(symbol string, interval string) (score float64, reason string) {
  // 放量上涨/缩量下跌 -> 做多
  // 放量下跌/缩量上涨 -> 做空
  volumeKey := fmt.Sprintf("binance:futures:klines:%s:%s:volume", interval, symbol)
  priceKey := fmt.Sprintf("binance:futures:klines:%s:%s:close", interval, symbol)

  volStr, err1 := r.Rdb.Get(r.Ctx, volumeKey).Result()
  priceStr, err2 := r.Rdb.Get(r.Ctx, priceKey).Result()

  if err1 != nil || err2 != nil {
    return 0, "成交量/价格数据不可用"
  }

  volume, _ := strconv.ParseFloat(volStr, 64)
  price, _ := strconv.ParseFloat(priceStr, 64)

  if volume == 0 || price == 0 {
    return 0, "成交量/价格数据无效"
  }

  // 获取历史成交量均值
  avgKey := fmt.Sprintf("binance:futures:indicators:%s:%s:volume:avg20", interval, symbol)
  avgStr, _ := r.Rdb.Get(r.Ctx, avgKey).Result()
  avgVolume, _ := strconv.ParseFloat(avgStr, 64)

  if avgVolume == 0 {
    return 0, "成交量均值无效"
  }

  volumeRatio := volume / avgVolume

  // 获取近期价格变化计算方向
  trend := r.getTrendDirection(symbol, interval, 5)

  if trend > 0 && volumeRatio > 1.2 {
    score = 0.6
    return score, fmt.Sprintf("放量上涨 %.2fx，确认趋势", volumeRatio)
  } else if trend < 0 && volumeRatio > 1.2 {
    score = -0.6
    return score, fmt.Sprintf("放量下跌 %.2fx，确认趋势", volumeRatio)
  } else if volumeRatio < 0.8 {
    score = trend * 0.3 // 缩量跟随趋势
    return score, fmt.Sprintf("缩量 %.2fx，趋势可能延续", volumeRatio)
  }

  return 0, fmt.Sprintf("成交量正常 %.2fx", volumeRatio)
}

// GetRecommendedQuantity 获取建议仓位
func (r *RiskManagerRepository) GetRecommendedQuantity(symbol string, baseQty float64, atr float64, confidence float64) float64 {
  // 1. 检查是否触发止损线
  r.dailyLossMu.RLock()
  dailyLoss := r.dailyLossMap[symbol]
  r.dailyLossMu.RUnlock()

  if dailyLoss < r.MaxDailyLoss {
    return 0 // 停止交易
  }

  // 2. ATR倍数调整 (波动率控制)
  var atrMultiplier float64
  switch {
  case atr > 0.05: // 高波动 >5%
    atrMultiplier = 0.3
  case atr > 0.03: // 中波动 3-5%
    atrMultiplier = 0.6
  case atr > 0.02: // 低波动 2-3%
    atrMultiplier = 0.8
  default: // 极低波动 <2%
    atrMultiplier = 1.0
  }

  // 3. 置信度调整
  var confidenceMultiplier float64
  switch {
  case confidence > r.HighConfidence:
    confidenceMultiplier = 1.2 // 高置信度，加仓
  case confidence > r.MinConfidence:
    confidenceMultiplier = 1.0 // 正常
  case confidence > r.LowConfidence:
    confidenceMultiplier = 0.5 // 低置信度，减仓
  default:
    confidenceMultiplier = 0.0 // 禁止交易
  }

  finalQty := baseQty * atrMultiplier * confidenceMultiplier

  // 4. 最大仓位限制
  maxQty := baseQty * 1.5
  if finalQty > maxQty {
    finalQty = maxQty
  }

  // 5. 最小仓位
  minQty := baseQty * 0.1
  if finalQty < minQty && finalQty > 0 {
    finalQty = minQty
  }

  return finalQty
}

// UpdateDailyLoss 更新日内亏损
func (r *RiskManagerRepository) UpdateDailyLoss(symbol string, pnl float64) {
  r.dailyLossMu.Lock()
  defer r.dailyLossMu.Unlock()
  r.dailyLossMap[symbol] += pnl
}

// GetDailyLoss 获取日内亏损
func (r *RiskManagerRepository) GetDailyLoss(symbol string) float64 {
  r.dailyLossMu.RLock()
  defer r.dailyLossMu.RUnlock()
  return r.dailyLossMap[symbol]
}

// ResetDailyLoss 重置日内亏损 (每日调用)
func (r *RiskManagerRepository) ResetDailyLoss() {
  r.dailyLossMu.Lock()
  defer r.dailyLossMu.Unlock()
  r.dailyLossMap = make(map[string]float64)
}

// ShouldTrade 判断是否应该交易
func (r *RiskManagerRepository) ShouldTrade(symbol string, confidence float64) bool {
  // 检查日内止损
  r.dailyLossMu.RLock()
  dailyLoss := r.dailyLossMap[symbol]
  r.dailyLossMu.RUnlock()

  if dailyLoss < r.MaxDailyLoss {
    return false
  }

  // 检查置信度
  return confidence >= r.MinConfidence
}

// cachePrediction 缓存预测结果到Redis
func (r *RiskManagerRepository) cachePrediction(symbol string, result *PredictionResult) {
  key := fmt.Sprintf("binance:futures:risk:prediction:%s", symbol)
  data := fmt.Sprintf("%d|%.4f|%.4f|%s|%d",
    result.Direction,
    result.Confidence,
    result.Quantity,
    strings.Join(result.Reasons, ";"),
    time.Now().Unix(),
  )
  r.Rdb.SetEx(r.Ctx, key, data, 5*time.Minute)
}

// GetCachedPrediction 获取缓存的预测
func (r *RiskManagerRepository) GetCachedPrediction(symbol string) (*PredictionResult, error) {
  key := fmt.Sprintf("binance:futures:risk:prediction:%s", symbol)
  val, err := r.Rdb.Get(r.Ctx, key).Result()
  if err != nil {
    return nil, err
  }

  parts := strings.Split(val, "|")
  if len(parts) < 4 {
    return nil, fmt.Errorf("invalid prediction format")
  }

  direction, _ := strconv.Atoi(parts[0])
  confidence, _ := strconv.ParseFloat(parts[1], 64)
  quantity, _ := strconv.ParseFloat(parts[2], 64)
  reasons := strings.Split(parts[3], ";")

  return &PredictionResult{
    Direction:  direction,
    Confidence: confidence,
    Quantity:   quantity,
    Reasons:    reasons,
  }, nil
}

// DefaultStrategyWeights 默认策略权重配置
func (r *RiskManagerRepository) DefaultStrategyWeights() {
  r.mu.Lock()
  defer r.mu.Unlock()

  // SMC策略权重最高 (结构分析更可靠)
  r.strategyWeights["smc"] = 2.0        // Smart Money Concepts
  r.strategyWeights["supertrend"] = 1.5 // SuperTrend趋势
  r.strategyWeights["rsistoch"] = 1.0   // RSI+KDJ
  r.strategyWeights["zlema"] = 1.0      // ZLEMA均线
  r.strategyWeights["bbands"] = 1.0     // 布林带
}

// GetAllFactors 获取所有因子详情 (用于调试)
func (r *RiskManagerRepository) GetAllFactors(symbol string, interval string) map[string]interface{} {
  trendScore, _, _ := r.calculateTrendFactor(symbol, interval)
  fundingScore, _ := r.calculateFundingFactor(symbol)
  volumeScore, _ := r.calculateVolumeFactor(symbol, interval)

  return map[string]interface{}{
    "trend_score":   trendScore,
    "funding_score": fundingScore,
    "volume_score":  volumeScore,
    "daily_loss":    r.GetDailyLoss(symbol),
    "confidence_thresholds": map[string]float64{
      "min":  r.MinConfidence,
      "low":  r.LowConfidence,
      "high": r.HighConfidence,
    },
  }
}
