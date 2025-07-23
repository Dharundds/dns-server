package handlers

import (
	"dns-server/utils"
	"encoding/binary"
	"log"
	"net"
	"strings"
	"sync"
)

func HandleDNSQuery(wg *sync.WaitGroup, pc net.PacketConn, addr net.Addr, req []byte) {
	defer wg.Done()
	txid := req[0:2] // transaction ID
	flags := []byte{0x81, 0x80}
	qdCount := binary.BigEndian.Uint16(req[4:6])
	if qdCount == 0 {
		return
	}

	// Find end of question section (offset)
	off := 12
	for req[off] != 0 {
		off += int(req[off]) + 1
	}
	off += 5 // zero byte + qtype(2) + qclass(2)

	question := req[12 : off-4]
	domain := parseName(question)
	// TODO: use dynamic domain
	if domain != "vm.home.local" {
		forwardTo := "1.1.1.1:53"
		forwardAndRespond(pc, addr, req, forwardTo)

		return // ignore unknown domains
	}

	// Start building response
	res := append([]byte{}, txid...)
	res = append(res, flags...)
	res = append(res, req[4:6]...) // QDCOUNT
	res = append(res, 0x00, 0x01)  // ANCOUNT
	res = append(res, 0x00, 0x00)  // NSCOUNT
	res = append(res, 0x00, 0x00)  // ARCOUNT

	// Question
	res = append(res, req[12:off]...)

	// Answer section
	res = append(res, 0xc0, 0x0c)             // Name pointer to offset 12
	res = append(res, 0x00, 0x01)             // Type A
	res = append(res, 0x00, 0x01)             // Class IN
	res = append(res, 0x00, 0x00, 0x00, 0x3c) // TTL 60s
	res = append(res, 0x00, 0x04)             // RDLENGTH
	ipstr := utils.GetTsIP()
	if ipstr == "" {
		log.Println("Error cannot find tailscale inf")
		return
	}
	ip := net.ParseIP(ipstr).To4()
	res = append(res, ip...)

	pc.WriteTo(res, addr)
}

func forwardAndRespond(pc net.PacketConn, addr net.Addr, req []byte, forwardAddr string) {
	conn, err := net.Dial("udp", forwardAddr)
	if err != nil {
		return
	}
	defer conn.Close()

	conn.Write(req)
	reply := make([]byte, 512)
	n, err := conn.Read(reply)
	if err != nil {
		return
	}
	pc.WriteTo(reply[:n], addr)
}

func parseName(q []byte) string {
	var parts []string
	i := 0
	for {
		length := int(q[i])
		if length == 0 {
			break
		}
		i++
		parts = append(parts, string(q[i:i+length]))
		i += length
	}
	return strings.Join(parts, ".")
}
