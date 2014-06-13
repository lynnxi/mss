package mio

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"mss/config"
	"strconv"
)

type Reader struct {
	br *bufio.Reader
}

func NewReader(rd io.Reader) (reader *Reader) {
	reader = &Reader{}
	reader.br = bufio.NewReader(rd)
	return
}

// 读取一行
func (reader *Reader) ReadLine() (line []byte, err error) {
	line, err = reader.br.ReadSlice(config.LF)
	if err == bufio.ErrBufferFull {
		return nil, errors.New("line too long")
	}
	if err != nil {
		return
	}
	i := len(line) - 2
	if i < 0 || line[i] != config.CR {
		err = errors.New("bad line terminator:" + string(line))
	}
	return line[:i], nil
}

// 读取字符串，遇到CRLF换行为止
func (reader *Reader) ReadString() (str string, err error) {
	var line []byte
	if line, err = reader.ReadLine(); err != nil {
		return
	}
	str = string(line)
	return
}

func (reader *Reader) ReadInt() (i int, err error) {
	var line string
	if line, err = reader.ReadString(); err != nil {
		return
	}
	i, err = strconv.Atoi(line)
	return
}

func (reader *Reader) ReadInt64() (i int64, err error) {
	var line string
	if line, err = reader.ReadString(); err != nil {
		return
	}
	i, err = strconv.ParseInt(line, 10, 64)
	return
}

// 覆盖提供读buffer
func (reader *Reader) Read(p []byte) (n int, err error) {
	return reader.br.Read(p)
}

func (reader *Reader) ReadByte() (b byte, err error) {
	return reader.br.ReadByte()
}

// 验证并跳过指定的字节，用于开始符和结束符的判断
func (reader *Reader) SkipByte(b byte) (err error) {
	var tmp byte
	tmp, err = reader.br.ReadByte()
	if err != nil {
		return
	}
	if tmp != b {
		err = errors.New(fmt.Sprintf("Illegal Byte [%d] != [%d]", tmp, b))
	}
	return
}

func (reader *Reader) SkipBytes(bs []byte) (err error) {
	for _, b := range bs {
		err = reader.SkipByte(b)
		if err != nil {
			break
		}
	}
	return
}

func (reader *Reader) ReadFull(buf []byte) (n int, err error) {
	return io.ReadFull(reader, buf)
}
