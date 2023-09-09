package stream

import (
	"log"
	"net/url"
	"strconv"

	"github.com/Awareness-Labs/rainforest/pkg/config"
	"github.com/nats-io/nats-server/v2/server"
)

type StreamServer struct {
	natsServer *server.Server
}

type StreamServerConfig struct {
	Port       string   `mapstructure:"port"`
	Domain     string   `mapstructure:"domain"`
	HubURLs    []string `mapstructure:"hub-urls"`
	LeafPort   int      `mapstructure:"leaf-port"`
	StreamPath string   `mapstructure:"stream-path"`
	KVPath     string   `mapstructure:"kv-path"`
	SinkPath   string   `mapstructure:"sink-path"`
}

func NewStreamServer(cfg config.Config) *StreamServer {
	routes := []*url.URL{}
	for _, hub := range cfg.HubURLs {
		routes = append(routes, &url.URL{
			Host: hub,
		})
	}
	port, _ := strconv.Atoi(cfg.Port)
	s, err := server.NewServer(&server.Options{
		Port:            port,
		JetStream:       true,
		JetStreamDomain: cfg.Domain,
		StoreDir:        cfg.StreamPath,
		LeafNode: server.LeafNodeOpts{
			Host: "0.0.0.0",
			// Port: cfg.LeafPort, not use in lead server
			Remotes: []*server.RemoteLeafOpts{
				{
					URLs: routes,
				},
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	s.ConfigureLogger()

	return &StreamServer{
		natsServer: s,
	}
}

func (s *StreamServer) Start() {
	s.natsServer.Start()
}
