// gotop (C) 2015 Paul Buetow (gotop@dev.buetow.org)

package main

import (
	"container/list"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
	"syscall"
	"time"
	"golang.org/x/crypto/ssh/terminal"
	"github.com/mek-x/gotop/diskstats"
	"github.com/mek-x/gotop/process"
	"github.com/mek-x/gotop/utils"
)

var config struct {
	banner   string
	interval time.Duration
	binary   *bool
	mode     *int
	leaveStale *bool
	modeName string
}

type twoP struct {
	first, second      process.Process
	diff, diffR, diffW int
	accu, accuR, accuW int
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
		seconds := int32(config.interval.Seconds())
		if val.first.Timestamp+seconds < nowTimestamp {
			// Schedule remove obsolete pids from lastP
			if !(*config.leaveStale) {
				remove.PushBack(val.first.Id)
			}
			// Display this process one more time, but in a fancy way
			val.exited = true
		}

		if val.accu > 0 {
			// Insertion sort
			if sorted.Len() > 0 {
				inserted := false
				for e := sorted.Front(); e != nil; e = e.Next() {
					accu := e.Value.(twoP).accu
					if val.accu >= accu {
						//fmt.Printf("Inserting %d before %d\n", val.first.Pid, e.Value.(twoP).first.Pid)
						sorted.InsertBefore(val, e)
						inserted = true
						break
					}
				}
				if !inserted {
					sorted.PushBack(val)
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
	fmt.Printf("%5s %5s %5s %5s %5s %s\n", "WRITE", "READS", "W/s", "R/s", "PID", "COMMAND")

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
		var cmdstr string
		var W, R, Ws, Rs string

		if *config.binary {
			W, R = utils.HumanBinary(val.accuW), utils.HumanBinary(val.accuR)
			Ws, Rs = utils.HumanBinary(val.diffW), utils.HumanBinary(val.diffR)
		} else {
			W, R = utils.Human(val.accuW), utils.Human(val.accuR)
			Ws, Rs = utils.Human(val.diffW), utils.Human(val.diffR)
		}

		if val.exited {
			cmdstr = "[exited]" + first.Cmdline
		} else {
			cmdstr = first.Cmdline
		}

		outstr = fmt.Sprintf("%5s %5s %5s %5s %5d %s",
			W, R,
			Ws, Rs,
			first.Pid,
			cmdstr)

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

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func receiveP(pRxChan <-chan process.Process) {
	lastP := make(mapP)
	flag := false

	readKey, writeKey, modeName, err := modeNames()
	if err != nil {
		log.Fatal(err)
	}
	config.modeName = modeName
	seconds := float64(config.interval.Milliseconds()) / 1000.0

	makeDiff := func(first, second process.Process) twoP {
		firstValR, firstValW := first.Count[readKey], first.Count[writeKey]
		secondValR, secondValW := second.Count[readKey], second.Count[writeKey]
		diffR, diffW := utils.Abs(firstValR-secondValR), utils.Abs(firstValW-secondValW)

		accuR := max(first.Count[readKey], second.Count[readKey])
		accuW := max(first.Count[writeKey], second.Count[writeKey])
		accu := accuR + accuW

		// Calculate averages, so we have always per second valus
		diff := int(float64(diffR+diffW) / seconds)
		diffR = int(float64(diffR) / seconds)
		diffW = int(float64(diffW) / seconds)

		return twoP{first, second, diff, diffR, diffW, accu, accuR, accuW, false}
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
				lastP[p.Id] = twoP{
					first: p,
					accu: p.Count[readKey] + p.Count[writeKey],
					accuW: p.Count[writeKey],
					accuR: p.Count[readKey]}
			}
		}
	}
}

func parseFlags() {
	helpF := flag.Bool("v", false, "Print the version")
	interF := flag.Float64("i", 1.0, "Update interval in seconds")

	config.binary = flag.Bool("b", false, "Use binary instead of decimal (e.g. kiB an not kB)")
	config.mode = flag.Int("m", 0, "The stats mode: 0:bytes 1:syscalls 2:chars")
	config.leaveStale = flag.Bool("l", false, "Leave stale processes on the list")

	flag.Parse()

	config.banner = "gotop v0.1 (C) 2015 Paul Buetow <https://gotop.buetow.org>"

	if *helpF {
		fmt.Println(config.banner)
		os.Exit(0)
	}

	config.interval = time.Duration(*interF * 1000) * time.Millisecond
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

	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	if user.Username != "root" {
		fmt.Println("Warning: You are not root, so you will not see everything!")
		time.Sleep(time.Duration(2) * time.Second)
	}

	for {
		tChan <- true
		time.Sleep(config.interval)
	}
}
