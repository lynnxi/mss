package main

import (
	"bufio"
	"bytes"
	"errors"
	//"fmt"
	"mss/config"
	"mss/lib/mio"
	"mss/lib/stdlog"
	"net"
	"strconv"
)

type Connection struct {
	net.Conn
	reader *mio.Reader
	writer *bufio.Writer
}

func NewConnection(conn net.Conn) (c *Connection) {
	//这里net.conn为什么不是引用？
	c = &Connection{
		Conn: conn,
	}
	c.reader = mio.NewReader(c.Conn)
	c.writer = bufio.NewWriter(c.Conn)

	return
}

// 从客户端连接获取指令
// (下面读取过程，线上应用前需要增加错误校验，数据大小限制)
/*
*<number of arguments> CR LF
$<number of bytes of argument 1> CR LF
<argument data> CR LF
...
$<number of bytes of argument N> CR LF
<argument data> CR LF
*/
func (c *Connection) ReadCommand() (cmd *Command, err error) {
	// Read ( *<number of arguments> CR LF )
	err = c.reader.SkipByte('*')
	if err != nil { // io.EOF
		return
	}
	// number of arguments
	var argCount int
	if argCount, err = c.reader.ReadInt(); err != nil {
		return
	}
	args := make([][]byte, argCount)
	for i := 0; i < argCount; i++ {
		// Read ( $<number of bytes of argument 1> CR LF )
		err = c.reader.SkipByte('$')
		if err != nil {
			return
		}

		var argSize int
		argSize, err = c.reader.ReadInt()
		if err != nil {
			return
		}

		// Read ( <argument data> CR LF )
		args[i] = make([]byte, argSize)
		_, err = c.reader.ReadFull(args[i])
		if err != nil {
			return
		}

		err = c.reader.SkipBytes([]byte{config.CR, config.LF})
		if err != nil {
			return
		}
	}
	cmd = NewCommand(args...)
	return
}

func (c *Connection) WriteCommand(cmd *Command) (err error) {
	stdlog.Println("write cmd " + string(cmd.String()))
	_, err = c.Write(cmd.Bytes())
	return
}

// 从连接里读取回复
/*
In a Status Reply the first byte of the reply is "+"
In an Error Reply the first byte of the reply is "-"
In an Integer Reply the first byte of the reply is ":"
In a Bulk Reply the first byte of the reply is "$"
In a Multi Bulk Reply the first byte of the reply s "*"
*/
func (c *Connection) ReadReply() (reply *Reply, err error) {
	var b byte
	if b, err = c.reader.ReadByte(); err != nil {
		return
	}

	reply = &Reply{}
	switch b {
	case '+':
		reply.Type = ReplyTypeStatus
		reply.Value, err = c.reader.ReadString()
	case '-':
		reply.Type = ReplyTypeError
		reply.Value, err = c.reader.ReadString()
	case ':':
		reply.Type = ReplyTypeInteger
		reply.Value, err = c.reader.ReadInt()
	case '$':
		reply.Type = ReplyTypeBulk
		var bufsize int
		bufsize, err = c.reader.ReadInt()
		if err != nil {
			break
		}
		buf := make([]byte, bufsize)
		_, err = c.reader.ReadFull(buf)
		if err != nil {
			break
		}
		reply.Value = buf
		c.reader.SkipBytes([]byte{config.CR, config.LF})
	case '*':
		reply.Type = ReplyTypeMultiBulks
		var argCount int
		argCount, err = c.reader.ReadInt()
		if err != nil {
			break
		}
		if argCount == -1 {
			reply.Value = nil // *-1
		} else {
			args := make([]interface{}, argCount)
			for i := 0; i < argCount; i++ {
				// TODO multi bulk 的类型 $和:
				err = c.reader.SkipByte('$')
				if err != nil {
					break
				}
				var argSize int
				argSize, err = c.reader.ReadInt()
				if err != nil {
					return
				}
				if argSize == -1 {
					args[i] = nil
				} else {
					arg := make([]byte, argSize)
					_, err = c.reader.ReadFull(arg)
					if err != nil {
						break
					}
					args[i] = arg
				}
				c.reader.SkipBytes([]byte{config.CR, config.LF})
			}
			reply.Value = args
		}
	default:
		err = errors.New("Bad Reply Flag:" + string([]byte{b}))
	}
	return
}

