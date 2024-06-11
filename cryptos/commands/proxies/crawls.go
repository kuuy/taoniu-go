package proxies

import (
  "context"
  "github.com/urfave/cli/v2"
  "log"
  pool "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/proxies"
)

type CrawlsHandler struct {
  Repository *repositories.CrawlsRepository
}

func NewCrawlsCommand() *cli.Command {
  var h CrawlsHandler
  return &cli.Command{
    Name:  "crawls",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = CrawlsHandler{
        Repository: &repositories.CrawlsRepository{
          Db:  pool.NewDB(1),
          Rdb: pool.NewRedis(1),
          Ctx: context.Background(),
        },
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "request",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Request(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *CrawlsHandler) Request() error {
  log.Println("proxies crawl processing...")

  //url := "https://www.socks-proxy.net/"
  //source := &repositories.CrawlSource{
  //	Url: url,
  //	Headers: map[string]string{
  //		"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
  //	},
  //	HtmlRules: &repositories.HtmlExtractRules{
  //		Container: &repositories.HtmlExtractNode{
  //			Selector: ".fpl-list",
  //		},
  //		List: &repositories.HtmlExtractNode{
  //			Selector: "tbody tr",
  //		},
  //		Fields: []*repositories.HtmlExtractField{
  //			&repositories.HtmlExtractField{
  //				Name: "ip",
  //				Node: &repositories.HtmlExtractNode{
  //					Selector: "td",
  //					Index:    0,
  //				},
  //			},
  //			&repositories.HtmlExtractField{
  //				Name: "port",
  //				Node: &repositories.HtmlExtractNode{
  //					Selector: "td",
  //					Index:    1,
  //				},
  //			},
  //		},
  //	},
  //}

  url := "http://free-proxy.cz/en/proxylist/country/all/socks5/ping/all"
  source := &repositories.CrawlSource{
    Url: url,
    Headers: map[string]string{
      "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
    },
    UseProxy: true,
    Timeout:  10,
    HtmlRules: &repositories.HtmlExtractRules{
      Container: &repositories.HtmlExtractNode{
        Selector: "#proxy_list",
      },
      List: &repositories.HtmlExtractNode{
        Selector: "tbody tr",
      },
      Fields: []*repositories.HtmlExtractField{
        &repositories.HtmlExtractField{
          Name: "ip",
          Node: &repositories.HtmlExtractNode{
            Selector: "td",
            Index:    0,
          },
        },
        &repositories.HtmlExtractField{
          Name: "port",
          Node: &repositories.HtmlExtractNode{
            Selector: "td",
            Index:    1,
          },
        },
      },
    },
  }

  err := h.Repository.Request(source)
  if err != nil {
    log.Println("error", err)
  }

  return nil
}
