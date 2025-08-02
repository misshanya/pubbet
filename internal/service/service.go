package service

import (
	"context"
	"github.com/misshanya/pubbet/internal/errorz"
	"log/slog"
	"sync"
)

type Topic struct {
	mu       *sync.Mutex
	messages [][]byte
}

type Service struct {
	l      *slog.Logger
	topics map[string]*Topic
}

func New(l *slog.Logger) *Service {
	topics := make(map[string]*Topic)
	return &Service{l: l, topics: topics}
}

func (s *Service) SendMessage(topicName string, message []byte) {
	if s.topics[topicName] == nil {
		s.l.Info("Initializing a new topic")
		s.topics[topicName] = &Topic{
			mu:       &sync.Mutex{},
			messages: [][]byte{},
		}
	}

	topic := s.topics[topicName]

	topic.mu.Lock()
	s.l.Info("Appending slice to a message", "message", message)
	topic.messages = append([][]byte{message}, topic.messages...) // put message at the beginning of the slice
	topic.mu.Unlock()
}

func (s *Service) ListenMessages(ctx context.Context, topicName string) (<-chan []byte, error) {
	ch := make(chan []byte, 1)

	if s.topics[topicName] == nil {
		return nil, errorz.ErrTopicNotExists
	}

	s.l.Info("Going to stream", "topic", s.topics[topicName])

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(ch)
				return
			default:
				topic := s.topics[topicName]
				if len(topic.messages) > 0 {
					s.l.Info("len(topic.messages) > 0")
					topic.mu.Lock()
					s.l.Info("Sending message to a channel", "message", topic.messages[len(topic.messages)-1])
					ch <- topic.messages[len(topic.messages)-1]
					topic.messages = topic.messages[:len(topic.messages)-1]
					s.l.Info("New slice", "slice", topic.messages)
					topic.mu.Unlock()
				}
			}
		}
	}()

	return ch, nil
}
