// gstat (C) 2015 Paul Buetow (gstat@dev.buetow.org)

package main

import (
	"container/list"
	"flag"
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

var config struct {
	banner   string
	interval time.Duration
	binary   *bool
	mode     *int
}

type twoP struct {
	first, second      process.Process
	diff, diffR, diffW int
}
type mapP map[string]twoP

func timedGather(tChan <-chan bool, dRxChan chan<- diskstats.Diskstats, pRxChan chan<- process.Process) {
	for {
		switch <-tChan {
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
		if val.first.Timestamp+int32(config.interval)*2 < nowTimestamp {
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
	fmt.Printf("\033[H\033[2J")
	fmt.Printf("(Hit Ctr-C to quit, re-run with -h for flags)\n")
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

		var humanW, humanR string
		if *config.binary {
			humanW, humanR = utils.HumanBinary(val.diffW), utils.HumanBinary(val.diffR)
		} else {
			humanW, humanR = utils.Human(val.diffW), utils.Human(val.diffR)
		}

		outstr := fmt.Sprintf("%6s %6s %6d %s", humanW, humanR, first.Pid, first.Cmdline)

		l := len(outstr)
		if l > tWidth {
			l = tWidth
		}
		fmt.Printf("%s\n", outstr[0:l])
	}
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

func parseFlags() {
	helpF := flag.Bool("v", false, "Print the version")
	interF := flag.Int("i", 2, "Update interval in seconds")

	config.binary = flag.Bool("b", false, "Use binary instead of deciman (e.g. kiB an not kB")
	config.mode = flag.Int("m", 0, "The stats mode: 0:bytes 1:syscalls 2:chars")

	flag.Parse()

	config.banner = "gstat v0.1 (C) 2015 Paul buetow <http://gstat.buetow.org>"

	if *helpF {
		fmt.Println(config.banner)
		os.Exit(0)
	}

	config.interval = time.Duration(*interF)
}

func main() {
	parseFlags()

	tChan := make(chan bool)
	dRxChan := make(chan diskstats.Diskstats)
	pRxChan := make(chan process.Process)

	go timedGather(tChan, dRxChan, pRxChan)
	go receiveD(dRxChan)
	go receiveP(pRxChan)

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt)
	signal.Notify(termChan, syscall.SIGTERM)

	go func() {
		<-termChan
		tChan <- false
		fmt.Println("Good bye! This was:")
		fmt.Println(config.banner)
		os.Exit(1)
	}()

	for {
		tChan <- true
		time.Sleep(time.Second * config.interval)
	}
}
