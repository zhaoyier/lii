package client

import "crypto/tls"

type onConnectFunc func(*ClientConn) bool
type onMessageFunc func(*Message, *ClientConn)
type onCloseFunc func(*ClientConn)
type onErrorFunc func(*ClientConn)

type options struct {
	tlsCfg     *tls.Config
	codec      Codec
	onConnect  onConnectFunc
	onMessage  onMessageFunc
	onClose    onCloseFunc
	onError    onErrorFunc
	workerSize int  // numbers of worker go-routines
	bufferSize int  // size of buffered channel
	reconnect  bool // for ClientConn use only
}

// ReconnectOption returns a ServerOption that will make ClientConn reconnectable.
func ReconnectOption() ServerOption {
	return func(o *options) {
		o.reconnect = true
	}
}

// OnConnectOption returns a ServerOption that will set callback to call when new
// client connected.
func OnConnectOption(cb func(*ClientConn) bool) ServerOption {
	return func(o *options) {
		o.onConnect = cb
	}
}

// OnMessageOption returns a ServerOption that will set callback to call when new
// message arrived.
func OnMessageOption(cb func(*Message, *ClientConn)) ServerOption {
	return func(o *options) {
		o.onMessage = cb
	}
}

// OnCloseOption returns a ServerOption that will set callback to call when client
// closed.
func OnCloseOption(cb func(*ClientConn)) ServerOption {
	return func(o *options) {
		o.onClose = cb
	}
}

// OnErrorOption returns a ServerOption that will set callback to call when error
// occurs.
func OnErrorOption(cb func(*ClientConn)) ServerOption {
	return func(o *options) {
		o.onError = cb
	}
}

// ServerOption sets server options.
type ServerOption func(*options)
