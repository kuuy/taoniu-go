package proxies

import (
  "context"
  "errors"
  "fmt"
  "io/ioutil"
  "net"
  "net/http"
  "strconv"
  "strings"
  "time"

  "h12.io/socks"

  "github.com/go-redis/redis/v8"
)

type TorRepository struct {
  Rdb *redis.Client
  Ctx context.Context
}

type TorSession struct {
  Port int
}

func (s *TorSession) DialContext(ctx context.Context, net, addr string) (net.Conn, error) {
  proxy := fmt.Sprintf(
    "socks5://127.0.0.1:%d?timeout=5s",
    s.Port,
  )
  dialer := socks.Dial(proxy)
  return dialer(net, addr)
}

func (r *TorRepository) Failed(port int) error {
  r.Rdb.ZIncrBy(r.Ctx, "proxies:tor:failed", 1, strconv.Itoa(port))
  return nil
}

func (r *TorRepository) Add(port int) error {
  r.Rdb.SAdd(r.Ctx, "proxies:tor:pool", port)
  return nil
}

func (r *TorRepository) Online(port int) error {
  r.Rdb.SAdd(r.Ctx, "proxies:tor:online", port)
  r.Rdb.SRem(r.Ctx, "proxies:tor:offline", port)
  return nil
}

func (r *TorRepository) Offline(port int) error {
  r.Rdb.SAdd(r.Ctx, "proxies:tor:offline", port)
  r.Rdb.SRem(r.Ctx, "proxies:tor:online", port)
  return nil
}

func (r *TorRepository) Start() error {
  starttime := time.Now().Unix()
  var err error
  for {
    err = r.Cmd(
      "tor start",
      "starting tor server",
    )
    if err == nil {
      break
    }
    if time.Now().Unix()-30 > starttime {
      break
    }
    time.Sleep(1 * time.Second)
  }
  r.Rdb.Del(r.Ctx, "proxies:tor:online")
  r.Rdb.Del(r.Ctx, "proxies:tor:offline")
  ports, _ := r.Rdb.SMembers(r.Ctx, "proxies:tor:pool").Result()
  for _, port := range ports {
    r.Rdb.SAdd(r.Ctx, "proxies:tor:offline", port)
  }

  return err
}

func (r *TorRepository) Stop() error {
  return r.Cmd(
    "tor stop",
    "stopping tor server",
  )
}

func (r *TorRepository) Checker(port int) error {
  session := TorSession{port}
  tr := &http.Transport{
    DialContext:       session.DialContext,
    DisableKeepAlives: true,
  }
  httpClient := &http.Client{
    Transport: tr,
  }
  url := "http://httpbin.org/get"
  resp, err := httpClient.Get(url)
  if err != nil {
    return err
  }
  if resp.StatusCode != http.StatusOK {
    return errors.New(
      fmt.Sprintf(
        "response status invalid code[%d]",
        resp.StatusCode,
      ),
    )
  }
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return err
  }
  if !strings.Contains(string(body), url) {
    return errors.New("response body invalid")
  }
  r.Online(port)
  r.Rdb.ZRem(r.Ctx, "proxies:tor:checker", port)

  return nil
}

func (r *TorRepository) ChangeIp(port int) error {
  return r.Cmd(
    fmt.Sprintf("changeip onceport %d", port),
    "changing ip once for port",
  )
}

func (r *TorRepository) Cmd(cmd string, want string) error {
  conn, err := net.Dial("tcp", "127.0.0.1:14999")
  if err != nil {
    return err
  }
  defer conn.Close()
  reply := make([]byte, 1024)
  _, err = conn.Read(reply)
  if err != nil {
    return err
  }
  if !strings.Contains(string(reply), "IPCHANGER API") {
    return errors.New("reply not valid")
  }

  _, err = conn.Write([]byte(fmt.Sprintf("%s\r\n", cmd)))
  if err != nil {
    return err
  }
  _, err = conn.Read(reply)
  if err != nil {
    return errors.New("reply not valid")
  }

  return nil
}
