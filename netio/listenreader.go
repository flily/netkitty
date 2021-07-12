package netio

import "net"

type ListenReader struct {
	listener net.Listener
	conn     net.Conn
	signConn chan struct{}
}

func NewListenReader(listener net.Listener) *ListenReader {
	return &ListenReader{
		listener: listener,
		signConn: make(chan struct{}),
	}
}

func (r *ListenReader) Close() error {
	r.conn.Close()
	return r.listener.Close()
}

func (r *ListenReader) Read(buffer []byte) (int, error) {
	if r.conn == nil {
		conn, err := r.listener.Accept()
		if err != nil {
			return -1, err
		}
		r.conn = conn
		r.signConn <- struct{}{}
	}

	return r.conn.Read(buffer)
}

func (r *ListenReader) Write(buffer []byte) (int, error) {
	if r.conn == nil {
		// blocking Write() until connection is made.
		<-r.signConn
	}
	return r.conn.Write(buffer)
}
