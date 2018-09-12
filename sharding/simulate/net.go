package simulate

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/net/message"
	"io"
	"net"
	"time"
)

var (
	log = elog.NewLogger("sdsimulate", elog.DebugLog)
)

var netMsgChain chan message.EcoBallNetMsg
var listenPort string

func Sendto(addr string, port string, packet message.EcoBallNetMsg) error {
	addrPort := addr + ":" + port
	conn, err := net.DialTimeout("tcp", addrPort, 2*time.Second)
	if err != nil {
		log.Debug("connect to peer %s:%s error:%s", addr, port, err)
		return err
	}

	return send(conn, packet)

}

func Subscribe(port string) (<-chan message.EcoBallNetMsg, error) {
	netMsgChain = make(chan message.EcoBallNetMsg)

	listenPort = port
	go recvRoutine()

	return netMsgChain, nil
}

func recvRoutine() {
	l, err := net.Listen("tcp", "0.0.0.0:"+fmt.Sprint(listenPort))
	if err != nil {
		log.Error("start server listen error: %s", err)
		panic(err)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error("Listening() Failed to accept")
			continue
		}

		go recv(conn)
	}

	return
}

func send(conn net.Conn, packet message.EcoBallNetMsg) error {
	defer conn.Close()

	var length uint32
	length = uint32(len(packet.Data()) + 4)

	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.BigEndian, length)
	if err != nil {
		log.Error("write packet length error")
		return err
	}

	err = binary.Write(buf, binary.BigEndian, packet.Type())
	if err != nil {
		log.Error("write packet length error")
		return err
	}

	err = binary.Write(buf, binary.BigEndian, packet.Data())
	if err != nil {
		log.Error(" write packet  error")
		return err
	}

	_, err = conn.Write(buf.Bytes())
	return err
}

func recv(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	bl := make([]byte, 4)
	readerr := false

	for {
		_, err := io.ReadFull(reader, bl)
		if err != nil {
			//log.Debug("p2p recvRoutine read head error:%s", err)
			break
		}

		packetLen := binary.BigEndian.Uint32(bl)
		if packetLen > 1024*1024*10 {
			log.Error("drop packet wrong packet lenght %d", packetLen)
			break
		}

		buf := make([]byte, packetLen)
		readLen, err := io.ReadFull(reader, buf)
		if err != nil {
			log.Error("read data error:%s", err)
			break
		}

		if uint32(readLen) < packetLen {
			for {
				length, err := io.ReadFull(reader, buf[readLen:])
				if err != nil {
					log.Error("continue read data error:%s", err)
					return
				}

				readLen += length

				if uint32(readLen) < packetLen {
					continue
				} else if uint32(readLen) == packetLen {
					break
				} else {
					log.Error("continue read data length error:%s", err)
					readerr = true
					break
				}
			}
		}

		if readerr {
			break
		}

		packetType := uint32(binary.BigEndian.Uint32(buf))
		data := buf[4:packetLen]

		packet := message.New(packetType, data)

		log.Debug("recv packet type:%d", packetType)
		netMsgChain <- packet
	}
}
