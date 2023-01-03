package gengo

import (
	"context"
	"errors"

	__grain "github.com/jaym/go-orleans/grain"
	__descriptor "github.com/jaym/go-orleans/grain/descriptor"
	__generic "github.com/jaym/go-orleans/grain/generic"
	__services "github.com/jaym/go-orleans/grain/services"

	gen "github.com/jaym/go-orleans-chat-example/gen"
)

type ChatRoomGrainActivation interface {
	__grain.Activation
	_ChatRoomGrainActivation()
}

func NewChatRoomGrainActivation(siloClient __grain.SiloClient, impl ChatRoomGrain) ChatRoomGrainActivation {
	return &_ChatRoomGrainActivation{
		siloClient: siloClient,
		impl:       impl,
	}
}

type ChatRoomGrainActivator interface {
	Activate(ctx context.Context, identity __grain.Identity, services __services.CoreGrainServices) (ChatRoomGrain, error)
}

func RegisterChatRoomGrainActivator(registrar __descriptor.Registrar, activator ChatRoomGrainActivator) {
	registrar.RegisterV2(
		"ChatRoomGrain",
		func(ctx context.Context, identity __grain.Identity,
			services __services.CoreGrainServices) (__grain.Activation, error) {
			a, err := activator.Activate(ctx, identity, services)
			if err != nil {
				return nil, err
			}
			return NewChatRoomGrainActivation(services.SiloClient(), a), nil
		},
	)
}

type _ChatRoomGrainActivation struct {
	siloClient __grain.SiloClient
	impl       ChatRoomGrain
}

func (__s *_ChatRoomGrainActivation) _ChatRoomGrainActivation() {}

func (__s *_ChatRoomGrainActivation) InvokeMethod(ctx context.Context, method string, sender __grain.Identity,
	dec __grain.Deserializer, respSerializer __grain.Serializer) error {
	siloClient := __s.siloClient
	_ = siloClient

	switch method {

	case "Join":

		argusernameUncasted, err := dec.String()
		if err != nil {
			return err
		}
		argusername := string(argusernameUncasted)

		arglistenUncasted, err := dec.Bool()
		if err != nil {
			return err
		}
		arglisten := bool(arglistenUncasted)

		out0, err := __s.impl.Join(ctx,

			argusername,

			arglisten,
		)

		if err != nil {
			return err
		}

		if err := respSerializer.Interface(out0); err != nil {
			return err
		}

		return nil

	case "RegisterMessageObserver":

		argobserverStr, err := dec.String()
		if err != nil {
			return err
		}
		argobserverIdentity := __grain.Identity{}
		err = argobserverIdentity.UnmarshalText([]byte(argobserverStr))
		if err != nil {
			return err
		}

		argobserver := MessageObserverRefFromIdentity(siloClient, argobserverIdentity)

		argreq := new(gen.ListenRequest)

		if err := dec.Interface(argreq); err != nil {
			return err
		}

		out0, err := __s.impl.RegisterMessageObserver(ctx,
			argobserver,

			argreq,
		)

		if err != nil {
			return err
		}

		if text, err := out0.MarshalText(); err != nil {
			return err
		} else {
			respSerializer.String(string(text))
		}

		return nil

	case "Ping":

		__s.impl.Ping(ctx)

		return nil

	case "SendMessage":

		argmsg := new(gen.ChatMessage)

		if err := dec.Interface(argmsg); err != nil {
			return err
		}

		out0, out1, err := __s.impl.SendMessage(ctx,

			argmsg,
		)

		if err != nil {
			return err
		}

		if err := respSerializer.Interface(out0); err != nil {
			return err
		}

		respSerializer.Bool(bool(out1))

		return nil

	case "ListUsers":

		out0, err := __s.impl.ListUsers(ctx)

		if err != nil {
			return err
		}

		respSerializer.StringList(out0)

		return nil

	case "UnregisterObserver":

		argp1 := new(__grain.ObserverRegistrationToken)
		argp1Str, err := dec.String()
		if err != nil {
			return err
		}
		err = argp1.UnmarshalText([]byte(argp1Str))
		if err != nil {
			return err
		}

		__s.impl.UnregisterObserver(ctx,

			*argp1,
		)

		return nil

	case "RefreshObserver":

		argp1 := new(__grain.ObserverRegistrationToken)
		argp1Str, err := dec.String()
		if err != nil {
			return err
		}
		err = argp1.UnmarshalText([]byte(argp1Str))
		if err != nil {
			return err
		}

		out0, err := __s.impl.RefreshObserver(ctx,

			*argp1,
		)

		if err != nil {
			return err
		}

		if text, err := out0.MarshalText(); err != nil {
			return err
		} else {
			respSerializer.String(string(text))
		}

		return nil

	}
	return errors.New("unknown method")
}

