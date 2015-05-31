// gstat (C) 2015 Paul Buetow (gstat@dev.buetow.org)

package main

import (
	"container/list"
	"errors"
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
	modeName string
}

type twoP struct {
	first, second      process.Process
	diff, diffR, diffW int
	exited             bool
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
		if val.first.Timestamp+2 < nowTimestamp {
			// Schedule remove obsolete pids from lastP
			remove.PushBack(val.first.Id)
			// Display this process one more time, but in a fancy way
			val.exited = true
		}

		if val.diff > 0 {
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
		id := e.Value.(string)
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
	fmt.Printf("Mode: %s, Interval: %s\n", config.modeName, config.interval)
	fmt.Printf("(Hit Ctr-C to quit, re-run with -h for flags)\n")
	fmt.Printf("%5s/%5s %5s %s\n", "WRITE", "READS", "PID", "COMMAND")

	// Print the results
	row := 3
	for e := sortedP.Front(); e != nil; e = e.Next() {
		row++
		if row > tHeight {
			break
		}
		val := e.Value.(twoP)
		first := val.first

		var outstr string

		if val.exited {
			outstr = fmt.Sprintf("XXXXXXXXXXX %5d %s", first.Pid, first.Cmdline)

		} else {
			var humanW, humanR string

			if *config.binary {
				humanW, humanR = utils.HumanBinary(val.diffW), utils.HumanBinary(val.diffR)
			} else {
				humanW, humanR = utils.Human(val.diffW), utils.Human(val.diffR)
			}

			outstr = fmt.Sprintf("%5s %5s %5d %s", humanW, humanR, first.Pid, first.Cmdline)
		}

		l := len(outstr)
		if l > tWidth {
			l = tWidth
		}
		fmt.Printf("%s\n", outstr[0:l])
	}
}

func modeNames() (string, string, string, error) {
	switch *config.mode {
	case 0:
		return "read_bytes", "write_bytes", "bytes", nil
	case 1:
		return "syscr", "syscw", "syscalls", nil
	case 2:
		return "rchar", "wchar", "chars", nil
	}

	errstr := fmt.Sprintf("No such mode: %d\n", *config.mode)
	return "", "", "", errors.New(errstr)
}

func receiveP(pRxChan <-chan process.Process) {
	lastP := make(mapP)
	flag := false

	readKey, writeKey, modeName, err := modeNames()
	if err != nil {
		log.Fatal(err)
	}
	config.modeName = modeName

	makeDiff := func(first, second process.Process) twoP {
		firstValR, firstValW := first.Count[readKey], first.Count[writeKey]
		secondValR, secondValW := second.Count[readKey], second.Count[writeKey]
		diffR, diffW := utils.Abs(firstValR-secondValR), utils.Abs(firstValW-secondValW)
		diff := diffR + diffW
		return twoP{first, second, diff, diffR, diffW, false}
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
	config.mode = flag.Int("m", 1, "The stats mode: 0:bytes 1:syscalls 2:chars")

	flag.Parse()

	config.banner = "gstat v0.1 (C) 2015 Paul buetow <http://gstat.buetow.org>"

	if *helpF {
		fmt.Println(config.banner)
		os.Exit(0)
	}

	config.interval = time.Duration(*interF) * time.Second
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
		os.Exit(0)
	}()

	for {
		tChan <- true
		time.Sleep(config.interval)
	}
}
