package main

import (
	"fmt"
	"github.com/buetow/gstat/process"
)

const (
	STOP    = 0
	GETPIDS = 1
)

func gstatGather(action <-chan int, res chan<- process.Process) {
	for {
		switch <-action {
		case STOP:
			{
				break
			}
		case GETPIDS:
			{
				process.Gather(res)
			}
		}
		close(res)
	}
}

func main() {
	action := make(chan int)
	res := make(chan process.Process)

	go gstatGather(action, res)

	action <- GETPIDS
	for {
		process, more := <-res
		if more {
			fmt.Println(process.Cmdline)
		} else {
			break
		}
	}
	action <- STOP

	fmt.Println("Good bye")
}
