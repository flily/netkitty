package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type ListenReader struct {
	listener net.Listener
	conn     net.Conn
}

func NewListenReader(listener net.Listener) *ListenReader {
	return &ListenReader{
		listener: listener,
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
	}

	return r.conn.Read(buffer)
}

func MakeListener(port int, useUDP bool) (io.ReadCloser, error) {
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

func MakeSource(s string, useUDP bool) (io.ReadCloser, error) {
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

func MakeTarget(s string, useUDP bool) (io.WriteCloser, error) {
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
	listenPort := flag.Int("listen", 0, "listen port")
	useUDP := flag.Bool("udp", false, "use UDP")
	flag.Parse()

	if *listenPort < 0 || *listenPort >= 65536 {
		fmt.Printf("invalid port number %d", *listenPort)
		return
	}

	var source io.ReadCloser = os.Stdin
	var target io.WriteCloser = os.Stdout

	if *listenPort != 0 {
		conn, err := MakeListener(*listenPort, *useUDP)
		if err != nil {
			fmt.Printf("listen error: %s", err)
			return
		}
		source = conn
	}

	if flag.NArg() > 0 {
		endpointString := flag.Arg(0)
		conn, err := MakeTarget(endpointString, *useUDP)
		if err != nil {
			fmt.Printf("dial error: %s", err)
			return
		}
		target = conn
	}

	defer source.Close()
	defer target.Close()

	tee := io.TeeReader(source, target)
	io.ReadAll(tee)
}
