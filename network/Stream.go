package network

import (
	"context"
)

type Stream interface {
	Close() error

	NextMessage() (*Message, error)
	SendMessage(ctx context.Context, message *Message) error
	NodeId() string
}
