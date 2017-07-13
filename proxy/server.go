package proxy

import (
	"crypto/tls"

	"zhao.com/lii/server"
)

func Start() {

}

type options struct {
	tlsCfg     *tls.Config
	codec      server.Codec
	onConnect  onConnectFunc
	onMessage  onMessageFunc
	onClose    onCloseFunc
	onError    onErrorFunc
	workerSize int  // numbers of worker go-routines
	bufferSize int  // size of buffered channel
	reconnect  bool // for ClientConn use only
}

type onConnectFunc func(WriteCloser) bool
type onMessageFunc func(server.Message, WriteCloser)
type onCloseFunc func(WriteCloser)
type onErrorFunc func(WriteCloser)
