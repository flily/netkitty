package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/flily/netkitty/aio"
)

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
		<-r.signConn
	}
	return r.conn.Write(buffer)
}

func MakeListener(port int, useUDP bool) (io.ReadWriteCloser, error) {
	network := "tcp"
	if useUDP {
		network = "udp"
	}

	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	return NewListenReader(listener), nil
}

func MakeSource(s string, useUDP bool) (io.ReadWriteCloser, error) {
	network := "tcp"
	if useUDP {
		network = "udp"
	}

	if strings.Contains(s, ":") {
		// IPv4 or IPv6 address
		listener, err := net.Listen(network, s)
		if err != nil {
			return nil, err
		}
		return NewListenReader(listener), nil

	} else {
		// File
		file, err := os.Open(s)
		if err != nil {
			return nil, err
		}

		return file, nil
	}
}

func MakeTarget(s string, useUDP bool) (io.ReadWriteCloser, error) {
	network := "tcp"
	if useUDP {
		network = "udp"
	}

	if strings.Contains(s, ":") {
		// IPv4 or IPv6 address
		conn, err := net.Dial(network, s)
		if err != nil {
			return nil, err
		}
		return conn, nil

	} else {
		// File
		file, err := os.Create(s)
		if err != nil {
			return nil, err
		}

		return file, nil
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	listenPort := flag.Int("listen", 0, "listen port")
	useUDP := flag.Bool("udp", false, "use UDP")
	flag.Parse()

	if *listenPort < 0 || *listenPort >= 65536 {
		log.Printf("invalid port number %d", *listenPort)
		return
	}

	source := aio.NewSimpleDuplexer(aio.NewTerminal())
	target := aio.NewSimpleDuplexer(aio.NewTerminal())

	if *listenPort != 0 {
		conn, err := MakeListener(*listenPort, *useUDP)
		if err != nil {
			log.Printf("listen error: %s", err)
			return
		}
		source = aio.NewSimpleDuplexer(conn)
	}

	if flag.NArg() > 0 {
		endpointString := flag.Arg(0)
		conn, err := MakeTarget(endpointString, *useUDP)
		if err != nil {
			log.Printf("dial error: %s", err)
			return
		}
		target = aio.NewSimpleDuplexer(conn)
	}

	defer source.Close()
	defer target.Close()

	source.Run()
	target.Run()

	rc, wc := false, false

forloop:
	for {
		select {
		case data := <-source.Recv():
			if data.Err != nil {
				if !errors.Is(data.Err, io.EOF) {
					log.Printf("write error: %s", data.Err)
					break forloop
				}

				wc = true
				if rc && wc {
					break forloop
				}
			}

			target.Send() <- data.Data

		case data := <-target.Recv():
			if data.Err != nil {
				if !errors.Is(data.Err, io.EOF) {
					log.Printf("read error: %s", data.Err)
					break forloop
				}

				rc = true
				if rc && wc {
					break forloop
				}
			}

			source.Send() <- data.Data
		}
	}
}
