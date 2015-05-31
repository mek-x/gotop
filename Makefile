install:
	go install github.com/buetow/gstat/gstat
run:
	go run gstat/main.go
docu: install
	sh -c '($(GOPATH)/bin/gstat -h 2>&1)|sed 1d > help.txt;exit 0'
