package main

import (
	"fmt"
	"github.com/buetow/gstat/diskstats"
	"github.com/buetow/gstat/process"
	"time"
)

type processMap map[string]process.Process

var lastP processMap

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

func receive1(diskstats <-chan diskstats.Diskstats) {
	for diskstats := range diskstats {
		diskstats.Print()
	}
}

func receive2(processes <-chan process.Process) {
	lastP = make(processMap)
	for process := range processes {
		lastP[process.Id] = process
		process.Print()
	}
}

func main() {
	timer := make(chan bool)
	diskstats := make(chan diskstats.Diskstats)
	processes := make(chan process.Process)

	go gather(timer, diskstats, processes)
	go receive1(diskstats)
	go receive2(processes)

	for counter := 0; counter < 3; counter++ {
		timer <- true
		time.Sleep(time.Second * 2)

		fmt.Printf("Next... %d\n", counter)
	}
	timer <- false

	fmt.Println("Good bye")
}
