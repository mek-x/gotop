package main

import (
	"container/list"
	"fmt"
	"github.com/buetow/gstat/diskstats"
	"github.com/buetow/gstat/process"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var interval time.Duration
var tWidth int
var tHeight int

type twoP struct {
	first  process.Process
	second process.Process
	diff   int
}
type mapP map[string]twoP

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

func printP(lastP *mapP) {
	remove := list.New()
	sorted := list.New()

	for _, val := range *lastP {
		nowTimestamp := int32(time.Now().Unix())
		if val.first.Timestamp+int32(interval)*2 < nowTimestamp {
			// Schedule remove obsolete pids from lastP
			remove.PushBack(val.first.Id)

		} else if val.diff > 0 {
			// Insertion sort
			if sorted.Len() > 0 {
				for e := sorted.Front(); e != nil; e = e.Next() {
					diff := e.Value.(twoP).diff
					if diff >= val.diff {
						//fmt.Printf("Inserting %d before %d\n", val.diff, diff)
						sorted.InsertBefore(val, e)
						break
					}
				}
			} else {
				sorted.PushBack(val)
			}
		}
	}

	fmt.Println("===>")
	for e := sorted.Front(); e != nil; e = e.Next() {
		val := e.Value.(twoP)
		outstr := fmt.Sprintf("%d %s", val.diff, val.first.Id)
		l := len(outstr) - 1
		if l > tWidth {
			l = tWidth
		}
		fmt.Printf("%s\n", outstr[0:l])
	}

	// Rremove obsolete pids from lastP
	for e := remove.Front(); e != nil; e = e.Next() {
		id := e.Value.(twoP).first.Id
		fmt.Println("STALE: " + id)
		delete(*lastP, id)
	}
}

func receiveP(pRxChan <-chan process.Process) {
	lastP := make(mapP)
	flag := false

	makeDiff := func(first, second process.Process) twoP {
		// TODO: make "syscr" choosable/configurable
		firstVal := first.Count["syscr"]
		secondVal := second.Count["syscr"]
		diff := firstVal - secondVal
		if diff < 0 {
			diff = -diff
		}
		return twoP{first, second, diff}
	}

	for p := range pRxChan {
		if p.Last {
			if flag {
				printP(&lastP)
			}
			flag = !flag
		} else {
			if val, ok := lastP[p.Id]; ok {
				if flag {
					lastP[p.Id] = makeDiff(val.first, p)
				} else {
					lastP[p.Id] = makeDiff(p, val.second)
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

	interval = 2
	width, height, err := terminal.GetSize(0)
	if err != nil {
		log.Fatal(err)
	}
	tWidth, tHeight = width, height

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
		time.Sleep(time.Second * interval)
	}
}
