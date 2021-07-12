package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/flily/netkitty/aio"
	"github.com/flily/netkitty/netio"
)

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
	return netio.NewListenReader(listener), nil
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
		return netio.NewListenReader(listener), nil

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

	forwarder := aio.NewForwarder(source, target)

	defer forwarder.Close()
	err := forwarder.Forward()
	if err != nil {
		log.Printf("io error: %s", err)
	}
}
