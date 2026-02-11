package tsujido

import (
	"encoding/binary"
	"errors"
	"io"
)

// Wire format:
//
// Request:
//   [4 bytes: length][8 bytes: requestID][1 byte: opType][2 bytes: keyLen][key][2 bytes: valLen][value]
//
// Response:
//   [4 bytes: length][8 bytes: requestID][1 byte: success][2 bytes: valLen][value]

func EncodeRequest(reqID uint64, op Operation) []byte {
	keyBytes := []byte(op.Key)
	valBytes := []byte(op.Value)
	payloadLen := 8 + 1 + 2 + len(keyBytes) + 2 + len(valBytes)

	buf := make([]byte, 4+payloadLen)
	binary.BigEndian.PutUint32(buf[0:4], uint32(payloadLen))
	binary.BigEndian.PutUint64(buf[4:12], reqID)
	buf[12] = byte(op.Type)
	binary.BigEndian.PutUint16(buf[13:15], uint16(len(keyBytes)))
	copy(buf[15:15+len(keyBytes)], keyBytes)
	off := 15 + len(keyBytes)
	binary.BigEndian.PutUint16(buf[off:off+2], uint16(len(valBytes)))
	copy(buf[off+2:], valBytes)

	return buf
}

func DecodeRequest(data []byte) (reqID uint64, op Operation, err error) {
	if len(data) < 11 {
		return 0, Operation{}, errors.New("request too short")
	}
	reqID = binary.BigEndian.Uint64(data[0:8])
	op.Type = OpType(data[8])

	off := 9
	if off+2 > len(data) {
		return 0, Operation{}, errors.New("truncated key length")
	}
	keyLen := int(binary.BigEndian.Uint16(data[off : off+2]))
	off += 2
	if off+keyLen > len(data) {
		return 0, Operation{}, errors.New("truncated key")
	}
	op.Key = string(data[off : off+keyLen])
	off += keyLen

	if off+2 > len(data) {
		return 0, Operation{}, errors.New("truncated value length")
	}
	valLen := int(binary.BigEndian.Uint16(data[off : off+2]))
	off += 2
	if off+valLen > len(data) {
		return 0, Operation{}, errors.New("truncated value")
	}
	op.Value = string(data[off : off+valLen])

	return reqID, op, nil
}

func EncodeResponse(reqID uint64, result Result) []byte {
	valBytes := []byte(result.Value)
	payloadLen := 8 + 1 + 2 + len(valBytes)

	buf := make([]byte, 4+payloadLen)
	binary.BigEndian.PutUint32(buf[0:4], uint32(payloadLen))
	binary.BigEndian.PutUint64(buf[4:12], reqID)
	if result.Success {
		buf[12] = 1
	} else {
		buf[12] = 0
	}
	binary.BigEndian.PutUint16(buf[13:15], uint16(len(valBytes)))
	copy(buf[15:], valBytes)

	return buf
}

func DecodeResponse(data []byte) (reqID uint64, result Result, err error) {
	if len(data) < 11 {
		return 0, Result{}, errors.New("response too short")
	}
	reqID = binary.BigEndian.Uint64(data[0:8])
	result.Success = data[8] == 1

	off := 9
	if off+2 > len(data) {
		return 0, Result{}, errors.New("truncated value length")
	}
	valLen := int(binary.BigEndian.Uint16(data[off : off+2]))
	off += 2
	if off+valLen > len(data) {
		return 0, Result{}, errors.New("truncated value")
	}
	result.Value = string(data[off : off+valLen])

	return reqID, result, nil
}

// ReadFrame reads a length-prefixed frame from a reader.
// Returns the payload (after the 4-byte length prefix).
func ReadFrame(r io.Reader) ([]byte, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lenBuf[:])
	if length > 1<<20 { // 1MB max
		return nil, errors.New("frame too large")
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}
	return payload, nil
}
