package aio

import (
	"errors"
	"io"
)

type Forwarder interface {
	Forward() error
	Close() error
}

type forwarder struct {
	source Duplexer
	target Duplexer
}

func NewForwarder(source Duplexer, target Duplexer) Forwarder {
	return &forwarder{
		source: source,
		target: target,
	}
}

func (f *forwarder) Close() error {
	if err := f.source.Close(); err != nil {
		return err
	}

	if err := f.target.Close(); err != nil {
		return err
	}

	return nil
}

func (f *forwarder) Forward() error {
	f.source.Run()
	f.target.Run()

	var err error
	rc, wc := false, false
forloop:
	for {
		select {
		case data := <-f.source.Recv():
			if data.Err != nil {
				if !errors.Is(data.Err, io.EOF) {
					err = data.Err
					break forloop
				}

				wc = true
				if rc && wc {
					break forloop
				}
			}

			f.target.Send() <- data.Data

		case data := <-f.target.Recv():
			if data.Err != nil {
				if !errors.Is(data.Err, io.EOF) {
					err = data.Err
					break forloop
				}

				rc = true
				if rc && wc {
					break forloop
				}
			}

			f.source.Send() <- data.Data
		}
	}

	return err
}
