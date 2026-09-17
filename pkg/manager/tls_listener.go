package manager

import (
	"bytes"
	"io"
	"net"
	"sync"
	"time"
)

// smartTLSListener accepts TCP connections and demultiplexes TLS vs plaintext HTTP on the exact same port.
// - If a client sends a TLS Handshake (0x16), it is forwarded directly to the TLS handler.
// - If a client sends plaintext HTTP (GET, POST, etc.), it is forwarded to the plaintext HTTP handler.
type smartTLSListener struct {
	net.Listener
	tlsConns   chan net.Conn
	plainConns chan net.Conn
	done       chan struct{}
	once       sync.Once
}

func newSmartTLSListener(ln net.Listener) *smartTLSListener {
	sl := &smartTLSListener{
		Listener:   ln,
		tlsConns:   make(chan net.Conn, 128),
		plainConns: make(chan net.Conn, 128),
		done:       make(chan struct{}),
	}
	go sl.acceptLoop()
	return sl
}

func (sl *smartTLSListener) acceptLoop() {
	for {
		conn, err := sl.Listener.Accept()
		if err != nil {
			select {
			case <-sl.done:
				return
			default:
				continue
			}
		}
		go sl.classifyConn(conn)
	}
}

func (sl *smartTLSListener) classifyConn(conn net.Conn) {
	// Read 1 byte with a short deadline to inspect the protocol header
	_ = conn.SetReadDeadline(time.Now().Add(4 * time.Second))
	buf := make([]byte, 1)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		conn.Close()
		return
	}
	_ = conn.SetReadDeadline(time.Time{})

	replayConn := &bufferedConn{
		Conn: conn,
		r:    io.MultiReader(bytes.NewReader(buf[:n]), conn),
	}

	if buf[0] == 0x16 { // TLS record type Handshake
		select {
		case sl.tlsConns <- replayConn:
		case <-sl.done:
			conn.Close()
		}
		return
	}

	// Plaintext HTTP request (GET, POST, HEAD, OPTIONS, PUT, DELETE, etc.)
	select {
	case sl.plainConns <- replayConn:
	case <-sl.done:
		conn.Close()
	}
}

func (sl *smartTLSListener) TLSListener() net.Listener {
	return &subListener{
		AddrVal: sl.Listener.Addr(),
		conns:   sl.tlsConns,
		done:    sl.done,
	}
}

func (sl *smartTLSListener) PlainListener() net.Listener {
	return &subListener{
		AddrVal: sl.Listener.Addr(),
		conns:   sl.plainConns,
		done:    sl.done,
	}
}

func (sl *smartTLSListener) Close() error {
	var err error
	sl.once.Do(func() {
		close(sl.done)
		err = sl.Listener.Close()
	})
	return err
}

type subListener struct {
	AddrVal net.Addr
	conns   chan net.Conn
	done    chan struct{}
}

func (s *subListener) Accept() (net.Conn, error) {
	select {
	case conn, ok := <-s.conns:
		if !ok {
			return nil, net.ErrClosed
		}
		return conn, nil
	case <-s.done:
		return nil, net.ErrClosed
	}
}

func (s *subListener) Close() error {
	return nil
}

func (s *subListener) Addr() net.Addr {
	return s.AddrVal
}

type bufferedConn struct {
	net.Conn
	r io.Reader
}

func (c *bufferedConn) Read(p []byte) (int, error) {
	return c.r.Read(p)
}
