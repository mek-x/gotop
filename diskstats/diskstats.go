package diskstats

import (
	"fmt"
	"io/ioutil"
)

type Diskstats struct {
	debug string
}

func new() (Diskstats, error) {
	var raw string
	d := Diskstats{}

	if err := d.gatherRaw(&raw, "/proc/diskstats"); err != nil {
		return d, err
	}

	return d, nil
}

func (self *Diskstats) gatherRaw(what *string, path string) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	} else {
		*what = string(bytes)
	}
	return nil
}

func (self *Diskstats) String() string {
	str := "DISKSTATS=========================\n"

	if self.debug != "" {
		str += fmt.Sprintf("debug: %s\n", self.debug)
	}

	return str
}

func (self *Diskstats) Print() {
	fmt.Println(self)
}

func Gather(diskstats chan<- Diskstats) {
	if d, err := new(); err == nil {
		diskstats <- d
	}
}
