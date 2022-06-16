package listener

import (
	"errors"
	"net"
)

type Listener struct {
	channel chan net.Conn
	addr    net.Addr
}

func New(addr net.Addr) *Listener {
	return &Listener{
		channel: make(chan net.Conn),
		addr:    addr,
	}
}

func (listener *Listener) Accept() (net.Conn, error) {
	conn, ok := <-listener.channel
	if !ok {
		return nil, errors.New("closed")
	}
	return conn, nil
}

func (listener *Listener) Close() error {
	return nil
}

func (listener *Listener) Addr() net.Addr {
	return listener.addr
}

func (listener *Listener) Handle(conn net.Conn) error {
	listener.channel <- conn
	return nil
}
