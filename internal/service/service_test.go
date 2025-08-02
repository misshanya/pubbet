package service

import (
	"context"
	"github.com/misshanya/pubbet/internal/errorz"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"sync"
	"testing"
)

func TestService_SendMessage(t *testing.T) {
	tests := []struct {
		Name           string
		InputTopicName string
		InputMessage   []byte
		ExceptedTopics map[string]*Topic
	}{
		{
			Name:           "Successfully sent",
			InputTopicName: "some-topic",
			InputMessage:   []byte("some data"),
			ExceptedTopics: map[string]*Topic{
				"some-topic": {
					mu:       &sync.Mutex{},
					messages: [][]byte{[]byte("some data")},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			svc := New(
				slog.New(
					slog.NewTextHandler(os.Stdout, nil),
				),
			)

			svc.SendMessage(tt.InputTopicName, tt.InputMessage)
			assert.Equal(t, tt.ExceptedTopics, svc.topics)
		})
	}
}

func TestService_ListenMessages(t *testing.T) {
	svc := New(
		slog.New(
			slog.NewTextHandler(os.Stdout, nil),
		),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := svc.ListenMessages(ctx, "non-exist-topic")
	assert.Equal(t, errorz.ErrTopicNotExists, err)

	messages := [][]byte{
		[]byte("hello world"),
		[]byte("smth cool"),
		[]byte("ye"),
	}

	svc.topics["exist-topic"] = &Topic{
		mu:       &sync.Mutex{},
		messages: messages,
	}

	ch, err := svc.ListenMessages(ctx, "exist-topic")
	assert.NoError(t, err)

	pointer := len(messages) - 1
	for v := range ch {
		assert.Equal(t, messages[pointer], v)

		if pointer <= 0 {
			cancel()
		}

		pointer--
	}
}
