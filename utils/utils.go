package utils

import "io/ioutil"

func GatherRaw(what *string, path string) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	} else {
		*what = string(bytes)
	}
	return nil
}
