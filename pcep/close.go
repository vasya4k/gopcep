package pcep

import (
	"bytes"
	"encoding/binary"
)

// https://tools.ietf.org/html/rfc5440#section-7.17
func parseClose(data []byte) uint8 {
	return uint8(data[3])
}

// https://tools.ietf.org/html/rfc5440#section-7.17
func newCloseObj(reason uint8) ([]byte, error) {
	var (
		reserved uint16
		flags    uint8
	)
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.BigEndian, reserved)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, flags)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, reason)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func newCloseMsg(reason uint8) ([]byte, error) {
	closeObj, err := newCloseObj(reason)
	if err != nil {
		return nil, err
	}

	ch, err := newCommonHeader(7, uint16(len(closeObj)))
	if err != nil {
		return nil, err
	}
	return append(ch, closeObj...), nil
}
