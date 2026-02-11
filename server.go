package tsujido

import (
	"context"
	"log"
	"net"
)

// RequestHandler is implemented by each protocol.
// HandleRequest blocks until consensus is complete and the result is available.
type RequestHandler interface {
	HandleRequest(ctx context.Context, op Operation) (Result, error)
}

type Server struct {
	addr     string
	handler  RequestHandler
	listener net.Listener
}

func NewServer(addr string, handler RequestHandler) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
	}
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = ln
	log.Printf("[tsujido] Server listening on %s", s.addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			if s.listener == nil {
				return nil // closed
			}
			log.Printf("[tsujido] Accept error: %v", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) Close() error {
	if s.listener != nil {
		ln := s.listener
		s.listener = nil
		return ln.Close()
	}
	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	for {
		payload, err := ReadFrame(conn)
		if err != nil {
			return // client disconnected or error
		}

		reqID, op, err := DecodeRequest(payload)
		if err != nil {
			log.Printf("[tsujido] decode error: %v", err)
			return
		}

		result, err := s.handler.HandleRequest(context.Background(), op)
		if err != nil {
			result = Result{Success: false, Value: err.Error()}
		}

		resp := EncodeResponse(reqID, result)
		if _, err := conn.Write(resp); err != nil {
			return
		}
	}
}
