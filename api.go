package link

import (
	"errors"
	"io"
	"net"
	"time"
)

type Protocol interface {
	NewCodec(rw io.ReadWriteCloser) Codec
}

type ProtocolFunc func(rw io.ReadWriteCloser) Codec

func (pf ProtocolFunc) NewCodec(rw io.ReadWriteCloser) Codec {
	return pf(rw)
}

type Codec interface {
	Receive() (interface{}, error)
	Send(interface{}) error
	Close() error
}

type ClearSendChan interface {
	ClearSendChan(<-chan interface{})
}

func Listen(network, address string, protocol Protocol, sendChanSize int, handler Handler) (*Server, error) {
	if network == "test" {
		if testServerMap[address] != nil {
			return nil, errors.New("address has bind")
		}
		server := newServer(nil, address, protocol, sendChanSize, handler)
		testServerMap[address] = server
		return server, nil
	}
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return newServer(listener, address, protocol, sendChanSize, handler), nil
}

func Dial(network, address string, protocol Protocol, sendChanSize int) (*Session, error) {
	if network == "test" {
		server := testServerMap[address]
		if server == nil {
			return nil, errors.New("address error")
		}
		serverConn, clientConn := net.Pipe()
		go func() {
			codec := server.protocol.NewCodec(serverConn)
			session := server.manager.NewSession(codec, server.sendChanSize)
			server.handler.HandleSession(session)
		}()
		codec  := protocol.NewCodec(clientConn)
		return NewSession(codec, sendChanSize), nil
	}
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	codec  := protocol.NewCodec(conn)
	return NewSession(codec, sendChanSize), nil
}

func DialTimeout(network, address string, timeout time.Duration, protocol Protocol, sendChanSize int) (*Session, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	codec  := protocol.NewCodec(conn)
	return NewSession(codec, sendChanSize), nil
}

func accept(listener net.Listener) (net.Conn, error) {
	var tempDelay time.Duration
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			//if strings.Contains(err.Error(), "use of closed network connection") {
			//	return nil, io.EOF
			//}
			return nil, err
		}
		return conn, nil
	}
}
