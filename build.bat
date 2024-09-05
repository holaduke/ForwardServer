@echo off


set GOOS=windows
set GOARCH=386
go build -o ForwardServer.exe -ldflags "-s -w" -trimpath ForwardServer.go 

set GOOS=linux
set GOARCH=386
go build -o ForwardServer.bin -ldflags "-s -w" -trimpath ForwardServer.go 