package router

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"

	"github.com/matrix-org/pinecone/types"
)

type PingType uint8

const (
	Ping PingType = iota
	Pong
)

const pingPreamble = "pineping"
const pingSize = len(pingPreamble) + (ed25519.PublicKeySize * 2) + 3

type PingPayload struct {
	pingType    PingType
	origin      types.PublicKey
	destination types.PublicKey
	hops        uint16
}

func (p *PingPayload) MarshalBinary(buffer []byte) (int, error) {
	if len(buffer) < pingSize {
		return 0, fmt.Errorf("buffer too small")
	}
	offset := copy(buffer, []byte(pingPreamble))
	buffer[offset] = uint8(p.pingType)
	offset++
	binary.BigEndian.PutUint16(buffer[offset:offset+2], p.hops)
	offset += 2
	offset += copy(buffer[offset:], p.origin[:ed25519.PublicKeySize])
	offset += copy(buffer[offset:], p.destination[:ed25519.PublicKeySize])
	return offset, nil
}

func (p *PingPayload) UnmarshalBinary(buffer []byte) (int, error) {
	if len(buffer) < pingSize {
		return 0, fmt.Errorf("buffer too small")
	}
	if string(buffer[:len(pingPreamble)]) != pingPreamble {
		return 0, fmt.Errorf("not a ping")
	}
	offset := len(pingPreamble)
	p.pingType = PingType(buffer[offset])
	offset++
	p.hops = binary.BigEndian.Uint16(buffer[offset : offset+2])
	offset += 2
	offset += copy(p.origin[:], buffer[offset:])
	offset += copy(p.destination[:], buffer[offset:])
	return offset, nil
}
