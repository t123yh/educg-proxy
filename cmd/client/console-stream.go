package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
)

type WebsocketConsoleStream struct {
	downlink <-chan byte
	uplink   chan<- byte
	context  context.Context
	close    context.CancelFunc
}

func CreateConsoleStream(dl <- chan byte, ul chan<- byte, ctx context.Context) *WebsocketConsoleStream {
	mctx, mcancel := context.WithCancel(ctx)
	return &WebsocketConsoleStream{downlink: dl, uplink: ul, context: mctx, close: mcancel}
}

func (s WebsocketConsoleStream) Read(p []byte) (int, error) {
	cap := len(p) * 2
	sb := bytes.Buffer{}
	sb.Grow(cap)
	for i := 0; i < cap; i++ {
		select {
		case <- s.context.Done():
			return 0, errors.New("context canceled")
		case v := <-s.downlink:
			sb.WriteByte(v)
		}
	}
	_, err := hex.Decode(p, sb.Bytes())
	if err != nil {
		panic("解码失败！")
	}
	return len(p), nil
}

func (s WebsocketConsoleStream) Write(p []byte) (n int, err error) {
	encoded := make([]byte, len(p) * 2)
	hex.Encode(encoded, p)
	for i := 0; i < len(encoded); i++ {
		select {
		case <- s.context.Done():
			return i, errors.New("context canceled")
		case s.uplink <- encoded[i]:
		}
	}
	return len(p), nil
}

func (s WebsocketConsoleStream) Close() error {
	s.close()
	return nil
}
