install:
	go install github.com/buetow/gotop/gotop
run:
	go run gotop/main.go
docu: install
	sh -c '($(GOPATH)/bin/gotop -h 2>&1)|sed 1d > help.txt;exit 0'
