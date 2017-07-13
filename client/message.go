package client

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/leesper/holmes"
)

const (
	// HeartBeat is the default heart beat message number.
	HeartBeat = 0
)

// Handler takes the responsibility to handle incoming messages.
type Handler interface {
	Handle(context.Context, interface{})
}

// HandlerFunc serves as an adapter to allow the use of ordinary functions as handlers.
type HandlerFunc func(context.Context, *ClientConn)

// Handle calls f(ctx, c)
func (f HandlerFunc) Handle(ctx context.Context, c *ClientConn) {
	f(ctx, c)
}

type HandlerFunc2 func(msg *Message, s *ClientConn)

func ProcessMessage(msg *Message, s *ClientConn) {
	data, _ := msg.Serialize()
	s.rawConn.Write(data)
}

// UnmarshalFunc unmarshals bytes into Message.
type UnmarshalFunc func([]byte) (Message, error)

// handlerUnmarshaler is a combination of unmarshal and handle functions for message.
type handlerUnmarshaler struct {
	handler     HandlerFunc
	unmarshaler UnmarshalFunc
}

var (
	buf *bytes.Buffer
	// messageRegistry is the registry of all
	// message-related unmarshal and handle functions.
	messageRegistry map[int32]handlerUnmarshaler
)

func init() {
	messageRegistry = map[int32]handlerUnmarshaler{}
	buf = new(bytes.Buffer)
}

// Register registers the unmarshal and handle functions for msgType.
// If no unmarshal function provided, the message will not be parsed.
// If no handler function provided, the message will not be handled unless you
// set a default one by calling SetOnMessageCallback.
// If Register being called twice on one msgType, it will panics.
func Register(msgType int32, unmarshaler func([]byte) (Message, error), handler func(context.Context, *ClientConn)) {
	if _, ok := messageRegistry[msgType]; ok {
		panic(fmt.Sprintf("trying to register message %d twice", msgType))
	}

	messageRegistry[msgType] = handlerUnmarshaler{
		unmarshaler: unmarshaler,
		handler:     HandlerFunc(handler),
	}
}

// GetUnmarshalFunc returns the corresponding unmarshal function for msgType.
func GetUnmarshalFunc(msgType int32) UnmarshalFunc {
	entry, ok := messageRegistry[msgType]
	if !ok {
		return nil
	}
	return entry.unmarshaler
}

// GetHandlerFunc returns the corresponding handler function for msgType.
func GetHandlerFunc(msgType int32) HandlerFunc {
	entry, ok := messageRegistry[msgType]
	if !ok {
		return nil
	}
	return entry.handler
}

// Message represents the structured data that can be handled.
type Message struct {
	ReqType int32
	BodyLen int32
	Body    []byte
}

func (m *Message) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Reset()
	binary.Write(buf, binary.LittleEndian, m.ReqType)
	binary.Write(buf, binary.LittleEndian, int32(len(m.Body)))
	buf.WriteString(string(m.Body))
	fmt.Println("[Serialize] 序列化数据：", len(m.Body), "|", buf.String(), buf.Len())
	return buf.Bytes(), nil
}

// HeartBeatMessage for application-level keeping alive.
type HeartBeatMessage struct {
	Timestamp int64
}

