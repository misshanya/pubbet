package grpc

import (
	"context"
	"errors"
	pb "github.com/misshanya/pubbet/gen/go/pubbet/v1"
	"github.com/misshanya/pubbet/internal/errorz"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log/slog"
)

type service interface {
	SendMessage(topicName string, message []byte)
	ListenMessages(ctx context.Context, topicName string) (<-chan []byte, error)
}

type Handler struct {
	l       *slog.Logger
	service service
	pb.UnimplementedPubbetServer
}

func NewHandler(l *slog.Logger, grpcServer *grpc.Server, service service) {
	shortenerGrpc := &Handler{l: l, service: service}
	pb.RegisterPubbetServer(grpcServer, shortenerGrpc)
}

func (h *Handler) PublishMessages(stream pb.Pubbet_PublishMessagesServer) error {
	req, err := stream.Recv()
	if err != nil {
		h.l.Error("failed to read first message", "error", err)
		return err
	}

	var messagesSent int64

	h.l.Info("Starting message publishing")

	for {
		req, err = stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.PublishMessagesResponse{
				MessagesSent: messagesSent,
			})
		}
		if err != nil {
			h.l.Error("failed to read stream", "error", err)
			return err
		}

		topicName := req.GetTopicName()
		if topicName == "" {
			h.l.Error("topic name is empty")
			return status.Error(codes.InvalidArgument, "topic name is empty")
		}

		message := req.GetMessage()
		if len(message) <= 0 {
			h.l.Error("len(message) <= 0")
			return status.Error(codes.InvalidArgument, "len(message) <= 0")
		}

		h.l.Info("Sending a message")
		h.service.SendMessage(topicName, message)
		messagesSent++
	}
}

func (h *Handler) ListenMessages(req *pb.ListenTopicRequest, stream pb.Pubbet_ListenMessagesServer) error {
	topicName := req.TopicName

	messagesChan, err := h.service.ListenMessages(stream.Context(), topicName)
	if err != nil {
		switch {
		case errors.Is(err, errorz.ErrTopicNotExists):
			return status.Error(codes.NotFound, err.Error())
		default:
			return status.Error(codes.Internal, err.Error())
		}
	}

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case msg := <-messagesChan:
			h.l.Info("New message in a messagesChan", "message", msg)
			resp := &pb.TopicMessages{Message: msg}
			if err := stream.Send(resp); err != nil {
				h.l.Error("failed to send message to stream")
				return err
			}
		}
	}
}
