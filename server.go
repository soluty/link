package link

import "net"

var testServerMap map[string]*Server = map[string]*Server{}

type Server struct {
	manager      *Manager
	listener     net.Listener
	address      string
	protocol     Protocol
	handler      Handler
	sendChanSize int
}

type Handler interface {
	HandleSession(*Session)
}

var _ Handler = HandlerFunc(nil)

type HandlerFunc func(*Session)

func (f HandlerFunc) HandleSession(session *Session) {
	f(session)
}

func newServer(listener net.Listener, address string, protocol Protocol, sendChanSize int, handler Handler) *Server {
	return &Server{
		manager:      newManager(),
		listener:     listener,
		protocol:     protocol,
		address:      address,
		handler:      handler,
		sendChanSize: sendChanSize,
	}
}

func (server *Server) Listener() net.Listener {
	return server.listener
}

func (server *Server) Serve() error {
	if server.listener == nil {
		return nil
	}
	for {
		conn, err := accept(server.listener)
		if err != nil {
			return err
		}
		go func() {
			codec := server.protocol.NewCodec(conn)
			session := server.manager.NewSession(codec, server.sendChanSize)
			server.handler.HandleSession(session)
		}()
	}
}

func (server *Server) GetSession(sessionID uint64) *Session {
	return server.manager.GetSession(sessionID)
}

func (server *Server) Stop() {
	if server.listener != nil {
		server.listener.Close()
		server.manager.Dispose()
	} else {
		server.manager.Dispose()
		delete(testServerMap, server.address)
	}
}
