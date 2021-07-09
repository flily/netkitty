package aio

import (
	"io"
	"os"
)

type TerminalDuplex struct {
	stdout io.WriteCloser
	stdin  io.ReadCloser
}

func NewTerminal() *TerminalDuplex {
	return &TerminalDuplex{
		stdout: os.Stdout,
		stdin:  os.Stdin,
	}
}

func (t *TerminalDuplex) Read(b []byte) (int, error) {
	return t.stdin.Read(b)
}

func (t *TerminalDuplex) Write(b []byte) (int, error) {
	return t.stdout.Write(b)
}

func (t *TerminalDuplex) Close() error {
	if t.stdin != os.Stdin {
		err := t.stdin.Close()
		if err != nil {
			return err
		}
	}

	if t.stdout != os.Stdout {
		err := t.stdout.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
