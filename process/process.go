package process

import (
	"errors"
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

func new(pidstr string) (Process, error) {
	pid, _ := strconv.Atoi(pidstr)
	bytes, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		log.Fatal(err)
	}
	if len(bytes) > 0 {
		return Process{Pid: pid, Cmdline: string(bytes)}, nil
	} else {
		return Process{}, errors.New("Can not read process information")
	}
}

func (self *Process) String() string {
	str := "========================="

	str = str + fmt.Sprintf("PID: %d\n", self.Pid)
	str = str + fmt.Sprintf("cmdline: %s\n", self.Cmdline)

	return str
}

func Gather(processes chan<- Process) {
	re, _ := regexp.Compile("^[0-9]+$")

	dir, err := ioutil.ReadDir("/proc/")
	if err != nil {
		log.Fatal(err)
	}

	for _, direntry := range dir {
		name := direntry.Name()
		if re.MatchString(name) {
			process, err := new(name)
			if err == nil {
				processes <- process
			}
		}
	}
}
