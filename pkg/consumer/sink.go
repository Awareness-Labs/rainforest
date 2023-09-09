package consumer

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Awareness-Labs/rainforest/pkg/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type SinkConsumer struct {
	conn          *nats.Conn
	streamManager jetstream.JetStream
	sinkPath      string
}

type MessageData struct {
	Data []byte
	Msg  jetstream.Msg
}

func NewSinkCunsumer(nc *nats.Conn, sinkPath string) *SinkConsumer {

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal(err)
	}
	return &SinkConsumer{
		conn:          nc,
		streamManager: js,
		sinkPath:      sinkPath,
	}
}

// Consume from EventStream, sink event into json file (time-slice algo)
func (c *SinkConsumer) Start() {

	js, err := jetstream.New(c.conn)
	if err != nil {
		log.Printf("Failed to initialize JetStream: %v", err)
		return
	}
	// list stream names
	go c.PeriodSched(js)

	if err != nil {
		log.Printf("Failed to initialize Consumer: %v", err)
		return
	}
}

func (c *SinkConsumer) PeriodSched(js jetstream.JetStream) {
	deployedConsumer := map[string]bool{}
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		// log.Println("Sched is working bitch!")
		ctx := context.TODO() // Use the appropriate context

		names := c.streamManager.StreamNames(ctx)
		for name := range names.Name() {

			_, ok := deployedConsumer[name]
			if ok {
				continue
			}

			if strings.HasPrefix(name, server.EventDataProductPrefix) {

				cons, err := js.OrderedConsumer(ctx, name, jetstream.OrderedConsumerConfig{})

				if err != nil {
					log.Printf("Failed to initialize Consumer: %v", err)
					return
				}
				go SinkConsumerFunc(cons, c.sinkPath)
				deployedConsumer[name] = true
			}
		}
	}
}

func SinkConsumerFunc(cons jetstream.Consumer, sinkPath string) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	iter, err := cons.Messages()
	if err != nil {
		log.Println(err)
		return
	}

	buffer := make(chan *MessageData, 4096) // Adjusted buffer size to 4096

	// Goroutine to read messages and send to buffer channel
	go func() {
		for {
			msg, err := iter.Next()
			if err != nil {
				log.Println(err)
				continue
			}

			buffer <- &MessageData{Data: msg.Data(), Msg: msg}
		}
	}()

	for {
		select {
		case <-ticker.C:
			select {
			case messageData := <-buffer:
				subject := string(messageData.Msg.Subject())
				subject = strings.TrimPrefix(subject, "$RAINFOREST.DP.EVENT.")
				subject = strings.Split(subject, ".")[0] // This will get 'ConversationEvent' from 'ConversationEvent.1'

				filename := subject + ".json"
				filePath := filepath.Join(sinkPath, filename)

				// Ensure directory exists before writing
				dir := filepath.Dir(filePath)
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					os.MkdirAll(dir, 0755)
				}

				var existingData []json.RawMessage

				// If file exists, read its contents into the existingData slice
				if _, err := os.Stat(filePath); !os.IsNotExist(err) {
					fileBytes, readErr := ioutil.ReadFile(filePath)
					if readErr != nil {
						log.Println("Error reading existing file:", readErr)
						continue
					}

					if len(fileBytes) > 0 {
						unmarshalErr := json.Unmarshal(fileBytes, &existingData)
						if unmarshalErr != nil {
							log.Println("Error unmarshalling existing JSON data:", unmarshalErr)
							continue
						}
					}
				}

				// Append the new message data to the existing data
				existingData = append(existingData, json.RawMessage(messageData.Data))

				// Convert the combined data back to JSON
				combinedData, marshalErr := json.Marshal(existingData)
				if marshalErr != nil {
					log.Println("Error marshaling combined JSON data:", marshalErr)
					continue
				}

				// Write combined data back to the file
				writeErr := ioutil.WriteFile(filePath, combinedData, 0644)
				if writeErr != nil {
					log.Println("Error writing to file:", writeErr)
					continue
				}

				// Acknowledge the message only after it's written to disk
				messageData.Msg.Ack()

			default:
				// No data in the buffer channel
			}
		}
	}
}
