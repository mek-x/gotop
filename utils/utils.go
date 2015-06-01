package utils

import (
	"fmt"
	"io/ioutil"
)

func Slurp(what *string, path string) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	} else {
		for _, byte := range bytes {
			if byte == 0 {
				*what += " "
			} else {
				*what += string(byte)
			}
		}
	}
	return nil
}

func Abs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

func Human(x int) string {
	units := []string{"", "k", "M", "G", "T", "P", "E", "Z", "Y"}

	f := float32(x)
	i := 0

	for ; f >= 1000; i++ {
		f /= 1000
	}

	return fmt.Sprintf("%d%s", int(f), units[i])
}

func HumanBinary(x int) string {
	units := []string{"", "k", "M", "G", "T", "P", "E", "Z", "Y"}

	f := float32(x)
	i := 0

	for ; f >= 1024; i++ {
		f /= 1024
	}

	return fmt.Sprintf("%d%s", int(f), units[i])
}
