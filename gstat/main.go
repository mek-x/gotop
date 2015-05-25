package main

import (
	"fmt"
	"github.com/buetow/gstat/diskstats"
	"github.com/buetow/gstat/process"
	"time"
)

type twoProcesses struct {
	flag   bool
	first  process.Process
	second process.Process
}
type processMap map[string]twoProcesses

func gather(timer <-chan bool, d chan<- diskstats.Diskstats, p chan<- process.Process) {
	for {
		switch <-timer {
		case false:
			{
				break
			}
		case true:
			{
				diskstats.Gather(d)
				process.Gather(p)
			}
		}
	}
	close(d)
	close(p)
}

func receiveD(diskstats <-chan diskstats.Diskstats) {
	for d := range diskstats {
		//diskstats.Print()
		// Implemented later
		_ = d
	}
}

func receiveP(processes <-chan process.Process) {
	lastP := make(processMap)
	for p := range processes {
		if val, ok := lastP[p.Id]; ok {
			if val.flag {
				val.second = p
			} else {
				val.first = p
			}
			val.flag = !val.flag
		} else {
			lastP[p.Id] = twoProcesses{flag: true, first: p}
		}
	}
}

func main() {
	timer := make(chan bool)
	diskstats := make(chan diskstats.Diskstats)
	processes := make(chan process.Process)

	go gather(timer, diskstats, processes)

	go receiveD(diskstats)
	go receiveP(processes)

	for counter := 0; counter < 3; counter++ {
		timer <- true
		time.Sleep(time.Second * 2)
	}
	timer <- false

	fmt.Println("Good bye")
}
