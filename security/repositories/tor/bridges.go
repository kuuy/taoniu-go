package tor

import (
  "bufio"
  "context"
  "encoding/binary"
  "errors"
  "fmt"
  "io"
  "log"
  "math/rand"
  "net"
  "net/http"
  "os"
  "os/exec"
  "strconv"
  "strings"
  "syscall"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"
  "gorm.io/gorm"

  models "taoniu.local/security/models/tor"
)

type BridgesRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *BridgesRepository) Import(path string) error {
  file, err := os.Open(path)
  if err != nil {
    return err
  }
  defer file.Close()
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    content := strings.TrimSpace(scanner.Text())
    data := strings.Split(content, " ")
    if len(data) < 5 {
      continue
    }
    protocol := data[0]
    item := strings.Split(data[1], ":")
    ip := net.ParseIP(item[0])
    if ip == nil {
      continue
    }
    port, err := strconv.Atoi(item[1])
    if err != nil {
      continue
    }
    secret := data[2]
    cert := strings.TrimLeft(data[3], "cert=")
    mode, _ := strconv.Atoi(strings.TrimLeft(data[4], "iat-mode="))

    r.Save(protocol, r.IpToLong(ip), port, secret, cert, mode)
  }

  return nil
}

func (r *BridgesRepository) Flush() error {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }
  session := &net.Dialer{}
  tr.DialContext = session.DialContext
  httpClient := &http.Client{
    Transport: tr,
    Timeout:   60 * time.Second,
  }

  url := "https://raw.githubusercontent.com/scriptzteam/Tor-Bridges-Collector/main/bridges-obfs4"
  req, _ := http.NewRequest("GET", url, nil)
  resp, err := httpClient.Do(req)
  if err != nil {
    return err
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return errors.New(fmt.Sprintf("request error: status[%s] code[%d]", resp.Status, resp.StatusCode))
  }

  body, err := io.ReadAll(resp.Body)
  if err != nil {
    return err
  }

  for _, row := range strings.Split(string(body), "\n") {
    data := strings.Split(row, " ")
    if len(data) < 5 {
      continue
    }
    protocol := data[0]
    item := strings.Split(data[1], ":")
    ip := net.ParseIP(item[0])
    if ip == nil {
      continue
    }
    port, err := strconv.Atoi(item[1])
    if err != nil {
      continue
    }
    secret := data[2]
    cert := strings.TrimLeft(data[3], "cert=")
    mode, _ := strconv.Atoi(strings.TrimLeft(data[4], "iat-mode="))

    r.Save(protocol, r.IpToLong(ip), port, secret, cert, mode)
  }

  return nil
}

func (r *BridgesRepository) Checker() error {
  var entities []*models.Bridge
  r.Db.Where("status", 0).Limit(20).Find(&entities)
  var bridges []string
  for _, entity := range entities {
    bridge := fmt.Sprintf(
      "%s %s:%d %s cert=%s iat-mode=%d",
      entity.Protocol,
      r.LongToIp(entity.Ip),
      entity.Port,
      entity.Secret,
      entity.Cert,
      entity.Mode,
    )
    bridges = append(bridges, bridge)
  }
  if len(bridges) > 0 {
    r.Monitor(0, bridges, true)
  }
  return nil
}

