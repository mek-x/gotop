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

func gather(timerChan <-chan bool, dRxChan chan<- diskstats.Diskstats, pRxChan chan<- process.Process) {
	for {
		switch <-timerChan {
		case false:
			{
				break
			}
		case true:
			{
				diskstats.Gather(dRxChan)
				process.Gather(pRxChan)
			}
		}
	}
	close(dRxChan)
	close(pRxChan)
}

func receiveD(dRxChan <-chan diskstats.Diskstats) {
	for d := range dRxChan {
		//diskstats.Print()
		// Implemented later
		_ = d
	}
}

func compareP() {
	fmt.Println("Comparing")
}

func receiveP(pRxChan <-chan process.Process) {
	lastP := make(processMap)
	for p := range pRxChan {
		if p.Last {
			compareP()
		} else {
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
}

func main() {
	timerChan := make(chan bool)
	dRxChan := make(chan diskstats.Diskstats)
	pRxChan := make(chan process.Process)

	go gather(timerChan, dRxChan, pRxChan)
	go receiveD(dRxChan)
	go receiveP(pRxChan)

	for counter := 0; counter < 3; counter++ {
		timerChan <- true
		time.Sleep(time.Second * 2)
	}
	timerChan <- false

	fmt.Println("Good bye")
}
