package aio

import (
	"io"
)

type DuplexData struct {
	Data []byte
	Err  error
}

type Duplexer interface {
	Send() chan<- []byte
	Recv() <-chan DuplexData
	Run() error
	Close() error
}

type SimpleDuplexPipe struct {
	sendCh     chan []byte
	recvCh     chan DuplexData
	isWritable bool
	isReadable bool
	rw         io.ReadWriteCloser
}

func NewSimpleDuplexer(o io.ReadWriteCloser) *SimpleDuplexPipe {
	p := &SimpleDuplexPipe{
		rw:         o,
		sendCh:     make(chan []byte),
		recvCh:     make(chan DuplexData),
		isWritable: true,
		isReadable: true,
	}

	return p
}

func (p *SimpleDuplexPipe) Close() error {
	return p.rw.Close()
}

func (p *SimpleDuplexPipe) Send() chan<- []byte {
	return p.sendCh
}

func (p *SimpleDuplexPipe) Recv() <-chan DuplexData {
	return p.recvCh
}

func (p *SimpleDuplexPipe) Run() error {
	go p.loopRecv()
	go p.loopSend()
	return nil
}

func (p *SimpleDuplexPipe) loopSend() {
	for {
		data, ok := <-p.sendCh
		if !ok {
			break
		}

		_, err := p.rw.Write(data)
		if err != nil {
			break
		}
	}
}

func (p *SimpleDuplexPipe) loopRecv() {
	for {
		buf := make([]byte, 8192)
		n, err := p.rw.Read(buf)
		if err != nil {
			p.recvCh <- DuplexData{Err: err}
			break
		}

		p.recvCh <- DuplexData{Data: buf[:n]}
	}
}
