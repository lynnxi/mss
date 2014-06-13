package main

import (
	"fmt"
	//"log"
	"mss/lib/stdlog"
	//"net/http"
	//_ "net/http/pprof"
	//"flag"
	//	"os"
	"runtime"
	//	"runtime/pprof"
	"time"
)

func main() {
	// （可选）设置函数前缀
	stdlog.SetPrefix(func() string {
		t := time.Now()
		return fmt.Sprintf("[%d-%02d-%02d %02d:%02d:%02d] ", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	})

	runtime.GOMAXPROCS(4)

	stdlog.Println("start...")
	server := newServer()
	server.Listen("127.0.0.1:6320")
}
