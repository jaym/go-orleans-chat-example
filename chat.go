package main

import (
	"context"
	"fmt"
	time "time"

	"github.com/jaym/go-orleans-chat-example/gen"
	"github.com/jaym/go-orleans/grain"
)

type ChatRoomGrainActivator struct {
}

func (*ChatRoomGrainActivator) Activate(ctx context.Context, identity grain.Identity,
	services gen.ChatRoomGrainServices) (gen.ChatRoomGrain, error) {
	return &ChatRoomGrain{
		services: services,
		identity: identity,
		users:    make(map[string]struct{}),
	}, nil
}

type ChatRoomGrain struct {
	services gen.ChatRoomGrainServices
	identity grain.Identity

	users map[string]struct{}
}

func (g *ChatRoomGrain) GetIdentity() grain.Identity {
	return g.identity
}

func (g *ChatRoomGrain) Join(ctx context.Context, req *gen.JoinRequest) (*gen.JoinResponse, error) {
	g.users[req.UserName] = struct{}{}

	_, _ = g.Publish(ctx, &gen.ChatMessage{
		From: "ROOM",
		Msg:  fmt.Sprintf("%s joined the room", req.UserName),
	})

	return &gen.JoinResponse{}, nil
}

func (g *ChatRoomGrain) RegisterListenObserver(ctx context.Context, observer grain.Identity,
	registrationTimeout time.Duration, req *gen.ListenRequest) error {

	err := g.services.AddListenObserver(ctx, observer, registrationTimeout, &gen.ListenRequest{})
	if err != nil {
		return err
	}
	return nil
}

func (g *ChatRoomGrain) UnsubscribeListenObserver(ctx context.Context, observer grain.Identity) error {
	return g.services.RemoveListenObserver(ctx, observer)
}

func (g *ChatRoomGrain) Leave(ctx context.Context, req *gen.LeaveRequest) (*gen.LeaveResponse, error) {
	_, err := g.Publish(ctx, &gen.ChatMessage{
		From: "ROOM",
		Msg:  fmt.Sprintf("%s has left the room", req.UserName),
	})
	if err != nil {
		return nil, err
	}

	delete(g.users, req.UserName)

	return &gen.LeaveResponse{}, nil
}

func (g *ChatRoomGrain) Publish(ctx context.Context, req *gen.ChatMessage) (*gen.PublishResponse, error) {
	observers, err := g.services.ListListenObservers(ctx)
	if err != nil {
		return nil, err
	}
	err = g.services.NotifyListenObservers(ctx, observers, req)
	if err != nil {
		return nil, err
	}
	return &gen.PublishResponse{}, nil
}