func (c *Connection) WriteReply(reply *Reply) (err error) {
	switch reply.Type {
	case ReplyTypeStatus:
		err = c.replyStatus(reply.Value.(string))
	case ReplyTypeError:
		err = c.replyError(reply.Value.(string))
	case ReplyTypeInteger:
		err = c.replyInteger(reply.Value.(int))
	case ReplyTypeBulk:
		err = c.replyBulk(reply.Value)
	case ReplyTypeMultiBulks:
		err = c.replyMultiBulks(reply.Value.([]interface{}))
	default:
		err = errors.New("Illegal ReplyType: " + strconv.Itoa(int(reply.Type)))
	}
	return
}

// Status reply
func (c *Connection) replyStatus(status string) (err error) {
	buf := bytes.Buffer{}
	buf.WriteString("+")
	buf.WriteString(status)
	buf.WriteString(config.CRLF)
	_, err = buf.WriteTo(c)
	return
}

// Error reply
func (c *Connection) replyError(errmsg string) (err error) {
	buf := bytes.Buffer{}
	buf.WriteString("-")
	buf.WriteString(errmsg)
	buf.WriteString(config.CRLF)
	_, err = buf.WriteTo(c)
	return
}

// Integer reply
func (c *Connection) replyInteger(i int) (err error) {
	buf := bytes.Buffer{}
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(i))
	buf.WriteString(config.CRLF)
	_, err = buf.WriteTo(c)
	return
}

// Bulk Reply
func (c *Connection) replyBulk(bulk interface{}) (err error) {
	// NULL Bulk Reply
	isnil := bulk == nil
	if !isnil {
		// []byte 需要类型转换后才能判断
		b, ok := bulk.([]byte)
		isnil = ok && b == nil
	}
	if isnil {
		_, err = c.Write([]byte("$-1\r\n"))
		return
	}
	buf := bytes.Buffer{}
	buf.WriteString("$")
	switch bulk.(type) {
	case []byte:
		b := bulk.([]byte)
		buf.WriteString(strconv.Itoa(len(b)))
		buf.WriteString(config.CRLF)
		buf.Write(b)
	default:
		b := []byte(bulk.(string))
		buf.WriteString(strconv.Itoa(len(b)))
		buf.WriteString(config.CRLF)
		buf.Write(b)
	}
	buf.WriteString(config.CRLF)
	_, err = buf.WriteTo(c)
	return
}

// Multi-bulk replies
func (c *Connection) replyMultiBulks(bulks []interface{}) (err error) {
	// Null Multi Bulk Reply
	if bulks == nil {
		_, err = c.Write([]byte("*-1\r\n"))
		return
	}
	bulkCount := len(bulks)
	// Empty Multi Bulk Reply
	if bulkCount == 0 {
		_, err = c.Write([]byte("*0\r\n"))
		return
	}
	buf := bytes.Buffer{}
	buf.WriteString("*")
	buf.WriteString(strconv.Itoa(bulkCount))
	buf.WriteString(config.CRLF)
	for i := 0; i < bulkCount; i++ {
		bulk := bulks[i]
		switch bulk.(type) {
		case string:
			buf.WriteString("$")
			b := []byte(bulk.(string))
			buf.WriteString(strconv.Itoa(len(b)))
			buf.WriteString(config.CRLF)
			buf.Write(b)
			buf.WriteString(config.CRLF)
		case []byte:
			b := bulk.([]byte)
			if b == nil {
				buf.WriteString("$-1")
				buf.WriteString(config.CRLF)
			} else {
				buf.WriteString("$")
				buf.WriteString(strconv.Itoa(len(b)))
				buf.WriteString(config.CRLF)
				buf.Write(b)
				buf.WriteString(config.CRLF)
			}
		case int:
			buf.WriteString(":")
			buf.WriteString(strconv.Itoa(bulk.(int)))
			buf.WriteString(config.CRLF)
		default:
			// nil element
			buf.WriteString("$-1")
			buf.WriteString(config.CRLF)
		}
	}
	// flush
	_, err = buf.WriteTo(c)
	return
}
