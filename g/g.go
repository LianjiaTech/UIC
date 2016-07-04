package g

import (
	"runtime"
)

const (
	VERSION = "0.6"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
