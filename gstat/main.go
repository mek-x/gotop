package main

import (
	"container/list"
	"fmt"
	"github.com/buetow/gstat/diskstats"
	"github.com/buetow/gstat/process"
	"github.com/buetow/gstat/utils"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var interval_ time.Duration
var banner_ string

type twoP struct {
	first  process.Process
	second process.Process
	diff   int
	diffR  int
	diffW  int
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

func sortP(lastP *mapP) *list.List {
	remove := list.New()
	sorted := list.New()

	for _, val := range *lastP {
		nowTimestamp := int32(time.Now().Unix())
		if val.first.Timestamp+int32(interval_)*2 < nowTimestamp {
			// Schedule remove obsolete pids from lastP
			remove.PushBack(val.first.Id)

		} else if val.diff > 0 {
			// Insertion sort
			if sorted.Len() > 0 {
				for e := sorted.Front(); e != nil; e = e.Next() {
					diff := e.Value.(twoP).diff
					if diff < val.diff {
						//fmt.Printf("Inserting %d before %d\n", val.diff, diff)
						sorted.InsertBefore(val, e)
						break
					}
				}
			} else {
				sorted.PushFront(val)
			}
		}
	}

	// Rremove obsolete pids from lastP
	for e := remove.Front(); e != nil; e = e.Next() {
		id := e.Value.(twoP).first.Id
		//fmt.Println("Removing stale process: " + id)
		delete(*lastP, id)
	}

	return sorted
}

func printP(sortedP *list.List) {

	tWidth, tHeight, err := terminal.GetSize(0)
	if err != nil {
		log.Fatal(err)
	}

	// Clear the screen + print header
	fmt.Println("\033[H\033[2J")
	fmt.Printf("%6s %6s %6s %s\n", "WRITES", "READS", "PID", "COMMAND")

	// Print the results
	row := 2
	for e := sortedP.Front(); e != nil; e = e.Next() {
		row++
		if row > tHeight {
			break
		}
		val := e.Value.(twoP)
		first := val.first
		outstr := fmt.Sprintf("%6s %6s %6d %s", utils.Human(val.diffW), utils.Human(val.diffR), first.Pid, first.Cmdline)

		l := len(outstr)
		if l > tWidth {
			l = tWidth
		}
		fmt.Printf("%s\n", outstr[0:l])
	}

	// Fill up the other rows + print banner
	for ; row < tHeight-1; row++ {
		fmt.Println()
	}
	fmt.Print(banner_)
}

func receiveP(pRxChan <-chan process.Process) {
	lastP := make(mapP)
	flag := false

	makeDiff := func(first, second process.Process) twoP {
		// TODO: make "rchar,wchar" configurable
		firstValR, firstValW := first.Count["rchar"], first.Count["wchar"]
		secondValR, secondValW := second.Count["rchar"], second.Count["wchar"]
		diffR, diffW := utils.Abs(firstValR-secondValR), utils.Abs(firstValW-secondValW)
		diff := diffR + diffW
		return twoP{first, second, diff, diffR, diffW}
	}

	for p := range pRxChan {
		if p.Last {
			if flag {
				printP(sortP(&lastP))
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

	interval_ = 2
	banner_ = "gstat 0.1"

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
		time.Sleep(time.Second * interval_)
	}
}
