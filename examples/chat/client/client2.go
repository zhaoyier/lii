package main

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"github.com/leesper/holmes"
	"zhao.com/lii/client"
)

func main() {
	defer holmes.Start().Stop()

	//tao.Register(chat.ChatMessage, chat.DeserializeMessage, nil)

	c, err := net.Dial("tcp", "127.0.0.1:12000")
	if err != nil {
		holmes.Fatalln(err)
	}

	onConnect := client.OnConnectOption(func(c *client.ClientConn) bool {
		holmes.Infoln("on connect")
		return true
	})

	onError := client.OnErrorOption(func(c *client.ClientConn) {
		holmes.Infoln("on error")
	})

	onClose := client.OnCloseOption(func(c *client.ClientConn) {
		holmes.Infoln("on close")
	})

	onMessage := client.OnMessageOption(func(msg *client.Message, c *client.ClientConn) {
		fmt.Printf("[OnMessageOption] 返回的数据是:%+v|%s", msg, msg.Body)
	})

	options := []client.ServerOption{
		onConnect,
		onError,
		onClose,
		onMessage,
		client.ReconnectOption(),
	}

	conn := client.NewClientConn(0, c, options...)
	defer conn.Close()

	conn.Start()
	for {
		reader := bufio.NewReader(os.Stdin)
		talk, _ := reader.ReadString('\n')
		if talk == "bye\n" {
			break
		} else {
			fmt.Println("[main] 发送的字符串是：", talk, len(talk))
			msg := client.Message{ReqType: 0xabcd, Body: []byte(talk)}
			if err := conn.Write(msg); err != nil {
				holmes.Infoln("error", err)
			}
		}
	}
	fmt.Println("goodbye")
}
