package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
)

const (
	STOP    = 0
	GETPIDS = 1
)

type Program struct {
	pid     int
	command string
}

func gstatGatherPids(res chan<- string) {
	re, _ := regexp.Compile("^[0-9]+$")

	dir, err := ioutil.ReadDir("/proc/")
	if err != nil {
		log.Fatal(err)
	}

	for _, fi := range dir {
		name := fi.Name()
		if re.MatchString(name) {
			res <- name
		}
	}
}

func gstatGather(action <-chan int, done chan<- bool, res chan<- string) {
	for {
		switch <-action {
		case STOP:
			{
				done <- true
				break
			}
		case GETPIDS:
			{
				gstatGatherPids(res)
				done <- true
			}
		}
	}
}

func main() {
	action := make(chan int)
	done := make(chan bool)
	res := make(chan string)

	go gstatGather(action, done, res)

	action <- GETPIDS
	for {
		switch <-done {
		case false:
			{
				fmt.Println(<-res)
			}
		case true:
			{
				break
			}
		}
	}
}
