package kv

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"strings"

	"github.com/Awareness-Labs/rainforest/pkg/processor"
	apiv1 "github.com/Awareness-Labs/rainforest/pkg/proto/api/v1"
	"github.com/Awareness-Labs/rainforest/pkg/storage"
	"github.com/Awareness-Labs/rainforest/pkg/storage/atomic"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/encoding/protojson"
)

type KVProcessor struct {
	conn          *nats.Conn
	streamManager jetstream.JetStream
	db            storage.AtomicStorage
	dir           string
}

func NewKVProcessor(nc *nats.Conn, dir string) processor.Processor {
	db := atomic.NewBadger(dir)

	js, err := jetstream.New(nc)
	if err != nil {
		log.Error().Msg(err.Error())
	}

	return &KVProcessor{
		streamManager: js,
		db:            db,
		conn:          nc,
		dir:           dir,
	}

}

func (p *KVProcessor) Start() {
	ctx := context.TODO()
	p.Subscribe()

	deployedConsumer := map[string]struct{}{}
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		streams := processor.GetDataProductStream(p.streamManager, "STATE_")
		for _, stream := range streams {
			_, ok := deployedConsumer[stream]
			if ok {
				continue
			}
			cons, err := p.streamManager.OrderedConsumer(ctx, stream, jetstream.OrderedConsumerConfig{})
			if err != nil {
				log.Error().Msg(err.Error())
				return
			}
			go p.Process(cons)
			deployedConsumer[stream] = struct{}{}
		}
	}

}

func (p *KVProcessor) Subscribe() {
	p.conn.Subscribe("$RAINFOREST.API.KV.*", func(msg *nats.Msg) {

		req := &apiv1.KeyValueRequest{}
		err := protojson.Unmarshal(msg.Data, req)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}
		log.Info().Msg(req.String())

		switch req.GetOperation().(type) {
		case *apiv1.KeyValueRequest_Scan:
			kvs, err := p.db.Scan(
				[]byte(req.GetScan().GetStartKey()),
				[]byte(req.GetScan().GetEndKey()),
				false,
				int(req.GetScan().GetLimit()),
			)

			if err != nil {
				sendErrorResponse(msg, err)
			}

			res := &apiv1.KeyValueDataResponse{
				Kvs: kvs,
			}
			resData, err := protojson.Marshal(res)
			if err != nil {
				sendErrorResponse(msg, err)
			}

			msg.Respond(resData)

		case *apiv1.KeyValueRequest_Get:
			msg.Respond([]byte("not support get op now"))
		}

	})
}

func (p *KVProcessor) Process(cons jetstream.Consumer) {
	for {
		msgs, err := cons.Fetch(1)
		if msgs.Error() != nil {
			// handle error
			log.Error().Msg(err.Error())
			continue
		}
		for msg := range msgs.Messages() {

			k := strings.TrimPrefix(msg.Subject(), "$RAINFOREST.DP.STATE.")
			k = strings.Replace(k, ".", "/", -1)
			fmt.Println("k:", []byte(k), "v:", msg.Data())

			// Storage engine
			p.db.Set([]byte(k), msg.Data())

			msg.Ack()
		}

	}
}

func sendErrorResponse(m *nats.Msg, err error) {
	res := &apiv1.KeyValueResponse{
		Response: &apiv1.KeyValueResponse_ErrorResponse{
			ErrorResponse: &apiv1.ErrorResponse{
				ErrorCode:    "400",
				ErrorMessage: err.Error(),
			},
		},
	}
	resData, _ := json.Marshal(res)
	m.Respond(resData)
}
