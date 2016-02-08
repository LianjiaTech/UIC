package g

import (
	"runtime"
)

const (
	VERSION = "0.0.5"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