// Serialize serializes HeartBeatMessage into bytes.
func (hbm HeartBeatMessage) Serialize() ([]byte, error) {
	buf.Reset()
	err := binary.Write(buf, binary.LittleEndian, hbm.Timestamp)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MessageNumber returns message number.
func (hbm HeartBeatMessage) MessageNumber() int32 {
	return HeartBeat
}

// DeserializeHeartBeat deserializes bytes into Message.
// func DeserializeHeartBeat(data []byte) (message Message, err error) {
// 	var timestamp int64
// 	if data == nil {
// 		return nil, errors.New("nil data")
// 	}
// 	buf := bytes.NewReader(data)
// 	err = binary.Read(buf, binary.LittleEndian, &timestamp)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return HeartBeatMessage{
// 		Timestamp: timestamp,
// 	}, nil
// }

// // HandleHeartBeat updates connection heart beat timestamp.
// func HandleHeartBeat(ctx context.Context, c *ClientConn) {
// 	msg := MessageFromContext(ctx)
// 	c.SetHeartBeat(msg.(HeartBeatMessage).Timestamp)
// }

// Codec is the interface for message coder and decoder.
// Application programmer can define a custom codec themselves.
type Codec interface {
	Decode(net.Conn) (*Message, error)
	Encode(Message) ([]byte, error)
}

// TypeLengthValueCodec defines a special codec.
// Format: type-length-value |4 bytes|4 bytes|n bytes <= 8M|
type TypeLengthValueCodec struct{}

// Decode decodes the bytes data into Message
func (codec TypeLengthValueCodec) Decode(raw net.Conn) (*Message, error) {
	fmt.Println("[Decode] 准备解析数据：")
	byteChan := make(chan []byte)
	errorChan := make(chan error)

	go func(bc chan []byte, ec chan error) {
		typeData := make([]byte, 4)
		_, err := io.ReadFull(raw, typeData)
		if err != nil {
			ec <- err
			close(bc)
			close(ec)
			fmt.Println("[Decode]读取网络消息头部异常:", err)
			holmes.Debugln("go-routine read message type exited")
			return
		}
		bc <- typeData
	}(byteChan, errorChan)

	var typeBytes []byte

	select {
	case err := <-errorChan:
		return nil, err

	case typeBytes = <-byteChan:
		if typeBytes == nil {
			holmes.Warnln("read type bytes nil")
			return nil, errors.New("more than 8M data")
		}
		typeBuf := bytes.NewReader(typeBytes)
		var msgType, msgLen int32
		if err := binary.Read(typeBuf, binary.LittleEndian, &msgType); err != nil {
			return nil, err
		}
		fmt.Println("[Decode] 请求网络消息类型是：", msgType)
		lengthBytes := make([]byte, 4)
		_, err := io.ReadFull(raw, lengthBytes)
		if err != nil {
			return nil, err
		}
		lengthBuf := bytes.NewReader(lengthBytes)
		if err = binary.Read(lengthBuf, binary.LittleEndian, &msgLen); err != nil {
			return nil, err
		}
		if msgLen > (1 << 23) {
			holmes.Errorf("message(type %d) has bytes(%d) beyond max %d\n", msgType, msgLen, (1 << 23))
			return nil, errors.New("more than 8M data")
		}

		// read application data
		msgBytes := make([]byte, msgLen)
		_, err = io.ReadFull(raw, msgBytes)
		if err != nil {
			return nil, err
		}

		msg := Message{
			ReqType: msgType,
			BodyLen: msgLen,
			Body:    msgBytes,
		}

		fmt.Printf("[Decode] 返回数据包:%+v|%s\n", msg, string(msg.Body))

		return &msg, nil

		// deserialize message from bytes
		//unmarshaler := GetUnmarshalFunc(msgType)
		//if unmarshaler == nil {
		//	return nil, ErrUndefined(msgType)
		//}
		//return unmarshaler(msgBytes)
	}
}

// Encode encodes the message into bytes data.
func (codec TypeLengthValueCodec) Encode(msg Message) ([]byte, error) {
	data, err := msg.Serialize()
	if err != nil {
		return nil, err
	}
	return data, err
}

// ContextKey is the key type for putting context-related data.
type contextKey string

// Context keys for messge, server and net ID.
const (
	messageCtx contextKey = "message"
	serverCtx  contextKey = "server"
	netIDCtx   contextKey = "netid"
)

// NewContextWithMessage returns a new Context that carries message.
func NewContextWithMessage(ctx context.Context, msg Message) context.Context {
	return context.WithValue(ctx, messageCtx, msg)
}

// MessageFromContext extracts a message from a Context.
func MessageFromContext(ctx context.Context) Message {
	return ctx.Value(messageCtx).(Message)
}

// NewContextWithNetID returns a new Context that carries net ID.
func NewContextWithNetID(ctx context.Context, netID int64) context.Context {
	return context.WithValue(ctx, netIDCtx, netID)
}

// NetIDFromContext returns a net ID from a Context.
func NetIDFromContext(ctx context.Context) int64 {
	return ctx.Value(netIDCtx).(int64)
}
