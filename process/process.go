package process

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type Process struct {
	Pid     int
	Cmdline string
	Count   map[string]int
	debug   string
}

func newError() (Process, error) {
	return Process{}, errors.New("Can not read process information")
}

func new(pidstr string) (Process, error) {
	pid, _ := strconv.Atoi(pidstr)
	process := Process{Pid: pid}
	var rawIo string

	err1 := process.gatherRaw(&rawIo, "/proc/%d/io")
	if err1 != nil {
		return process, err1
	}
	err2 := process.parseRawIo(rawIo)
	if err2 != nil {
		return process, err2
	}

	err3 := process.gatherRaw(&process.Cmdline, "/proc/%d/cmdline")
	return process, err3
}

func (self *Process) gatherRaw(what *string, pathf string) error {
	bytes, err := ioutil.ReadFile(fmt.Sprintf(pathf, self.Pid))
	if err != nil {
		return err
	} else {
		*what = string(bytes)
	}
	return nil
}

func (self *Process) parseRawIo(rawIo string) error {
	countMap := make(map[string]int)
	for _, line := range strings.Split(rawIo, "\n") {
		keyval := strings.Split(line, ": ")
		if len(keyval) == 2 {
			count, err := strconv.Atoi(keyval[1])
			if err != nil {
				return err
			}
			countMap[keyval[0]] = count
		}
	}
	self.Count = countMap
	return nil
}

func (self *Process) String() string {
	str := "=========================\n"

	str += fmt.Sprintf("PID: %d\n", self.Pid)
	str += fmt.Sprintf("Cmdline: %s\n", self.Cmdline)
	for key, val := range self.Count {
		str += fmt.Sprintf("%s=%d\n", key, val)
	}
	if self.debug != "" {
		str += fmt.Sprintf("debug: %s\n", self.debug)
	}

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
