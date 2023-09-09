package main

import (
	"runtime"
	"strings"

	"github.com/Awareness-Labs/rainforest/pkg/config"
	"github.com/Awareness-Labs/rainforest/pkg/consumer"
	"github.com/Awareness-Labs/rainforest/pkg/processor/kv"
	"github.com/Awareness-Labs/rainforest/pkg/server"
	"github.com/Awareness-Labs/rainforest/pkg/stream"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {

	pflag.String("port", "4222", "port to serve on")
	pflag.String("domain", "", "domain of the rainforest server (required)")
	pflag.StringArray("hub-urls", []string{}, "remote connection hub URLs")
	pflag.Int("leaf-port", 7422, "leaf port to start")
	pflag.String("stream-path", "./data/stream", "directory to store stream data")
	pflag.String("kv-path", "", "directory to store key value data")
	pflag.String("sink-path", "", "directory to sink event data for OLAP to query")

	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log.Info().Msg(">>> Rainforest is the ultra light-weight Data Mesh <<<")
	// Start embedded Stream Server
	cfg := config.Config{}

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Error().Msg(err.Error())
	}

	if strings.TrimSpace(cfg.Domain) == "" {
		log.Error().Msg("domain cannot be null")
	}

	// Start NATS, JetStream embedded server
	strServer := stream.NewStreamServer(cfg)
	strServer.Start()

	// Connect to NATS
	nc, err := nats.Connect("localhost:" + cfg.Port)
	if err != nil {
		log.Error().Msg(err.Error())
	}

	// Start Rainforest server
	rfServer := server.NewServer(nc, cfg)
	rfServer.Start()

	// Start KV consumer
	// kv := consumer.NewKeyValueConsumer(nc, cfg.KVPath)
	// go kv.Start()

	kv := kv.NewKVProcessor(nc, cfg.KVPath)
	go kv.Start()

	// Start Sink consumer
	sink := consumer.NewSinkCunsumer(nc, cfg.SinkPath)
	go sink.Start()

	// Wait to stop
	runtime.Goexit()
	defer nc.Close()
}
