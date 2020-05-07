package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/xtaci/kcptun/generic"
	"github.com/xtaci/smux"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
)


func handleClient(p1 *smux.Stream, sess *smux.Session) {
	pbuf := make([]byte, 1)
	n, err := p1.Read(pbuf)
	if n != len(pbuf) || err != nil {
		return
	}
	switch pbuf[0] {
	case 0: // connect
		handleForward(p1)
	case 1: // ping signal
		p1.Write([]byte{55})
		p1.Close()
	case 2:
		p1.Write([]byte{66})
		p1.Close()
		sess.Close()
	}
}

func handleForward(p1 *smux.Stream) {
	println("!")
	defer p1.Close()
	buf := make([]byte, 6)
	n := 0
	for n < len(buf) {
		r, err := p1.Read(buf[n:])
		if err != nil {
			p1.Write([]byte{2})
			return
		}
		n += r
	}

	println(",")
	dstIP := net.IPv4(buf[0], buf[1], buf[2], buf[3])
	port := int(binary.LittleEndian.Uint16(buf[4:]))
	// dstAddr = fmt.Sprintf("%s:%d", dstIP.String(), port)

	p2, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: dstIP, Port: port})
	if err != nil {
		// log.Print("远端连接失败")
		p1.Write([]byte{1})
		p1.Close()
		return
	}
	p1.Write([]byte{99})

	defer p2.Close()

	// start tunnel & wait for tunnel termination
	streamCopy := func(dst io.Writer, src io.ReadCloser) {
		if _, err := generic.Copy(dst, src); err != nil {
			if err == smux.ErrInvalidProtocol {
				log.Println("smux", err, "in:", fmt.Sprint(p1.RemoteAddr(), "(", p1.ID(), ")"), "out:", p2.RemoteAddr())
			}
		}
		p1.Close()
		p2.Close()
	}

	go streamCopy(p2, p1)
	streamCopy(p1, p2)
}


func main() {
	cmd := exec.Command("stty", "-echo")
	cmd.Start()
	cmd.Wait()
	argsWithoutProg := os.Args[1:]

	print(argsWithoutProg[0]) // 打印同步字串

	stream := CreateConsoleStream(context.TODO())

	session, err := smux.Server(stream, smux.DefaultConfig())

	if err != nil {
		panic("创建 yamux 服务失败。" + err.Error())
	}


	for {
		stream, err := session.AcceptStream()
		if err != nil {
			panic (err.Error())
		}

		go handleClient(stream, session)
	}
}
