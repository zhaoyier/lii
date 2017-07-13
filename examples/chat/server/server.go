package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/leesper/holmes"
	//"github.com/leesper/tao"
	//"open.com/tao"
	//"github.com/leesper/tao/examples/chat"
	//"open.com/tao/examples/chat"
	"zhao.com/lii/server"
)

// ChatServer is the chatting server.
type ChatServer struct {
	*server.Server
}

// NewChatServer returns a ChatServer.
func NewChatServer() *ChatServer {
	onConnectOption := server.OnConnectOption(func(conn server.WriteCloser) bool {
		holmes.Infoln("on connect")
		return true
	})
	onErrorOption := server.OnErrorOption(func(conn server.WriteCloser) {
		holmes.Infoln("on error")
	})
	onCloseOption := server.OnCloseOption(func(conn server.WriteCloser) {
		holmes.Infoln("close chat client")
	})
	return &ChatServer{
		server.NewServer(onConnectOption, onErrorOption, onCloseOption),
	}
}

func main() {
	defer holmes.Start().Stop()

	//tao.Register(chat.ChatMessage, chat.DeserializeMessage, chat.ProcessMessage)

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", "0.0.0.0", 12000))
	if err != nil {
		holmes.Fatalln("listen error", err)
	}
	chatServer := NewChatServer()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		chatServer.Stop()
	}()

	holmes.Infoln(chatServer.Start(l))
}
