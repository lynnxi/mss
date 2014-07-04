package main

import (
	"bytes"
	"mss/config"
	"mss/lib/utils"
	"strings"
)

// Command表示一个客户端指令
type Command struct {
	args [][]byte
}

func NewCommand(args ...[]byte) (cmd *Command) {
	cmd = &Command{
		args: args,
	}
	return
}

func (cmd *Command) Len() int {
	return len(cmd.args)
}

// Redis协议的Command数据
/*
*<number of arguments> CR LF
$<number of bytes of argument 1> CR LF
<argument data> CR LF
...
$<number of bytes of argument N> CR LF
<argument data> CR LF
*/
func (cmd *Command) Bytes() []byte {
	buf := bytes.Buffer{}
	buf.WriteByte('*')
	argCount := cmd.Len()
	buf.WriteString(utils.Itoa(argCount)) //<number of arguments>
	buf.WriteString(config.CRLF)
	for i := 0; i < argCount; i++ {
		buf.WriteByte('$')
		argSize := len(cmd.args[i])
		buf.WriteString(utils.Itoa(argSize)) //<number of bytes of argument i>
		buf.WriteString(config.CRLF)
		buf.Write(cmd.args[i]) //<argument data>
		buf.WriteString(config.CRLF)
	}

	return buf.Bytes()
}

func (cmd *Command) String() string {
	buf := bytes.Buffer{}
	for i, count := 0, cmd.Len(); i < count; i++ {
		if i > 0 {
			buf.WriteString(" ")
		}
		buf.Write(cmd.args[i])
	}
	return buf.String()
}

func (cmd *Command) Key() []byte {
	if len(cmd.args) > 1 {
		return bytes.ToUpper(cmd.args[1])
	} else {
		return bytes.ToUpper(cmd.args[0])
	}
}

func (cmd *Command) Keyo() []byte {
	if len(cmd.args) > 1 {
		return cmd.args[1]
	} else {
		return cmd.args[0]
	}
}

// 大写的指令名称
func (cmd *Command) Name() string {
	return string(bytes.ToUpper(cmd.args[0]))
}

// 原始数据
func (cmd *Command) Args() [][]byte {
	return cmd.args
}

func (cmd *Command) StringAtIndex(i int) string {
	if i >= cmd.Len() {
		return ""
	}
	return string(cmd.args[i])
}

// 验证指令参数数量、非法字符等
func (cmd *Command) verifyCommand() error {
	if cmd == nil || cmd.Len() == 0 {
		return config.BadCommandError
	}

	name := cmd.Name()
	rule, exist := config.Cmdrules[name]["argsRange"]
	if !exist {
		return config.BadCommandError
	}

	for i, count := 0, len(rule); i < count; i++ {
		switch i {
		case config.RI_MinCount:
			if val := rule[i].(int); val != -1 && cmd.Len() < val {
				return config.WrongArgumentCount
			}
		case config.RI_MaxCount:
			if val := rule[i].(int); val != -1 && cmd.Len() > val {
				return config.WrongArgumentCount
			}
		}
	}

	// 拒绝使用内部关键字 #[]
	if cmd.Len() > 1 {
		key := cmd.StringAtIndex(1)
		if strings.ContainsAny(key, "#[] ") {
			return config.WrongCommandKey
		}
	}
	return nil
}