type _ChatRoomGrainClient struct {
	__grain.Identity
	siloClient __grain.SiloClient
}

func ChatRoomGrainRef(siloClient __grain.SiloClient, id string) ChatRoomGrain {
	return &_ChatRoomGrainClient{
		siloClient: siloClient,
		Identity: __grain.Identity{
			GrainType: "ChatRoomGrain",
			ID:        id,
		},
	}
}

func ChatRoomGrainRefFromIdentity(siloClient __grain.SiloClient, id __grain.Identity) ChatRoomGrain {
	return &_ChatRoomGrainClient{
		siloClient: siloClient,
		Identity:   id,
	}
}

func (__c *_ChatRoomGrainClient) Join(ctx context.Context, username string, listen bool) (*gen.JoinResponse, error) {
	f := __c.siloClient.InvokeMethodV2(ctx, __c.Identity, "Join", func(respSerializer __grain.Serializer) error {

		respSerializer.String(string(username))

		respSerializer.Bool(bool(listen))

		return nil

	})

	var out0 *gen.JoinResponse

	resp, err := f.Await(ctx)
	if err != nil {
		return nil, err
	}

	err = resp.Get(func(dec __grain.Deserializer) error {
		var err error
		_ = err

		out0 = new(gen.JoinResponse)

		err = dec.Interface(out0)
		if err != nil {
			return err
		}

		return err
	})
	if err != nil {
		return nil, err
	}
	return out0, err

}

func (__c *_ChatRoomGrainClient) RegisterMessageObserver(ctx context.Context, observer MessageObserver, req *gen.ListenRequest) (__grain.ObserverRegistrationToken, error) {
	f := __c.siloClient.InvokeMethodV2(ctx, __c.Identity, "RegisterMessageObserver", func(respSerializer __grain.Serializer) error {

		if text, err := observer.GetIdentity().MarshalText(); err != nil {
			return err
		} else {
			respSerializer.String(string(text))
		}

		if err := respSerializer.Interface(req); err != nil {
			return err
		}

		return nil

	})

	var out0 __grain.ObserverRegistrationToken

	resp, err := f.Await(ctx)
	if err != nil {
		return out0, err
	}

	err = resp.Get(func(dec __grain.Deserializer) error {
		var err error
		_ = err

		out0Str, err := dec.String()
		if err != nil {
			return err
		}
		err = out0.UnmarshalText([]byte(out0Str))
		if err != nil {
			return err
		}

		return err
	})
	if err != nil {
		return out0, err
	}
	return out0, err

}

func (__c *_ChatRoomGrainClient) Ping(ctx context.Context) {

	__c.siloClient.InvokeOneWayMethod(ctx, []__grain.Identity{__c.Identity}, "Ping", func(respSerializer __grain.Serializer) error {

		return nil

	})

}

func (__c *_ChatRoomGrainClient) SendMessage(ctx context.Context, msg *gen.ChatMessage) (*gen.SendMessageResponse, bool, error) {
	f := __c.siloClient.InvokeMethodV2(ctx, __c.Identity, "SendMessage", func(respSerializer __grain.Serializer) error {

		if err := respSerializer.Interface(msg); err != nil {
			return err
		}

		return nil

	})

	var out0 *gen.SendMessageResponse

	var out1 bool

	resp, err := f.Await(ctx)
	if err != nil {
		return nil, out1, err
	}

	err = resp.Get(func(dec __grain.Deserializer) error {
		var err error
		_ = err

		out0 = new(gen.SendMessageResponse)

		err = dec.Interface(out0)
		if err != nil {
			return err
		}

		out1Uncasted, err := dec.Bool()
		if err != nil {
			return err
		}
		out1 = bool(out1Uncasted)

		return err
	})
	if err != nil {
		return nil, out1, err
	}
	return out0, out1, err

}

