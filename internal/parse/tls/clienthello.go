package tls

import (
	"encoding/binary"
)

type ClientHello struct {
	SNI string
}

func (m *ClientHello) Unmarshal(payload []byte) error {

	payloadLength := uint16(len(payload))
	if payloadLength < uint16(4) {
		return UnmarshalClientHelloError
	}

	handshakeProtocol := payload[4]
	// Only attempt to match on client hellos
	if handshakeProtocol != 0x01 {
		return UnmarshalNoTLSHandshakeError
	}

	offset, baseOffset, extensionOffset := uint16(0), uint16(42), uint16(2)
	if baseOffset+2 > uint16(len(payload)) {
		return UnmarshalClientHelloError
	}

	// Get the length of the session ID
	sessionIdLength := uint16(payload[baseOffset])
	if (sessionIdLength + baseOffset + 2) > payloadLength {
		return UnmarshalClientHelloError
	}

	// Get the length of the ciphers
	cipherLenStart := baseOffset + sessionIdLength + 1
	cipherLen := uint16(payload[cipherLenStart])<<8 | uint16(payload[cipherLenStart+1])

	offset = baseOffset + sessionIdLength + cipherLen + 2
	if offset > payloadLength {
		return UnmarshalClientHelloError
	}

	// Get the length of the compression methods list
	compressionLen := uint16(payload[offset+1])
	offset += compressionLen + 2
	if offset > payloadLength {
		return UnmarshalClientHelloError
	}

	// Get the length of the extensions
	extensionsLen := binary.BigEndian.Uint16(payload[offset : offset+2])

	// Add the full offset to were the extensions start
	extensionOffset += offset

	if extensionsLen > payloadLength {
		return UnmarshalClientHelloError
	}

	for extensionOffset < extensionsLen+offset {
		extensionId := binary.BigEndian.Uint16(payload[extensionOffset : extensionOffset+2])
		extensionOffset += 2

		extensionLen := binary.BigEndian.Uint16(payload[extensionOffset : extensionOffset+2])
		extensionOffset += 2

		if extensionId == 0 {
			// We don't need the server name list length or name_type, so skip that
			extensionOffset += 3

			// Get the length of the domain name
			nameLength := binary.BigEndian.Uint16(payload[extensionOffset : extensionOffset+2])
			extensionOffset += 2

			m.SNI = string(payload[extensionOffset : extensionOffset+nameLength])
			return nil
		}
		extensionOffset += extensionLen
	}
	return nil
}
