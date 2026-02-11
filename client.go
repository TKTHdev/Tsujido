package tsujido

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
)

type Submitter interface {
	Submit(ctx context.Context, op Operation) (Result, error)
	Close() error
}

// TCPClient connects to protocol nodes via TCP.
// writeAddr receives SET/DELETE, readAddr receives GET.
// For raft/pbft these are the same address; for chain, write=head, read=tail.
type TCPClient struct {
	writeAddr string
	readAddr  string

	writeConn net.Conn
	readConn  net.Conn // nil if same as writeConn

	nextReqID atomic.Uint64
	pending   map[uint64]chan Result
	mu        sync.Mutex

	writeMu sync.Mutex // serialize writes on writeConn
	readMu  sync.Mutex // serialize writes on readConn (if separate)

	closed chan struct{}
}

func NewTCPClient(writeAddr, readAddr string) (*TCPClient, error) {
	writeConn, err := net.Dial("tcp", writeAddr)
	if err != nil {
		return nil, fmt.Errorf("connect write addr %s: %w", writeAddr, err)
	}

	var readConn net.Conn
	if readAddr != writeAddr {
		readConn, err = net.Dial("tcp", readAddr)
		if err != nil {
			writeConn.Close()
			return nil, fmt.Errorf("connect read addr %s: %w", readAddr, err)
		}
	}

	c := &TCPClient{
		writeAddr: writeAddr,
		readAddr:  readAddr,
		writeConn: writeConn,
		readConn:  readConn,
		pending:   make(map[uint64]chan Result),
		closed:    make(chan struct{}),
	}

	go c.recvLoop(writeConn)
	if readConn != nil {
		go c.recvLoop(readConn)
	}

	return c, nil
}

func (c *TCPClient) Submit(ctx context.Context, op Operation) (Result, error) {
	reqID := c.nextReqID.Add(1)

	ch := make(chan Result, 1)
	c.mu.Lock()
	c.pending[reqID] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, reqID)
		c.mu.Unlock()
	}()

	data := EncodeRequest(reqID, op)

	// Choose connection: GET goes to readConn (if separate), everything else to writeConn
	conn := c.writeConn
	connMu := &c.writeMu
	if op.Type == OpGet && c.readConn != nil {
		conn = c.readConn
		connMu = &c.readMu
	}

	connMu.Lock()
	_, err := conn.Write(data)
	connMu.Unlock()
	if err != nil {
		return Result{}, fmt.Errorf("write: %w", err)
	}

	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	case result := <-ch:
		return result, nil
	case <-c.closed:
		return Result{}, fmt.Errorf("client closed")
	}
}

func (c *TCPClient) Close() error {
	select {
	case <-c.closed:
		return nil
	default:
		close(c.closed)
	}
	c.writeConn.Close()
	if c.readConn != nil {
		c.readConn.Close()
	}
	return nil
}

func (c *TCPClient) recvLoop(conn net.Conn) {
	for {
		payload, err := ReadFrame(conn)
		if err != nil {
			select {
			case <-c.closed:
			default:
			}
			return
		}

		reqID, result, err := DecodeResponse(payload)
		if err != nil {
			continue
		}

		c.mu.Lock()
		ch, ok := c.pending[reqID]
		c.mu.Unlock()

		if ok {
			select {
			case ch <- result:
			default:
			}
		}
	}
}
