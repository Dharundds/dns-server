package handlers

import (
	"context"
	"dns-server/internal/constants"
	"encoding/binary"
	"net"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func getIpForDN(ctx context.Context, domainName string) string {
	if constants.Redis == nil {
		log.Error().Msg("There is Redis connection available")
		return ""
	}

	val, err := constants.Redis.Get(ctx, domainName)

	if err != nil && err == redis.Nil && strings.Contains(domainName, ".home") {
		log.Error().Msgf("Error while fetching value for key -> %s err -> %v", domainName, err)
		return "error"

	}

	return val
}

func HandleDNSQuery(ctx context.Context, wg *sync.WaitGroup, pc net.PacketConn, addr net.Addr, req []byte) {
	defer wg.Done()

	// Log the raw query for debugging mobile data issues
	log.Debug().Msgf("Received DNS query from %s, length: %d bytes", addr.String(), len(req))

	// Minimum DNS query size check
	if len(req) < 12 {
		log.Error().Msgf("DNS query too short from %s: %d bytes (minimum 12 required)", addr.String(), len(req))
		return
	}

	// Additional safety: check if we have basic DNS header
	defer func() {
		if r := recover(); r != nil {
			log.Error().Msgf("Panic recovered in DNS handler for %s: %v", addr.String(), r)
		}
	}()

	txid := req[0:2] // transaction ID
	flags := []byte{0x81, 0x80}

	// Check DNS header flags to identify unusual queries
	qr := (req[2] & 0x80) >> 7 // Query/Response bit
	if qr != 0 {
		log.Warn().Msgf("Received DNS response instead of query from %s", addr.String())
		return
	}

	qdCount := binary.BigEndian.Uint16(req[4:6])
	if qdCount == 0 {
		log.Warn().Msgf("DNS query with no questions from %s", addr.String())
		return
	}

	if qdCount > 10 { // Reasonable limit to prevent abuse
		log.Warn().Msgf("DNS query with too many questions (%d) from %s", qdCount, addr.String())
		return
	}

	// Find end of question section (offset) with bounds checking
	off := 12
	questionCount := 0

	for off < len(req) && req[off] != 0 && questionCount < int(qdCount) {
		if off >= len(req) {
			log.Error().Msgf("Malformed DNS query from %s: unexpected end of packet", addr.String())
			return
		}

		labelLen := int(req[off])
		if labelLen == 0 {
			break
		}

		// Check for DNS compression pointers (mobile networks sometimes use these)
		if labelLen >= 192 { // 0xC0 indicates compression
			log.Debug().Msgf("DNS compression pointer detected from %s", addr.String())
			off += 2 // Skip compression pointer
			break
		}

		// Validate label length is reasonable
		if labelLen > 63 { // DNS labels can't be longer than 63 bytes
			log.Error().Msgf("Invalid DNS label length %d from %s", labelLen, addr.String())
			return
		}

		// Check if we have enough bytes for the label
		if off+1+labelLen >= len(req) {
			log.Error().Msgf("Malformed DNS query from %s: label extends beyond packet (off=%d, labelLen=%d, reqLen=%d)",
				addr.String(), off, labelLen, len(req))
			return
		}
		off += labelLen + 1

		// Safety check to prevent infinite loops
		if off > 512 { // DNS packets shouldn't be larger than 512 bytes over UDP
			log.Error().Msgf("DNS query too large from %s: %d bytes", addr.String(), off)
			return
		}
	}

	// Check if we have enough bytes for the remaining fields (qtype + qclass)
	if off+5 > len(req) {
		log.Error().Msgf("Malformed DNS query from %s: insufficient bytes for qtype and qclass (off=%d, reqLen=%d)",
			addr.String(), off, len(req))
		return
	}

	off += 5 // zero byte + qtype(2) + qclass(2)

	// Bounds check for question extraction
	if off-4 > len(req) || off-4 < 12 {
		log.Error().Msgf("Malformed DNS query from %s: invalid question section bounds", addr.String())
		return
	}

	question := req[12 : off-4]
	domain := parseName(question)

	// Validate domain name
	if domain == "" {
		log.Warn().Msgf("Empty domain name from %s", addr.String())
		return
	}

	if len(domain) > 253 { // Max domain name length
		log.Warn().Msgf("Domain name too long (%d chars) from %s: %s", len(domain), addr.String(), domain)
		return
	}

	log.Info().Msgf("Domain name received from %s: %s", addr.String(), domain)

	// TODO: use dynamic domain
	ipstr := getIpForDN(ctx, domain)
	switch ipstr {
	case "":
		forwardTo := "1.1.1.1:53"
		log.Debug().Msgf("Forwarding query for %s from %s to %s", domain, addr.String(), forwardTo)
		forwardAndRespond(pc, addr, req, forwardTo)
		return
	case "error":
		log.Error().Msgf("Error retrieving IP for domain %s from %s", domain, addr.String())
		return
	}

	res := append([]byte{}, txid...)
	res = append(res, flags...)
	res = append(res, req[4:6]...) // QDCOUNT
	res = append(res, 0x00, 0x01)  // ANCOUNT
	res = append(res, 0x00, 0x00)  // NSCOUNT
	res = append(res, 0x00, 0x00)  // ARCOUNT

	// Question - bounds check
	if off > len(req) || 12 > len(req) {
		log.Error().Msgf("Malformed DNS query from %s: invalid question for response", addr.String())
		return
	}
	res = append(res, req[12:off]...)

	// Answer section
	res = append(res, 0xc0, 0x0c)             // Name pointer to offset 12
	res = append(res, 0x00, 0x01)             // Type A
	res = append(res, 0x00, 0x01)             // Class IN
	res = append(res, 0x00, 0x00, 0x00, 0x3c) // TTL 60s
	res = append(res, 0x00, 0x04)             // RDLENGTH

	ip := net.ParseIP(ipstr).To4()
	if ip == nil {
		log.Error().Msgf("Invalid IP address %s for domain %s from %s", ipstr, domain, addr.String())
		return
	}
	res = append(res, ip...)

	log.Debug().Msgf("Sending DNS response to %s for %s -> %s", addr.String(), domain, ipstr)
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
	maxIterations := 100 // Prevent infinite loops
	iterations := 0

	for i < len(q) && iterations < maxIterations {
		iterations++

		if i >= len(q) {
			log.Error().Msg("Malformed DNS name: index out of bounds")
			break
		}

		length := int(q[i])
		if length == 0 {
			break
		}

		// Check for DNS compression pointers
		if length >= 192 { // 0xC0 indicates compression
			log.Debug().Msg("DNS compression pointer in name parsing")
			// For now, we'll just stop parsing at compression pointers
			// A full implementation would follow the pointer
			break
		}

		// Validate label length
		if length > 63 { // DNS labels can't be longer than 63 bytes
			log.Error().Msgf("Invalid DNS label length in parseName: %d", length)
			break
		}

		// Bounds check for label
		if i+1+length > len(q) {
			log.Error().Msgf("Malformed DNS name: label extends beyond data (i=%d, length=%d, qLen=%d)", i, length, len(q))
			break
		}

		i++
		if i+length <= len(q) {
			labelBytes := q[i : i+length]
			// Validate that label contains valid characters
			label := string(labelBytes)
			if len(label) > 0 {
				parts = append(parts, label)
			}
		}
		i += length
	}

	if iterations >= maxIterations {
		log.Error().Msg("DNS name parsing exceeded maximum iterations")
	}

	result := strings.Join(parts, ".")
	log.Debug().Msgf("Parsed DNS name: %s", result)
	return result
}