func (r *BridgesRepository) Monitor(id int, bridges []string, isChecker bool) error {
  score, _ := r.Rdb.ZScore(
    r.Ctx,
    "tor:proxies:pids",
    string(id),
  ).Result()
  if score > 0 {
    syscall.Kill(int(score), syscall.SIGKILL)
  }

  port := 9080 + id
  r.Rdb.SRem(r.Ctx, "tor:proxies:ports", port)

  var args []string
  args = append(args, "-f")
  args = append(args, "/data/tor/.torrc")
  for _, bridge := range bridges {
    args = append(args, "--Bridge")
    args = append(args, bridge)
  }
  args = append(args, "--log")
  args = append(args, "notice")
  args = append(args, "--AvoidDiskWrites")
  args = append(args, "1")
  args = append(args, "--SafeLogging")
  args = append(args, "0")
  args = append(args, "--GeoIPExcludeUnknown")
  args = append(args, "1")
  args = append(args, "--AutomapHostsSuffixes")
  args = append(args, ".onion")
  args = append(args, "--AutomapHostsOnResolve")
  args = append(args, "1")
  args = append(args, "--StrictNodes")
  args = append(args, "0")
  args = append(args, "--SocksPort")
  args = append(args, fmt.Sprintf("127.0.0.1:%d", port))
  args = append(args, "--DataDirectory")
  args = append(args, fmt.Sprintf("/opt/tor/data/%02d", id))
  cmd := exec.Command("/usr/local/sbin/tor", args...)
  stdout, err := cmd.StdoutPipe()
  cmd.Stderr = cmd.Stdout
  if err != nil {
    return err
  }
  if err = cmd.Start(); err != nil {
    return err
  }
  pid := cmd.Process.Pid
  defer func() {
    r.Rdb.SRem(r.Ctx, "tor:proxies:ports", port)
    syscall.Kill(pid, syscall.SIGKILL)
  }()
  isConnected := false
  scanner := bufio.NewScanner(stdout)
  for scanner.Scan() {
    content := scanner.Text()
    log.Println(content)
    data := strings.Split(content, " ")
    if strings.Contains(content, "[warn] Proxy Client: unable to connect OR connection (handshaking (proxy)) with") {
      item := strings.Split(data[14], ":")
      ip := net.ParseIP(item[0])
      if ip == nil {
        continue
      }
      port, err := strconv.Atoi(item[1])
      if err != nil {
        continue
      }
      r.Disabled(r.IpToLong(ip), port)
    } else if strings.Contains(content, "[notice] new bridge descriptor") {
      for _, bridge := range bridges {
        item := strings.Split(strings.Split(bridge, " ")[1], ":")
        ip := net.ParseIP(item[0])
        if ip == nil {
          continue
        }
        port, err := strconv.Atoi(item[1])
        if err != nil {
          continue
        }
        if strings.Contains(content, item[0]) {
          r.Enabled(r.IpToLong(ip), port)
        }
      }
    } else if strings.Contains(content, "Bootstrapped 100% (done): Done") {
      log.Println("starting okay")
      isConnected = true
      if isChecker {
        syscall.Kill(pid, syscall.SIGKILL)
        os.Exit(0)
      }
      r.Rdb.SAdd(r.Ctx, "tor:proxies:ports", port)
    } else if strings.Contains(content, "Bootstrapped 0% (starting): Starting") {
      r.Rdb.ZAdd(r.Ctx, "tor:proxies:pids", &redis.Z{
        Score:  float64(pid),
        Member: strconv.Itoa(id),
      })
      log.Println("waiting for starting...")
      go func() {
        startTime := time.Now().Unix()
        for {
          endTime := time.Now().Unix()
          if isConnected {
            break
          }
          if endTime-startTime > 30 {
            for _, bridge := range bridges {
              item := strings.Split(strings.Split(bridge, " ")[1], ":")
              ip := net.ParseIP(item[0])
              if ip == nil {
                continue
              }
              port, err := strconv.Atoi(item[1])
              if err != nil {
                continue
              }
              r.Timeout(r.IpToLong(ip), port)
            }
            log.Println("timeout for starting...")
            syscall.Kill(pid, syscall.SIGKILL)
            os.Exit(0)
          }
          time.Sleep(1 * time.Second)
        }
      }()
    }
  }

  return nil
}

func (r *BridgesRepository) Count(status []int, onlines []string) (int64, error) {
  var count int64
  query := r.Db.Model(&models.Bridge{}).Where("status", status)
  if len(onlines) > 0 {
    query.Not("id", onlines)
  }
  query.Count(&count)
  return count, nil
}

