package grpc_server

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"Go-Practice-VK/gen"
	"Go-Practice-VK/subpub"
)

type PubSubServer struct {
	gen.UnimplementedPubSubServer
	bus subpub.SubPub
}

func NewPubSubServer(bus subpub.SubPub) *PubSubServer {
	return &PubSubServer{bus: bus}
}

func (s *PubSubServer) Publish(ctx context.Context, req *gen.PublishRequest) (*emptypb.Empty, error) {
	if req.Key == "" || req.Data == "" {
		log.Printf("[ERROR][gRPC][Publish] invalid input: key or data empty, code=%s", codes.InvalidArgument)
		return nil, status.Error(codes.InvalidArgument, "key and data must be non-empty")
	}

	log.Printf("[INFO][gRPC][Publish] key=%s, data=%s", req.Key, req.Data)

	err := s.bus.Publish(req.Key, req.Data)
	if err != nil {
		log.Printf("[WARN][gRPC][Publish] failed to publish %v, code: %s", err, codes.Internal)
		return nil, status.Error(codes.Internal, "failed to publish message")
	}

	return &emptypb.Empty{}, err
}

func (s *PubSubServer) Subscribe(req *gen.SubscribeRequest, server gen.PubSub_SubscribeServer) error {
	if req.Key == "" {
		log.Printf("[ERROR][gRPC][Subscribe] empty key, code=%s", codes.InvalidArgument)
		return status.Error(codes.InvalidArgument, "Sub Key must be non-empty")
	}

	peerInfo, ok := peer.FromContext(server.Context())
	if ok {
		log.Printf("[INFO][gRPC][Subscribe] client=%s subscribed to key=%s", peerInfo.Addr.String(), req.Key)
	} else {
		log.Printf("[INFO][gRPC][Subscribe] unknown client subscribed to key=%s", req.Key)
	}

	var sub subpub.Subscription

	handler := func(msg interface{}) {
		data, ok := msg.(string)
		if !ok {
			log.Printf("[WARN][gRPC][Subscribe] unexpected type %T, code:%s", msg, codes.Internal)
			return
		}
		if err := server.Send(&gen.Event{Data: data}); err != nil {
			log.Printf("[ERROR][gRPC][Subscribe] send error: %v, code=%s", err, codes.Unavailable)
			if sub != nil {
				sub.Unsubscribe()
			}
			return
		}
		log.Printf("[DEBUG][gRPC][Subscribe] delivered message to client for key=%s: %s", req.Key, data)
	}

	var err error
	sub, err = s.bus.Subscribe(req.Key, handler)
	if err != nil {
		log.Printf("[ERROR][gRPC][Subscribe] failed to subscribe: %v, code:%s", err, codes.Internal)
		return status.Error(codes.Internal, "failed to subscribe")
	}

	defer func() {
		log.Printf("[INFO][gRPC][Subscribe] unsubscribing from key: %s", req.Key)
		sub.Unsubscribe()
	}()

	<-server.Context().Done()
	log.Printf("[INFO][gRPC][Subscribe] context done for key: %s", req.Key)
	return server.Context().Err()
}
