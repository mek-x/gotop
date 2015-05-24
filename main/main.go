package main

import (
	"fmt"
	"github.com/buetow/gstat/process"
	"time"
)

func gather(timer <-chan bool, processes chan<- process.Process) {
	for {
		switch <-timer {
		case false:
			{
				break
			}
		case true:
			{
				process.Gather(processes)
			}
		}
	}
	close(processes)
}

func receive(processes <-chan process.Process) {
	for process := range processes {
		fmt.Println(process)
	}
}

func main() {
	timer := make(chan bool)
	res := make(chan process.Process)

	go gather(timer, res)
	go receive(res)

	for counter := 0; counter < 3; counter++ {
		timer <- true
		time.Sleep(time.Second)
	}
	timer <- false

	fmt.Println("Good bye")
}
