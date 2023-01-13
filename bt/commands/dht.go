package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/hktalent/dht"
	"github.com/urfave/cli/v2"
	"taoniu.local/bt/repositories"
)

type file struct {
	Path   []interface{} `json:"path"`
	Length int           `json:"length"`
}

type bitTorrent struct {
	InfoHash string `json:"infohash"`
	Name     string `json:"name"`
	Files    []file `json:"files,omitempty"`
	Length   int    `json:"length,omitempty"`
}

type DhtHandler struct {
	Repository *repositories.DhtRepository
}

func NewDhtCommands() *cli.Command {
	var h DhtHandler
	return &cli.Command{
		Name:  "dht",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = DhtHandler{}
			h.Repository = &repositories.DhtRepository{}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "crawl",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Crawl(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *DhtHandler) Crawl() error {
	w := dht.NewWire(65536, 1024, 256)
	go func() {
		for resp := range w.Response() {
			metadata, err := dht.Decode(resp.MetadataInfo)
			if err != nil {
				continue
			}
			info := metadata.(map[string]interface{})

			if _, ok := info["name"]; !ok {
				continue
			}

			bt := bitTorrent{
				InfoHash: hex.EncodeToString(resp.InfoHash),
				Name:     info["name"].(string),
			}

			if v, ok := info["files"]; ok {
				files := v.([]interface{})
				bt.Files = make([]file, len(files))

				for i, item := range files {
					f := item.(map[string]interface{})
					bt.Files[i] = file{
						Path:   f["path"].([]interface{}),
						Length: f["length"].(int),
					}
				}
			} else if _, ok := info["length"]; ok {
				bt.Length = info["length"].(int)
			}

			data, err := json.Marshal(bt)
			if err == nil {
				fmt.Printf("%s\n\n", data)
			}
		}
	}()
	go w.Run()

	var d *dht.DHT

	config := dht.NewCrawlConfig()
	config.Network = "udp6"
	config.PublicIp = "240e:380:1b68:8300:d401:b838:b42:31d6"
	config.PacketWorkerLimit = 2560
	config.OnAnnouncePeer = func(infoHash, ip string, port int) {
		w.Request([]byte(infoHash), ip, port)
		if infoHash == d.LocalNodeId && ip != d.Config.PublicIp {
			fmt.Printf("找到 : %s:%d\n", ip, port)
		}
	}

	d = dht.New(config)
	d.OnGetPeersResponse = func(infoHash string, peer *dht.Peer) {
		if infoHash == d.LocalNodeId {
			fmt.Printf("my private net: <%s:%d>\n", peer.IP, peer.Port)
		}
	}
	d.AnnouncePeer(d.LocalNodeId)
	d.Log("wait join DHT net for 1 ~ 2 minute ...")
	d.Run()

	return h.Repository.Crawl()
}
