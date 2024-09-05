// +build linux
package utils

import "syscall"


func setFSLimit(limit uint64) uint64 {
	var rLimit syscall.Rlimit
	syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if rLimit.Cur < limit {
		rLimit.Cur = limit
		rLimit.Max = limit
		syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	}
	syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	return rLimit.Cur
}