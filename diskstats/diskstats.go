package diskstats

import (
	"fmt"
	"github.com/buetow/gstat/utils"
)

type Diskstats struct {
	debug string
}

func new() (Diskstats, error) {
	var raw string
	d := Diskstats{}

	if err := utils.Slurp(&raw, "/proc/diskstats"); err != nil {
		return d, err
	}

	return d, nil
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
