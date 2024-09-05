package utils


func SetFSLimit(maxOpenFiles uint64 ) uint64 {
	return setFSLimit(maxOpenFiles)
}