package proxy

import (
	"net"
)

type Connect struct {
	id   string
	host string
	port uint32
	conn net.Conn
}

type Connects struct {
	name     string
	min, max int32
	conns    []*Connect
}

var clients []*Connects

type Server struct {
	ID   string `json:"id"`
	Host string `json:"host"`
	Port int16  `json:"port"`
}

type Servers struct {
	Chat  []Server `json:"chat"`
	Logic []Server `json:"logic"`
}

var config struct {
	Dev Servers `json:"development"`
	Pro Servers `json:"production"`
}
