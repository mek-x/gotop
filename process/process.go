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

func (self *Process) String() string {
	str := "========================="

	str = str + fmt.Sprintf("PID: %d\n", self.Pid)
	str = str + fmt.Sprintf("cmdline: %s\n", self.Cmdline)

	return str
}

func (self *Process) Print() {
	fmt.Println(self)
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
