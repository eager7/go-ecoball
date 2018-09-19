package simulate

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/common/elog"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"io"
	"net"
	"time"
)

var (
	log = elog.NewLogger("sdsimulate", elog.DebugLog)
)

var netMsgChain chan *sc.NetPacket
var listenPort string

func Sendto(addr string, port string, packet *sc.NetPacket) error {
	addrPort := addr + ":" + port
	conn, err := net.DialTimeout("tcp", addrPort, 100*time.Millisecond)
	if err != nil {
		log.Debug("connect to peer ", addr, " port ", port, " ", err)
		return err
	}

	log.Debug("send to peer ", addr, " port ", port)
	return send(conn, packet)

}

func Subscribe(port string, chanSize uint16) (<-chan *sc.NetPacket, error) {
	netMsgChain = make(chan *sc.NetPacket, chanSize)

	listenPort = port
	go recvRoutine()

	return netMsgChain, nil
}

func recvRoutine() {
	l, err := net.Listen("tcp", "0.0.0.0:"+fmt.Sprint(listenPort))
	if err != nil {
		log.Error("start server listen error ", err)
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

func send(conn net.Conn, packet *sc.NetPacket) error {
	defer conn.Close()

	data, err := json.Marshal(packet)
	if err != nil {
		log.Error("wrong packet")
		return err
	}

	var length uint32
	length = uint32(len(data))

	buf := &bytes.Buffer{}
	err = binary.Write(buf, binary.BigEndian, length)
	if err != nil {
		log.Error("write packet length error")
		return err
	}

	err = binary.Write(buf, binary.BigEndian, data)
	if err != nil {
		log.Error(" write packet  error")
		return err
	}

	_, err = conn.Write(buf.Bytes())
	if err != nil {
		log.Error("conn write error ", err)
	}

	log.Debug("send packet packet type ", packet.PacketType, " block type ", packet.BlockType)

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
			log.Error("drop packet wrong packet lenght ", packetLen)
			break
		}

		buf := make([]byte, packetLen)
		readLen, err := io.ReadFull(reader, buf)
		if err != nil {
			log.Error("read data error s", err)
			break
		}

		if uint32(readLen) < packetLen {
			for {
				length, err := io.ReadFull(reader, buf[readLen:])
				if err != nil {
					log.Error("continue read data error ", err)
					return
				}

				readLen += length

				if uint32(readLen) < packetLen {
					continue
				} else if uint32(readLen) == packetLen {
					break
				} else {
					log.Error("continue read data length error ", err)
					readerr = true
					break
				}
			}
		}

		if readerr {
			break
		}

		var packet sc.NetPacket
		err = json.Unmarshal(buf, &packet)
		if err != nil {
			log.Error("unmarshal packet error")
			return
		}

		log.Debug("recv packet ", packet.PacketType, "block ", packet.BlockType)
		netMsgChain <- &packet
	}
}
