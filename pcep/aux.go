package pcep

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
)

const uintToBoolErr = "Bool value is not 1 or zero"

func uintToBool(i uint) (bool, error) {
	if i == 0 {
		return false, nil
	} else if i == 1 {
		return true, nil
	}
	return false, errors.New(uintToBoolErr)
}

func readBits(by byte, subset ...uint) (r uint) {
	b := uint(by)
	i := uint(0)
	for _, v := range subset {
		if b&(1<<v) > 0 {
			r = r | 1<<uint(i)
		}
		i++
	}
	return
}

// fmt.Printf("Data: %08b \n", data[:4])
func printAsJSON(i interface{}) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(b))
}

func padBytes(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, errors.New("invalid blocksize")
	}
	if len(b) == 0 {
		return nil, errors.New("invalid data ")
	}
	n := blocksize - (len(b) % blocksize)
	pb := make([]byte, len(b)+n)
	copy(pb, b)
	// copy(pb[len(b):], bytes.Repeat([]byte{byte(n)}, n))
	return pb, nil
}

func ipToUnit32(ipStr string) (uint32, error) {
	var res uint32
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0, errors.New("not a valid address")
	}
	ipv4 := ip.To4()
	if ipv4 == nil {
		return 0, errors.New("not a IPv4 address")
	}
	binary.Read(bytes.NewBuffer(ipv4), binary.BigEndian, &res)
	return res, nil
}
