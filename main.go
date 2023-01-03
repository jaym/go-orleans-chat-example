package main

import (
	"bufio"
	"context"
	"fmt"
	stdlog "log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	time "time"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/jaym/go-orleans-chat-example/gen"
	gengo "github.com/jaym/go-orleans-chat-example/gen-go"
	"github.com/jaym/go-orleans/grain"
	"github.com/jaym/go-orleans/plugins/discovery/static"
	psql_quicksetup "github.com/jaym/go-orleans/quicksetup/psql"
	"github.com/jaym/go-orleans/silo"
)

func main() {
	var membershipPort int
	var rpcPort int
	var servicePort int
	switch os.Args[1] {
	case "node1":
		membershipPort = 9991
		rpcPort = 8991
		servicePort = 9501
	case "node2":
		membershipPort = 9992
		rpcPort = 8992
		servicePort = 9502
	case "node3":
		membershipPort = 9993
		rpcPort = 8993
		servicePort = 9503
	default:
		panic("wrong")
	}

	stdr.SetVerbosity(10)
	log := stdr.NewWithOptions(stdlog.New(os.Stderr, "", stdlog.LstdFlags), stdr.Options{LogCaller: stdr.All})

	nodeName := os.Args[1]

	s, err := psql_quicksetup.Setup(
		context.Background(),
		psql_quicksetup.WithLogr(log),
		psql_quicksetup.WithNodeName(nodeName),
		psql_quicksetup.WithPGEnvironment(),
		psql_quicksetup.WithRPCAddr("127.0.0.1:"+strconv.Itoa(rpcPort)),
		psql_quicksetup.WithMembershipAddr("127.0.0.1:"+strconv.Itoa(membershipPort)),
		psql_quicksetup.WithSiloOptions(
			silo.WithDiscovery(static.New([]string{"127.0.0.1:9991", "127.0.0.1:9992"})),
			silo.WithMaxGrains(10),
		),
	)
	if err != nil {
		panic(err)
	}

	// gen.RegisterChatRoomGrainActivator(s, &ChatRoomGrainActivator{})
	gengo.RegisterChatRoomGrainActivator(s, &ChatRoomGrainActivatorImpl{})
	if err := s.Start(context.Background()); err != nil {
		panic(err)
	}

	stop := make(chan os.Signal, 1)

	// Register the signals we want to be notified, these 3 indicate exit
	// signals, similar to CTRL+C
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", servicePort))
	if err != nil {
		panic(err)
	}

	go listen(listener, servicePort, s, log)

	<-stop
	listener.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Stop(ctx); err != nil {
		panic(err)
	}
}

func listen(listener net.Listener, servicePort int, s *silo.Silo, log logr.Logger) {
	client := s.Client()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.V(0).Error(err, "failed to accept connection")
			return
		}

		ctx := context.Background()
		go func() {
			defer conn.Close()
			scanner := bufio.NewScanner(conn)
			observerGrain, err := s.CreateGrain()
			if err != nil {
				log.V(0).Error(err, "failed to create generic grain")
			}
			defer s.DestroyGrain(observerGrain)

			outChan := make(chan string, 10)
			defer close(outChan)

			go func() {
				for line := range outChan {
					if line != "" {
						conn.Write([]byte(line))
					}
				}
			}()

			msgObserver, err := gengo.NewGenericMessageObserverBuilder().
				OnReceive(func(ctx context.Context, sender grain.Identity, cm *gen.ChatMessage) {
					outChan <- fmt.Sprintf("[%s@%s]: %s\n", cm.From,
						sender.String(), cm.Msg)
				}).Build(observerGrain)
			if err != nil {
				panic(err)
			}

			outChan <- "Enter username:\n"
			var username string
			if !scanner.Scan() {
				outChan <- "[error] must enter username"
				return
			} else {
				username = scanner.Text()
			}
			if scanner.Err() != nil {
				log.Error(err, "an error occurred")
				return
			}

			outChan <- "Enter Commands:\n"

			for scanner.Scan() {
				data := scanner.Text()
				parts := strings.Split(data, " ")
				if len(parts) == 0 {
					continue
				}
				action := strings.TrimSpace(parts[0])
				args := parts[1:]

				line := "[ok] success\n"
				switch action {
				case "/join":
					if len(args) != 1 {
						line = "[error] incorrect args"
						break
					}
					g := gengo.ChatRoomGrainRef(client, args[0])
					_, err := g.Join(ctx, username, true)
					if err != nil {
						line = err.Error() + "\n"
						break
					}
					token, err := g.RegisterMessageObserver(ctx, msgObserver, &gen.ListenRequest{})
					if err != nil {
						line = err.Error() + "\n"
						break
					}
					line = token.String() + "\n"
				case "/ping":
					if len(args) != 1 {
						line = "[error] incorrect args\n"
						break
					}
					g := gengo.ChatRoomGrainRef(client, args[0])
					g.Ping(ctx)
				case "/msg":
					if len(args) < 2 {
						line = "[error] incorrect args\n"
						break
					}
					g := gengo.ChatRoomGrainRef(client, args[0])
					resp, ok, err := g.SendMessage(ctx, &gen.ChatMessage{
						Msg:  strings.Join(args[1:], " "),
						From: username,
					})
					if err != nil {
						line = fmt.Sprintf("[error] %s\n", err.Error())
					} else {
						line = fmt.Sprintf("[ok] count=%d ok=%v\n", resp.Count, ok)
					}
				case "/users":
					if len(args) != 1 {
						line = "[error] incorrect args\n"
						break
					}
					g := gengo.ChatRoomGrainRef(client, args[0])
					users, err := g.ListUsers(ctx)
					if err != nil {
						line = fmt.Sprintf("[error] %s\n", err.Error())
					} else {
						line = fmt.Sprintf("[ok] =>\n%s\n", strings.Join(users, "\n"))
					}
				default:
					line = "[error] unknown command\n"
				}
				outChan <- line
			}
		}()
	}
}

