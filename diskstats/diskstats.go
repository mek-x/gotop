package diskstats

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type Diskstats struct {
	Pid     int
	Cmdline string
	Count   map[string]int
	debug   string
}

func new(pidstr string) (Diskstats, error) {
	pid, err0 := strconv.Atoi(pidstr)
	if err0 != nil {
		return Diskstats{}, err0
	}
	diskstats := Diskstats{Pid: pid}
	var rawIo string

	err1 := diskstats.gatherRaw(&rawIo, "/proc/%d/io")
	if err1 != nil {
		return diskstats, err1
	}
	err2 := diskstats.parseRawIo(rawIo)
	if err2 != nil {
		return diskstats, err2
	}

	err3 := diskstats.gatherRaw(&diskstats.Cmdline, "/proc/%d/cmdline")
	return diskstats, err3
}

func (self *Diskstats) gatherRaw(what *string, pathf string) error {
	bytes, err := ioutil.ReadFile(fmt.Sprintf(pathf, self.Pid))
	if err != nil {
		return err
	} else {
		*what = string(bytes)
	}
	return nil
}

func (self *Diskstats) parseRawIo(rawIo string) error {
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

func (self *Diskstats) String() string {
	str := "DISKSTATS=========================\n"

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

func (self *Diskstats) Print() {
	fmt.Println(self)
}

func Gather(diskstatses chan<- Diskstats) {
	re, _ := regexp.Compile("^[0-9]+$")

	dir, err := ioutil.ReadDir("/proc/")
	if err != nil {
		log.Fatal(err)
	}

	for _, direntry := range dir {
		name := direntry.Name()
		if re.MatchString(name) {
			diskstats, err := new(name)
			if err == nil {
				diskstatses <- diskstats
			}
		}
	}
}
