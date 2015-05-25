package utils

import "io/ioutil"

func Slurp(what *string, path string) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	} else {
		*what = string(bytes)
	}
	return nil
}
