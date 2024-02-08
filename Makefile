.PHONY: all build run gotool clean help

DISTGO="distgo"
WEB="service"
MASTER="master"
WORKER="worker"
COORDINATOR="coordinator"
#amd64:
#	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ${BINARY}
build:
	go build -ldflags "-s -w" -o ${MASTER} cmd/master/main.go
	go build -ldflags "-s -w" -o ${WORKER} cmd/worker/main.go
	go build -ldflags "-s -w" -o ${COORDINATOR} cmd/coordinator/main.go
run:
	@go run ./main.go

gotool:
	go fmt ./
	go vet ./

clean:
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

help:
	@echo "make - formatting Go code, compiling to binary file"
	@echo "make build - Compiling to binary file"
	@echo "make run - Running Go code"
	@echo "make clean - Removing binary file"
	@echo "make gotool - Running Go tools 'fmt' and 'vet'"