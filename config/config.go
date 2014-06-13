package config

import (
	"errors"
)

const (
	CR   = '\r'
	LF   = '\n'
	CRLF = "\r\n"
)

var (
	BadCommandError    = errors.New("bad command")
	WrongArgumentCount = errors.New("wrong argument count")
	WrongCommandKey    = errors.New("wrong command key")
)

// RuleIndex，对于cmdrules的索引位置
const (
	RI_MinCount = iota
	RI_MaxCount // -1 for undefined
)

// 存放指令格式规则，参数范围
var Cmdrules = map[string]map[string][]interface{}{
	"PING":      {"argsRange": []interface{}{1, -1}, "handler": []interface{}{"CmdPingHandler"}},
	"GET":       {"argsRange": []interface{}{2, 2}, "handler": []interface{}{"CmdGetHandler"}},
	"SET":       {"argsRange": []interface{}{3, -1}, "handler": []interface{}{"CmdSetHandler"}},
	"LOOKUP":    {"argsRange": []interface{}{2, -1}, "handler": []interface{}{"CmdLookupHandler"}},
	"CPUPROF_S": {"argsRange": []interface{}{1, -1}, "handler": []interface{}{"PprofHandler"}},
	"CPUPROF_E": {"argsRange": []interface{}{1, -1}, "handler": []interface{}{"PprofHandler"}},
	"MENPROF_G": {"argsRange": []interface{}{1, -1}, "handler": []interface{}{"PprofHandler"}},
}

var MssClusterConfig = map[string]string{
	"test": "['127.0.0.1', '127.0.0.2']",
}

var AppRedisConfig = [...]string{"127.0.0.1:7777", "127.0.0.1:7778", "127.0.0.1:7779"}