// func listen(listener net.Listener, servicePort int, s *silo.Silo, log logr.Logger) {
// 	client := s.Client()

// 	userNum := 0
// 	for {
// 		conn, err := listener.Accept()
// 		if err != nil {
// 			log.V(0).Error(err, "failed to accept connection")
// 			return
// 		}
// 		userName := fmt.Sprintf("user%04d%04d", servicePort, userNum)
// 		userNum++
// 		go func() {
// 			defer conn.Close()
// 			scanner := bufio.NewScanner(conn)

// 			subscriber, err := s.CreateGrain()
// 			if err != nil {
// 				log.V(0).Error(err, "failed to create subscriber grain")
// 				return
// 			}
// 			defer s.DestroyGrain(subscriber)

// 			stream, err := gen.CreateChatRoomGrainListenStream(subscriber)
// 			if err != nil {
// 				log.V(0).Error(err, "failed to create stream")
// 				return
// 			}

// 			ctx, cancel := context.WithCancel(context.Background())
// 			defer cancel()

// 			input := make(chan connInMsg, 10)
// 			go func() {
// 				streamC := stream.C()
// 				rooms := map[string]gen.ChatRoomGrainRef{}
// 				defer func() {
// 					for _, room := range rooms {
// 						_, err := room.Leave(context.Background(), &gen.LeaveRequest{
// 							UserName: userName,
// 						})
// 						if err != nil {
// 							log.V(0).Error(err, "leave failed")
// 						}
// 					}
// 				}()

// 				for {
// 					line := ""
// 				SEL:
// 					select {
// 					case <-ctx.Done():
// 						return
// 					case chatRoomMessage := <-streamC:
// 						line = fmt.Sprintf("[%s@%s]: %s\n", chatRoomMessage.Value.From,
// 							chatRoomMessage.Sender.ID, chatRoomMessage.Value.Msg)
// 					case inMsg := <-input:
// 						if inMsg.action == "/join" {
// 							if len(inMsg.args) != 1 {
// 								continue
// 							}
// 							ref := gen.GetChatRoomGrain(client, grain.Identity{
// 								GrainType: "ChatRoomGrain",
// 								ID:        inMsg.args[0],
// 							})

// 							_, err = ref.Join(ctx, &gen.JoinRequest{
// 								UserName: userName,
// 								Listen:   true,
// 							})
// 							if err != nil {
// 								log.V(0).Error(err, "join failed")
// 								line = "error\n"
// 								break SEL
// 							}

// 							err = stream.Observe(
// 								ctx,
// 								grain.Identity{
// 									GrainType: "ChatRoomGrain",
// 									ID:        inMsg.args[0],
// 								},
// 								&gen.ListenRequest{},
// 							)
// 							if err != nil {
// 								log.V(0).Error(err, "stream observe failed")
// 								line = "error\n"
// 								break SEL
// 							}

// 							rooms[inMsg.args[0]] = ref
// 						} else if inMsg.action == "/msg" {
// 							if len(inMsg.args) < 2 {
// 								continue
// 							}

// 							if room, ok := rooms[inMsg.args[0]]; ok {
// 								_, err := room.Publish(ctx, &gen.ChatMessage{
// 									From: userName,
// 									Msg:  strings.Join(inMsg.args[1:], " "),
// 								})
// 								if err != nil {
// 									log.V(0).Error(err, "stream observe failed")
// 									line = "error\n"
// 								} else {
// 									line = "messaged\n"
// 								}
// 							} else {
// 								line = "not a member of the room\n"
// 							}
// 						} else if inMsg.action == "/leave" {
// 							line = "left\n"
// 						}
// 					}
// 					if line != "" {
// 						conn.Write([]byte(line))
// 					}
// 				}

// 			}()

// 			for scanner.Scan() {
// 				data := scanner.Text()
// 				parts := strings.Split(data, " ")
// 				if len(parts) == 0 {
// 					continue
// 				}
// 				action := strings.TrimSpace(parts[0])
// 				args := parts[1:]
// 				input <- connInMsg{
// 					action: action,
// 					args:   args,
// 				}
// 			}
// 		}()
// 	}
// }

type connInMsg struct {
	action string
	args   []string
}

type connOutMsg struct {
	line string
}
