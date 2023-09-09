package processor

import (
	"context"
	"strings"

	"github.com/nats-io/nats.go/jetstream"
)

type Processor interface {
	Start()                          // Load state for sched eg: jetstream
	Subscribe()                      // Subscribe subject (sync)
	Process(cons jetstream.Consumer) // Consume the messages
}

func GetDataProductStream(js jetstream.JetStream, prodType string) []string {
	ctx := context.TODO()
	names := js.StreamNames(ctx)
	streams := []string{}
	for name := range names.Name() {
		if strings.HasPrefix(name, prodType) {
			streams = append(streams, name)
		}
	}
	return streams
}
