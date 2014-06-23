package main

import (
	"mss/config"
	"mss/lib/stdlog"
	"os"
	"runtime"
	"runtime/pprof"
)

var DescHandler = map[string]func(s *Server, cmd *Command) (r *Reply, err error){
	"CmdPingHandler":   CmdPingHandler,
	"CmdGetHandler":    CmdGetHandler,
	"CmdSetHandler":    CmdSetHandler,
	"CmdLookupHandler": CmdLookupHandler,
	"PprofHandler":     PprofHandler,
}

var cpuProfile *os.File

func PprofHandler(s *Server, cmd *Command) (r *Reply, err error) {
	switch cmd.Name() {
	case "lookup heap":
		p := pprof.Lookup("heap")
		p.WriteTo(os.Stdout, 2)
	case "lookup threadcreate":
		p := pprof.Lookup("threadcreate")
		p.WriteTo(os.Stdout, 2)
	case "lookup block":
		p := pprof.Lookup("block")
		p.WriteTo(os.Stdout, 2)
	case "CPUPROF_S":
		if cpuProfile == nil {
			if f, err := os.Create("mss.cpuprof"); err != nil {
				stdlog.Println("start cpu profile failed: %v", err)
			} else {
				stdlog.Println("start cpu profile")
				pprof.StartCPUProfile(f)
				cpuProfile = f
			}
		}
	case "CPUPROF_E":
		if cpuProfile != nil {
			pprof.StopCPUProfile()
			cpuProfile.Close()
			cpuProfile = nil
			stdlog.Println("stop cpu profile")
		}
	case "MENPROF_G":
		if f, err := os.Create("mss.memprof"); err != nil {
			stdlog.Println("record memory profile failed: %v", err)
		} else {
			runtime.GC()
			pprof.WriteHeapProfile(f)
			f.Close()
			stdlog.Println("record memory profile")
		}
	}
	r = StatusReply("OK")
	return
}

func CmdPingHandler(s *Server, cmd *Command) (r *Reply, err error) {
	r, err = s.GetProxy(cmd).Dispatch(cmd)
	r.Type = ReplyTypeStatus
	//r = StatusReply("PONG")
	return

}

func CmdGetHandler(s *Server, cmd *Command) (r *Reply, err error) {
	r, err = s.GetProxy(cmd).Dispatch(cmd)
	r.Type = ReplyTypeBulk
	return
}

func CmdSetHandler(s *Server, cmd *Command) (r *Reply, err error) {
	r, err = s.GetProxy(cmd).Dispatch(cmd)
	r.Type = ReplyTypeStatus
	return
}

func CmdLookupHandler(s *Server, cmd *Command) (r *Reply, err error) {
	appid := cmd.StringAtIndex(1)
	clusterConfig, exist := config.MssClusterConfig[appid]
	if !exist {
		r = ErrorReply(config.BadCommandError)
		return
	}

	r = BulkReply(clusterConfig)

	return
}
