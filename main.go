package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	. "github.com/polunzh/gdig/lib"
)

type QHeader struct {
	ID      uint16
	QR      uint8
	OPCode  uint8
	TC      uint8
	RD      uint8
	Z       uint8
	RCode   uint8
	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

type Question struct {
	QName  string
	QType  uint16
	QClass uint16
}

type Query struct {
	header   QHeader
	question Question
}

func (header QHeader) encode() []byte {
	var buffer bytes.Buffer

	// ID(16)
	binary.Write(&buffer, binary.BigEndian, header.ID)

	// 接下来的 8 个字节: QR(1), Opcode(4), AA(1), RD(1)
	binary.Write(&buffer, binary.BigEndian, byte(int(header.QR)<<7|int(header.OPCode)<<3|int(header.RD)))

	// 接下来的 8 个字节: RD(1) RA(1), Z(3), RCODE(4)
	binary.Write(&buffer, binary.BigEndian, byte(int(header.Z)<<4|int(header.RCode)<<4))

	binary.Write(&buffer, binary.BigEndian, header.QDCount) // QDCOUNT(16)
	binary.Write(&buffer, binary.BigEndian, header.ANCount) // ANCOUNT
	binary.Write(&buffer, binary.BigEndian, header.NSCount) // NSCOUNT
	binary.Write(&buffer, binary.BigEndian, header.ARCount) // ARCOUNT

	return buffer.Bytes()
}

func (question Question) encode() []byte {
	var buffer bytes.Buffer

	// 写入域名
	domainParts := strings.Split(question.QName, ".")
	for _, part := range domainParts {
		// 域名以 . 分隔的部分的长度
		binary.Write(&buffer, binary.BigEndian, byte(len(part)))

		for _, c := range part {
			binary.Write(&buffer, binary.BigEndian, uint8(c))
		}
	}
	// 写入 8 字节的 0, 结束 QNAME
	binary.Write(&buffer, binary.BigEndian, uint8(0))

	binary.Write(&buffer, binary.BigEndian, question.QType)
	binary.Write(&buffer, binary.BigEndian, question.QClass)

	return buffer.Bytes()
}

func (query Query) encode() []byte {
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, query.header.encode())
	binary.Write(&buffer, binary.BigEndian, query.question.encode())

	return buffer.Bytes()
}

func NewQuery(domain string, qType Type) []byte {
	rand.Seed(time.Now().UnixNano())
	id := uint16(rand.Intn(127))

	header := QHeader{
		ID:      id,
		QR:      0,
		OPCode:  0,
		TC:      0,
		RD:      1,
		Z:       0,
		RCode:   0,
		QDCount: 1,
	}

	question := Question{
		QName:  "baidu.com",
		QType:  A,
		QClass: IN,
	}

	return Query{header, question}.encode()
}

func query(data []byte) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	_, err = conn.Write(data)

	if err != nil {
		panic(err)
	}

	response := make([]byte, 4096) // TODO: []byte 的大小可能需要讨论

	_, err = conn.Read(response)
	fmt.Println(err)
	if err != nil {
		panic(err)
	}
}

func main() {
	queryData := NewQuery("d1.polunzh.com", A)
	query(queryData)
}