func (__c *_ChatRoomGrainClient) ListUsers(ctx context.Context) ([]string, error) {
	f := __c.siloClient.InvokeMethodV2(ctx, __c.Identity, "ListUsers", func(respSerializer __grain.Serializer) error {

		return nil

	})

	var out0 []string

	resp, err := f.Await(ctx)
	if err != nil {
		return out0, err
	}

	err = resp.Get(func(dec __grain.Deserializer) error {
		var err error
		_ = err

		out0, err = dec.StringList()
		if err != nil {
			return err
		}

		return err
	})
	if err != nil {
		return out0, err
	}
	return out0, err

}

func (__c *_ChatRoomGrainClient) UnregisterObserver(ctx context.Context, p1 __grain.ObserverRegistrationToken) {

	__c.siloClient.InvokeOneWayMethod(ctx, []__grain.Identity{__c.Identity}, "UnregisterObserver", func(respSerializer __grain.Serializer) error {

		if text, err := p1.MarshalText(); err != nil {
			return err
		} else {
			respSerializer.String(string(text))
		}

		return nil

	})

}

func (__c *_ChatRoomGrainClient) RefreshObserver(ctx context.Context, p1 __grain.ObserverRegistrationToken) (__grain.ObserverRegistrationToken, error) {
	f := __c.siloClient.InvokeMethodV2(ctx, __c.Identity, "RefreshObserver", func(respSerializer __grain.Serializer) error {

		if text, err := p1.MarshalText(); err != nil {
			return err
		} else {
			respSerializer.String(string(text))
		}

		return nil

	})

	var out0 __grain.ObserverRegistrationToken

	resp, err := f.Await(ctx)
	if err != nil {
		return out0, err
	}

	err = resp.Get(func(dec __grain.Deserializer) error {
		var err error
		_ = err

		out0Str, err := dec.String()
		if err != nil {
			return err
		}
		err = out0.UnmarshalText([]byte(out0Str))
		if err != nil {
			return err
		}

		return err
	})
	if err != nil {
		return out0, err
	}
	return out0, err

}

type _MessageObserverClient struct {
	__grain.Identity
	siloClient __grain.SiloClient
}

func MessageObserverRefFromIdentity(siloClient __grain.SiloClient, id __grain.Identity) MessageObserver {
	return &_MessageObserverClient{
		siloClient: siloClient,
		Identity:   id,
	}
}

func (__c *_MessageObserverClient) Receive(ctx context.Context, p1 *gen.ChatMessage) {

	__c.siloClient.InvokeOneWayMethod(ctx, []__grain.Identity{__c.Identity}, "Receive", func(respSerializer __grain.Serializer) error {

		if err := respSerializer.Interface(p1); err != nil {
			return err
		}

		return nil

	})

}

func MessageObserverReceiveInvokeBatch(ctx context.Context, siloClient __grain.SiloClient, observerList []MessageObserver, p1 *gen.ChatMessage) {
	idents := make([]__grain.Identity, len(observerList))
	for i, o := range observerList {
		idents[i] = o.GetIdentity()
	}
	siloClient.InvokeOneWayMethod(ctx, idents, "Receive", func(respSerializer __grain.Serializer) error {

		if err := respSerializer.Interface(p1); err != nil {
			return err
		}

		return nil

	})
}

type MessageObserverOnReceiveFunc func(ctx context.Context, sender __grain.Identity, p1 *gen.ChatMessage)

type GenericMessageObserverBuilder struct {
	onReceive MessageObserverOnReceiveFunc
}

func NewGenericMessageObserverBuilder() *GenericMessageObserverBuilder {
	return &GenericMessageObserverBuilder{

		onReceive: func(ctx context.Context, sender __grain.Identity, p1 *gen.ChatMessage) {},
	}
}

func (__b *GenericMessageObserverBuilder) OnReceive(f MessageObserverOnReceiveFunc) *GenericMessageObserverBuilder {
	__b.onReceive = f
	return __b
}

func (__b *GenericMessageObserverBuilder) Build(g *__generic.Grain) (MessageObserver, error) {

	g.OnMethod(
		"Receive",
		func(ctx context.Context, siloClient __grain.SiloClient, method string, sender __grain.Identity, dec __grain.Deserializer, respSerializer __grain.Serializer) error {
			var err error
			_ = err

			argp1 := new(gen.ChatMessage)

			if err := dec.Interface(argp1); err != nil {
				return err
			}

			__b.onReceive(ctx, sender,

				argp1,
			)
			return nil
		})

	return genericMessageObserver{g}, nil
}

type genericMessageObserver struct {
	*__generic.Grain
}

func (genericMessageObserver) Receive(ctx context.Context, p1 *gen.ChatMessage) {
	panic("unreachable")
}
