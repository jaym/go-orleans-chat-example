package gengo

//go:generate ../bin/goor-gen -out chat.goor.go .

import (
	"context"

	"github.com/jaym/go-orleans-chat-example/gen"
	"github.com/jaym/go-orleans/goor-gen/goor"
	"github.com/jaym/go-orleans/grain"
)

type MessageObserver interface {
	goor.Observer
	Receive(context.Context, *gen.ChatMessage)
}

type ChatRoomGrain interface {
	goor.ObservableGrain
	// Join is a normal request response RPC
	Join(ctx context.Context, username string, listen bool) (*gen.JoinResponse, error)
	// RegisterMessageObserver takes an observer
	RegisterMessageObserver(ctx context.Context, observer MessageObserver, req *gen.ListenRequest) (grain.ObserverRegistrationToken, error)
	Ping(ctx context.Context)
	SendMessage(ctx context.Context, msg *gen.ChatMessage) (*gen.SendMessageResponse, bool, error)
	ListUsers(ctx context.Context) ([]string, error)
}
