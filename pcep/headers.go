package pcep

import (
	"bytes"
	"encoding/binary"
	"math/bits"
)

// https://tools.ietf.org/html/rfc5440#section-7.2
// 0                   1                   2                   3
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | Object-Class  |   OT  |Res|P|I|   Object Length (bytes)       |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                                                               |
// //                        (Object body)                        //
// |                                                               |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

func newCommonObjHeader(class, objectType uint8, pFlag bool, data []byte) ([]byte, error) {
	// objectType is of 8 bits is going to be reused to hold
	// Res as well as P and I bits
	objectType = bits.RotateLeft8(objectType, 4)
	if pFlag {
		// setting P Flaf at possition 1
		objectType |= (1 << 1)
	}
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, uint16(len(data)+4))
	if err != nil {
		return nil, err
	}
	header := append([]byte{0: class, 1: objectType}, buf.Bytes()...)
	return append(header, data...), nil
}

// 6.1.  Common Header

//      0                   1                   2                   3
//      0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//     | Ver |  Flags  |  Message-Type |       Message-Length          |
//     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

//                 Figure 7: PCEP Message Common Header

//    Ver (Version - 3 bits):  PCEP version number.  Current version is
//       version 1.

//    Flags (5 bits):  No flags are currently defined.  Unassigned bits are
//       considered as reserved.  They MUST be set to zero on transmission
//       and MUST be ignored on receipt.

//    Message-Type (8 bits):  The following message types are currently
//       defined:

//          Value    Meaning
//            1        Open
//            2        Keepalive
//            3        Path Computation Request
//            4        Path Computation Reply
//            5        Notification
//            6        Error
//            7        Close

//    Message-Length (16 bits):  total length of the PCEP message including
//       the common header, expressed in bytes.

// current version is 1 and no flags defined no point
// to implement coresponding parametrs
// https://tools.ietf.org/html/rfc5440#section-6.1
func newCommonHeader(msgType uint8, length uint16) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, length+4)
	if err != nil {
		return nil, err
	}
	return append([]byte{
		// 32 is 00100000 in binary means ver is set to 1, flags to all zeroes
		0: 32,
		1: msgType,
	}, buf.Bytes()...), nil
}
