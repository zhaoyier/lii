package proxy

import (
	"context"
	"net"
	"sync"
	"time"

	"zhao.com/lii/server"
)

// WriteCloser is the interface that groups Write and Close methods.
type WriteCloser interface {
	Write(server.Message) error
	Close()
}

// ClientConn represents a client connection to a TCP server.
type ClientConn struct {
	addr      string
	opts      options
	netid     int64
	rawConn   net.Conn
	once      *sync.Once
	wg        *sync.WaitGroup
	sendCh    chan []byte
	handlerCh chan server.MessageHandler
	timing    *server.TimingWheel
	mu        sync.Mutex // guards following
	name      string
	heart     int64
	pending   []int64
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewClientConn returns a new client connection which has not started to
// serve requests yet.
func NewClientConn(netid int64, c net.Conn, opt ...server.ServerOption) *ClientConn {
	var opts options
	// for _, o := range opt {
	// 	//o(&opts)
	// }
	if opts.codec == nil {
		opts.codec = server.TypeLengthValueCodec{}
	}
	if opts.bufferSize <= 0 {
		opts.bufferSize = server.BufferSize256
	}
	return newClientConnWithOptions(netid, c, opts)
}

// func (cc *ClientConn) Start() {
// 	holmes.Infof("conn start, <%v -> %v>\n", cc.rawConn.LocalAddr(), cc.rawConn.RemoteAddr())
// 	onConnect := cc.opts.onConnect
// 	if onConnect != nil {
// 		onConnect(cc)
// 	}

// 	loopers := []func(ClientConn, *sync.WaitGroup){readLoop, writeLoop, handleLoop}
// 	for _, l := range loopers {
// 		looper := l
// 		cc.wg.Add(1)
// 		go looper(cc, cc.wg)
// 	}
// }

func newClientConnWithOptions(netid int64, c net.Conn, opts options) *ClientConn {
	cc := &ClientConn{
		addr:      c.RemoteAddr().String(),
		opts:      opts,
		netid:     netid,
		rawConn:   c,
		once:      &sync.Once{},
		wg:        &sync.WaitGroup{},
		sendCh:    make(chan []byte, opts.bufferSize),
		handlerCh: make(chan server.MessageHandler, opts.bufferSize),
		heart:     time.Now().UnixNano(),
	}
	cc.ctx, cc.cancel = context.WithCancel(context.Background())
	cc.timing = server.NewTimingWheel(cc.ctx)
	cc.name = c.RemoteAddr().String()
	cc.pending = []int64{}
	return cc
}
