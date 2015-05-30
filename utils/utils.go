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
		*what = string(bytes)
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
	units := []string{"kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}

	f := float32(x)
	i := 0

	for ; f >= 1000; i++ {
		f /= 1000
	}

	return fmt.Sprintf("%d%s", int(f), units[i])
}

func HumanMebi(x int) string {
	units := []string{"ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi", "Yi"}

	f := float32(x)
	i := 0

	for ; f >= 1024; i++ {
		f /= 1024
	}

	return fmt.Sprintf("%d%s", int(f), units[i])
}
