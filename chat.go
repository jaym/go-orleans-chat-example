package main

import (
	"context"
	"fmt"
	time "time"

	"github.com/jaym/go-orleans-chat-example/gen"
	"github.com/jaym/go-orleans/grain"
	"github.com/jaym/go-orleans/grain/services"
)

type ChatRoomGrainActivator struct {
}

func (*ChatRoomGrainActivator) Activate(ctx context.Context, identity grain.Identity,
	svcs services.CoreGrainServices) (gen.ChatRoomGrain, error) {

	observerManager := services.NewInMemoryGrainObserverManager[*gen.ListenRequest, *gen.ChatMessage](identity, "Listen", 5*time.Minute)

	return &ChatRoomGrain{
		services:        svcs,
		observerManager: observerManager,
		identity:        identity,
		users:           make(map[string]struct{}),
	}, nil
}

type ChatRoomGrain struct {
	observerManager *services.InMemoryGrainObserverManager[*gen.ListenRequest, *gen.ChatMessage]
	services        services.CoreGrainServices
	identity        grain.Identity

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

	g.observerManager.Add(observer, &gen.ListenRequest{})

	return nil
}

func (g *ChatRoomGrain) UnsubscribeListenObserver(ctx context.Context, observer grain.Identity) error {
	g.observerManager.Remove(observer)
	return nil
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
	g.observerManager.NotifyAll(ctx, g.services.SiloClient(), req)

	return &gen.PublishResponse{}, nil
}
