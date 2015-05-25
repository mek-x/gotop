package main

import (
	"fmt"
	"github.com/buetow/gstat/diskstats"
	"github.com/buetow/gstat/process"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type twoP struct {
	first  process.Process
	second process.Process
}
type processMap map[string]twoP

func timedGather(timerChan <-chan bool, dRxChan chan<- diskstats.Diskstats, pRxChan chan<- process.Process) {
	for {
		switch <-timerChan {
		case false:
			{
				break
			}
		case true:
			{
				go diskstats.Gather(dRxChan)
				go process.Gather(pRxChan)
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

func compareP(lastP processMap) {
	for id, val := range lastP {
		first := val.first.Count["syscr"]
		second := val.second.Count["syscr"]
		diff := first - second
		if diff < 0 {
			diff = -diff
		}
		fmt.Printf("%d %s\n", diff, id)
	}
}

func receiveP(pRxChan <-chan process.Process) {
	lastP := make(processMap)
	flag := false

	for p := range pRxChan {
		if p.Last {
			if flag {
				compareP(lastP)
			}
			flag = !flag
		} else {
			if val, ok := lastP[p.Id]; ok {
				if flag {
					lastP[p.Id] = twoP{first: val.first, second: p}
				} else {
					lastP[p.Id] = twoP{first: p, second: val.second}
				}
			} else {
				lastP[p.Id] = twoP{first: p}
			}
		}
	}
}

func main() {
	timerChan := make(chan bool)
	dRxChan := make(chan diskstats.Diskstats)
	pRxChan := make(chan process.Process)

	go timedGather(timerChan, dRxChan, pRxChan)
	go receiveD(dRxChan)
	go receiveP(pRxChan)

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt)
	signal.Notify(termChan, syscall.SIGTERM)

	go func() {
		<-termChan
		timerChan <- false
		fmt.Println("Good bye")
		os.Exit(1)
	}()

	for {
		timerChan <- true
		time.Sleep(time.Second * 2)
	}
}
