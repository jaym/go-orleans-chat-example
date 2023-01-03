package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jaym/go-orleans-chat-example/gen"
	gengo "github.com/jaym/go-orleans-chat-example/gen-go"
	"github.com/jaym/go-orleans/grain"
	"github.com/jaym/go-orleans/grain/services"
)

type ChatRoomGrainImpl struct {
	grain.Identity
	client    grain.SiloClient
	observers []gengo.MessageObserver
	msgCount  int
	users     map[string]struct{}
}

func (c *ChatRoomGrainImpl) Join(ctx context.Context, username string, listen bool) (*gen.JoinResponse, error) {
	c.users[username] = struct{}{}
	fmt.Printf("Join %s %v\n", username, listen)
	return &gen.JoinResponse{}, nil
}

func (s *ChatRoomGrainImpl) RegisterMessageObserver(ctx context.Context, observer gengo.MessageObserver, req *gen.ListenRequest) (grain.ObserverRegistrationToken, error) {
	s.observers = append(s.observers, observer)
	return grain.ObserverRegistrationToken{
		Observer:       observer.GetIdentity(),
		ObservableName: "message-observer",
		ID:             "123",
		Expires:        time.Now().Add(30 * time.Second),
	}, nil
}

func (s *ChatRoomGrainImpl) Ping(ctx context.Context) {
	gengo.MessageObserverReceiveInvokeBatch(ctx, s.client, s.observers, &gen.ChatMessage{
		From: s.GetIdentity().ID,
		Msg:  "PING",
	})
}

func (*ChatRoomGrainImpl) UnregisterObserver(context.Context, grain.ObserverRegistrationToken) {

}

func (*ChatRoomGrainImpl) RefreshObserver(ctx context.Context, t grain.ObserverRegistrationToken) (grain.ObserverRegistrationToken, error) {
	return grain.ObserverRegistrationToken{
		Observer:       t.Observer,
		ObservableName: "message-observer",
		ID:             "123",
		Expires:        time.Now().Add(30 * time.Second),
	}, nil
}

func (c *ChatRoomGrainImpl) SendMessage(ctx context.Context, msg *gen.ChatMessage) (*gen.SendMessageResponse, bool, error) {
	c.msgCount++
	gengo.MessageObserverReceiveInvokeBatch(ctx, c.client, c.observers, msg)
	return &gen.SendMessageResponse{
		Count: int64(c.msgCount),
	}, true, nil
}

func (c *ChatRoomGrainImpl) ListUsers(ctx context.Context) ([]string, error) {
	users := make([]string, len(c.users))
	i := 0
	for u := range c.users {
		users[i] = u
		i++
	}
	return users, nil
}

type ChatRoomGrainActivatorImpl struct {
}

func (*ChatRoomGrainActivatorImpl) Activate(ctx context.Context, identity grain.Identity, services services.CoreGrainServices) (gengo.ChatRoomGrain, error) {
	g := &ChatRoomGrainImpl{
		Identity: identity,
		client:   services.SiloClient(),
		users:    map[string]struct{}{},
	}
	return g, nil
}
