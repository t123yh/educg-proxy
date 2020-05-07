package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"os"
)

type StdioConsoleStream struct {
	reader *bufio.Reader
	writer *bufio.Writer
	context context.Context
	close    context.CancelFunc
}

func CreateConsoleStream(ctx context.Context) *StdioConsoleStream {
	mctx, mcancel := context.WithCancel(ctx)
	return &StdioConsoleStream{reader: bufio.NewReader(os.Stdin), writer: bufio.NewWriter(os.Stdout), context: mctx, close: mcancel}
}

func (s StdioConsoleStream) Read(p []byte) (int, error) {
	encoded := make([]byte, hex.EncodedLen(len(p)))
	// _, err := s.reader.Read(encoded)
	// println("!")
	for pos := 0; pos < len(encoded); {
		b, err := s.reader.ReadByte()
		if err != nil {
			panic("控制台读取失败！")
		}
		if b != '\r' && b != '\n' {
			encoded[pos] = b
			pos++
		}
	}
	// println(",")
	_, err := hex.Decode(p, encoded)
	if err != nil {
		panic("解码失败！")
	}
	return len(p), nil
}

func (s StdioConsoleStream) Write(p []byte) (n int, err error) {
	// println("?")
	size := hex.EncodedLen(len(p))
	encoded := make([]byte, size + 1)
	hex.Encode(encoded, p)
	encoded[size] = '\n'
	_, err = s.writer.Write(encoded)
	s.writer.Flush()
	if err != nil {
		panic("控制台写入失败！")
	}
	return len(p), nil
}

func (s StdioConsoleStream) Close() error {
	s.close()
	return nil
}

