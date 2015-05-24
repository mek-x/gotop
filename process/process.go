package process

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
)

type Process struct {
	Pid     int
	Cmdline string
}

func (self *Process) Print() {
	fmt.Printf("=====================\n")
	fmt.Printf("PID: %d\n", self.Pid)
	fmt.Printf("cmdline: %s\n", self.Cmdline)
}

func Gather(res chan<- Process) {
	re, _ := regexp.Compile("^[0-9]+$")

	dir, err := ioutil.ReadDir("/proc/")
	if err != nil {
		log.Fatal(err)
	}

	for _, direntry := range dir {
		name := direntry.Name()
		if re.MatchString(name) {
			pid, _ := strconv.Atoi(name)
			bytes, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
			if err != nil {
				log.Fatal(err)
			}
			if len(bytes) > 0 {
				res <- Process{Pid: pid, Cmdline: string(bytes)}
			}
		}
	}
}
