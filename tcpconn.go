package tcpconn

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

type Conn struct {
	n     int
	binds []string
	read  chan struct{}
	write chan struct{}
	close chan struct{}
	sync.Mutex
}

func New(binds []string) *Conn {
	return &Conn{
		binds: binds,
		read:  make(chan struct{}),
		write: make(chan struct{}),
		close: make(chan struct{}),
	}
}

func (c *Conn) StartRead() {
	close(c.read)
}
func (c *Conn) StartWrite() {
	close(c.write)
}
func (c *Conn) StartClose() {
	close(c.close)
}

func (c *Conn) GetLAddr() string {
	if len(c.binds) < 1 {
		return ""
	}
	c.Lock()
	i := c.n % len(c.binds)
	c.n = c.n + 1
	a := c.binds[i]
	c.Unlock()
	return a
}

func (c *Conn) Run(ctx context.Context, raddr string, dur time.Duration) error {
	in := []byte("hello")
	buf := make([]byte, len(in))
	var d net.Dialer
	d.LocalAddr = &net.TCPAddr{
		IP: net.ParseIP(c.GetLAddr()),
	}
	conn, err := d.DialContext(ctx, "tcp", raddr)
	if err != nil {
		return fmt.Errorf("c.Run d.DialContext: %w", err)
	}

	<-c.write
	if err := conn.SetDeadline(time.Now().Add(dur)); err != nil {
		defer conn.Close()
		return fmt.Errorf("c.Run setWriteDeadline: %w", err)
	}
	if _, err := conn.Write(in); err != nil {
		defer conn.Close()
		return fmt.Errorf("c.Run conn.Write: %w", err)
	}

	<-c.read
	if err := conn.SetDeadline(time.Now().Add(dur)); err != nil {
		defer conn.Close()
		return fmt.Errorf("c.Run setReadDeadline: %w", err)
	}
	n, err := conn.Read(buf)
	if err != nil {
		defer conn.Close()
		return fmt.Errorf("c.Run conn.Read: %w", err)
	}
	if n != len(in) {
		defer conn.Close()
		return fmt.Errorf("got reply of %d expecting %d, buf: %v", n, len(in), buf)
	}

	<-c.close
	if err := conn.Close(); err != nil {
		return fmt.Errorf("c.Run conn.Close: %w", err)
	}
	return nil
}
