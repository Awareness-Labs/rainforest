package server

import (
	"context"
	"errors"

	"time"

	"github.com/Awareness-Labs/rainforest/pkg/config"
	apiv1 "github.com/Awareness-Labs/rainforest/pkg/proto/api/v1"
	corev1 "github.com/Awareness-Labs/rainforest/pkg/proto/core/v1"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/encoding/protojson"
)

// TODU: log

const (
	StateDataProductSubjectPrefix  = "$RAINFOREST.DP.STATE."
	EventDataProductSubjectPrefix  = "$RAINFOREST.DP.EVENT."
	SourceDataProductSubjectPrefix = "$RAINFOREST.DP.SOURCE."

	StateDataProductPrefix  = "STATE_"
	EventDataProductPrefix  = "EVENT_"
	SourceDataProductPrefix = "SOURCE_"
)

type Server struct {
	conn          *nats.Conn
	streamManager jetstream.JetStream
	kvManager     nats.JetStreamContext
	objManager    nats.JetStreamContext
}

func NewServer(nc *nats.Conn, cfg config.Config) *Server {

	js, err := jetstream.New(nc)
	if err != nil {
		log.Printf("Failed to initialize JetStream: %v", err)
		return nil
	}

	kv, err := nc.JetStream()
	if err != nil {
		log.Printf("Failed to initialize Key-Value Manager: %v", err)
		return nil
	}

	obj, err := nc.JetStream()
	if err != nil {
		log.Printf("Failed to initialize Object Manager: %v", err)
		return nil
	}

	return &Server{
		conn:          nc,
		streamManager: js,
		kvManager:     kv,
		objManager:    obj,
	}
}

func (s *Server) Start() {

	apis := map[string]nats.MsgHandler{
		"$RAINFOREST.API.DP.CREATE.*": s.CreateDataProduct,
		"$RAINFOREST.API.DP.INFO.*":   s.InfoDataProduct,
		"$RAINFOREST.API.DP.UPDATE.*": s.UpdateDataProduct,
		"$RAINFOREST.API.DP.DELETE.*": s.DeleteDataProduct,
		"$RAINFOREST.API.DP.LIST":     s.ListDataProducts,
	}

	// API Subscribe to DataProduct topics
	for endpoint := range apis {
		log.Info().Msg(endpoint)
		_, err := s.conn.Subscribe(endpoint, apis[endpoint])
		if err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

func (s *Server) CreateDataProduct(m *nats.Msg) {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Unmarshal
	req := &apiv1.CreateDataProductRequest{}
	err := protojson.Unmarshal(m.Data, req)

	if err != nil {
		sendErrorResponse(m, err)
		return
	}

	// Implement create a DataProduct.
	prod := req.GetProduct()

	switch prod.GetType() {
	case corev1.DataProductType_DATA_PRODUCT_TYPE_STATE:
		createStateDataProduct(ctx, prod, s)
		if err != nil {
			sendErrorResponse(m, err)
			return
		}
		sendResponse(m, "data product created: "+prod.GetName())
	case corev1.DataProductType_DATA_PRODUCT_TYPE_EVENT:
		createEventDataProduct(ctx, prod, s)

		if err != nil {
			sendErrorResponse(m, err)
			return
		}
		sendResponse(m, "data product created: "+prod.GetName())

	default:
		sendErrorResponse(m, errors.New("not support this type data product"))
		return
	}

}

func sendErrorResponse(m *nats.Msg, err error) {
	res := &apiv1.ErrorResponse{
		ErrorCode:    "400",
		ErrorMessage: err.Error(),
	}

	resData, _ := protojson.Marshal(res)
	m.Respond(resData)
}

func sendResponse(m *nats.Msg, status string) {
	// res := &apiv1.DataProductResponse{}
	// resData, _ := json.Marshal(res)
	m.Respond([]byte(status))
}

func getSourceDataProduct(product *corev1.DataProduct) []*jetstream.StreamSource {
	sources := []*jetstream.StreamSource{}
	for _, dp := range product.SourceDataProducts {
		// Must align with parents type, or nats will not process
		sources = append(sources, &jetstream.StreamSource{
			Name:   dp.Name,
			Domain: dp.Domain,
		})
	}
	return sources
}

func createStateDataProduct(ctx context.Context, product *corev1.DataProduct, s *Server) error {
	// if this DP has source
	sources := getSourceDataProduct(product)
	_, err := s.streamManager.CreateStream(ctx, jetstream.StreamConfig{
		Name:              StateDataProductPrefix + product.GetName(),
		MaxMsgsPerSubject: 1,
		Subjects:          []string{StateDataProductSubjectPrefix + product.GetName() + ".*"},
		Description:       product.GetDescription(),
		Discard:           jetstream.DiscardOld,
		Sources:           sources,
	})

	if err != nil {
		return err
	}

	return nil
}

func createEventDataProduct(ctx context.Context, product *corev1.DataProduct, s *Server) error {
	sources := getSourceDataProduct(product)
	_, err := s.streamManager.CreateStream(ctx, jetstream.StreamConfig{
		Name:        EventDataProductPrefix + product.GetName(),
		Subjects:    []string{EventDataProductSubjectPrefix + product.GetName() + ".*"},
		Description: product.GetDescription(),
		Sources:     sources,
	})

	if err != nil {
		return err
	}

	return nil
}

// Not MVP feature!

// TODO:
func (s *Server) InfoDataProduct(m *nats.Msg) {
}

// TODO:
func (s *Server) UpdateDataProduct(m *nats.Msg) {
}

// TODO:
func (s *Server) DeleteDataProduct(m *nats.Msg) {
}

// TODO
func (s *Server) ListDataProducts(m *nats.Msg) {
}
