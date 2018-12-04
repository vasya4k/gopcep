package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
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

func bits(by byte, subset ...uint) (r uint) {
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
