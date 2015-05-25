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

func gather(timer <-chan bool, dChan chan<- diskstats.Diskstats, pChan chan<- process.Process) {
	for {
		switch <-timer {
		case false:
			{
				break
			}
		case true:
			{
				diskstats.Gather(dChan)
				process.Gather(pChan)
			}
		}
	}
	close(dChan)
	close(pChan)
}

func receiveD(dChan <-chan diskstats.Diskstats) {
	for d := range dChan {
		//diskstats.Print()
		// Implemented later
		_ = d
	}
}

func receiveP(pChan <-chan process.Process) {
	lastP := make(processMap)
	for p := range pChan {
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
	dChan := make(chan diskstats.Diskstats)
	pChan := make(chan process.Process)

	go gather(timer, dChan, pChan)

	go receiveD(dChan)
	go receiveP(pChan)

	for counter := 0; counter < 3; counter++ {
		timer <- true
		time.Sleep(time.Second * 2)
	}
	timer <- false

	fmt.Println("Good bye")
}
