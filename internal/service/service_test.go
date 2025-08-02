package service

import (
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