func (r *BridgesRepository) Random(i int, limit int) ([]string, error) {
  port := 9080 + i
  items, _ := r.Rdb.ZRangeByScore(
    r.Ctx,
    "tor:bridges",
    &redis.ZRangeBy{
      Min: fmt.Sprintf("%d", port),
      Max: fmt.Sprintf("%d", port),
    },
  ).Result()
  pipe := r.Rdb.Pipeline()
  for _, item := range items {
    pipe.ZRem(r.Ctx, "tor:bridges", item).Result()
  }
  pipe.Exec(r.Ctx)

  onlines, _ := r.Rdb.ZRevRange(r.Ctx, "tor:bridges", 0, -1).Result()

  count, _ := r.Count([]int{0, 1, 3}, onlines)
  if count < int64(limit) {
    return nil, errors.New("bridges not enough")
  }

  var ids []string
  var bridges []string
  for len(bridges) < limit {
    rand.Seed(time.Now().UnixNano())
    offset := rand.Uint64() % uint64(count)
    var entity models.Bridge
    query := r.Db.Where("status", []int{0, 1, 3})
    if len(onlines) > 0 {
      query.Not("id", onlines)
    }
    result := query.Offset(int(offset)).Take(&entity)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      return nil, errors.New("random failed")
    }
    score, _ := r.Rdb.ZScore(
      r.Ctx,
      "tor:bridges",
      entity.ID,
    ).Result()
    if score > 0 {
      continue
    }
    bridge := fmt.Sprintf(
      "%s %s:%d %s cert=%s iat-mode=%d",
      entity.Protocol,
      r.LongToIp(entity.Ip),
      entity.Port,
      entity.Secret,
      entity.Cert,
      entity.Mode,
    )
    bridges = append(bridges, bridge)
    ids = append(ids, entity.ID)
  }

  for _, id := range ids {
    r.Rdb.ZAdd(r.Ctx, "tor:bridges", &redis.Z{
      Score:  float64(port),
      Member: id,
    })
  }

  return bridges, nil
}

func (r *BridgesRepository) Enabled(ip uint32, port int) error {
  var entity *models.Bridge
  result := r.Db.Where("ip=? AND port=?", ip, port).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return result.Error
  }
  entity.Status = 1
  entity.TimeoutCount = 0
  r.Db.Model(&models.Bridge{ID: entity.ID}).Select([]string{"Status", "TimeoutCount"}).Updates(entity)
  return nil
}

func (r *BridgesRepository) Timeout(ip uint32, port int) error {
  var entity *models.Bridge
  result := r.Db.Where("ip=? AND port=?", ip, port).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return result.Error
  }
  if entity.Status != 2 {
    entity.Status = 3
  }
  entity.TimeoutCount++
  r.Db.Model(&models.Bridge{ID: entity.ID}).Updates(entity)
  return nil
}

func (r *BridgesRepository) Disabled(ip uint32, port int) error {
  r.Db.Model(&models.Bridge{}).Where("ip=? AND port=?", ip, port).Update("status", 2)
  return nil
}

func (r *BridgesRepository) Rescue() error {
  count, _ := r.Count([]int{0}, []string{})
  if count == 0 {
    r.Db.Model(&models.Bridge{}).Where("status", 3).Update("status", 0)
  }
  return nil
}

func (r *BridgesRepository) Show() (*models.Bridge, error) {
  var entity *models.Bridge
  result := r.Db.Where("status", 1).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, errors.New("usable bridge empty")
  }

  return entity, nil
}

func (r *BridgesRepository) Save(
  protocal string,
  ip uint32,
  port int,
  secret string,
  cert string,
  mode int,
) error {
  var entity models.Bridge
  result := r.Db.Where(
    "ip=? AND port=?",
    ip,
    port,
  ).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity = models.Bridge{
      ID:       xid.New().String(),
      Protocol: protocal,
      Ip:       ip,
      Port:     port,
      Secret:   secret,
      Cert:     cert,
      Mode:     mode,
    }
    r.Db.Create(&entity)
  } else {
    entity.Secret = secret
    entity.Cert = cert
    entity.Mode = mode
    r.Db.Model(&models.Bridge{ID: entity.ID}).Updates(entity)
  }
  return nil
}

func (*BridgesRepository) IpToLong(ip net.IP) uint32 {
  if len(ip) == 16 {
    return binary.BigEndian.Uint32(ip[12:16])
  }
  return binary.BigEndian.Uint32(ip)
}

func (*BridgesRepository) LongToIp(nn uint32) net.IP {
  ip := make(net.IP, 4)
  binary.BigEndian.PutUint32(ip, nn)
  return ip
}

func (r *BridgesRepository) contains(s []string, str string) bool {
  for _, v := range s {
    if v == str {
      return true
    }
  }
  return false
}
