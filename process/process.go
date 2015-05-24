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
	Pid      int
	Cmdline  string
	rawIo    string
	firstErr error
}

func newError() (Process, error) {
	return Process{}, errors.New("Can not read process information")
}

func new(pidstr string) (Process, error) {
	pid, _ := strconv.Atoi(pidstr)
	process := Process{Pid: pid}

	process.gatherRaw(&process.Cmdline, "/proc/%d/cmdline")
	process.gatherRaw(&process.rawIo, "/proc/%d/io")

	return process, process.firstErr
}

func (self *Process) gatherRaw(what *string, pathf string) {
	bytes, err := ioutil.ReadFile(fmt.Sprintf(pathf, self.Pid))
	if err != nil && self.firstErr == nil {
		self.firstErr = err
	} else {
		*what = string(bytes)
	}
}

func (self *Process) String() string {
	str := "=========================\n"

	str = str + fmt.Sprintf("PID: %d\n", self.Pid)
	str = str + fmt.Sprintf("Cmdline: %s\n", self.Cmdline)
	str = str + fmt.Sprintf("rawIo: %s\n", self.rawIo)

	return str
}

func (self *Process) Print() {
	fmt.Println(self)
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
