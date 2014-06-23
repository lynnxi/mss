package main

import (
	"fmt"
	"hash/crc32"
	"mss/config"
	"mss/lib/stdlog"
	"net"
	//	"reflect"
)

type Server struct {
	redisProxys map[string]*Proxy
}

func newServer() (server *Server) {
	server = &Server{redisProxys: map[string]*Proxy{}}
	for _, host := range config.AppRedisConfig {
		server.redisProxys[host] = NewProxy(host)
	}
	return
}

func (server *Server) GetProxy(cmd *Command) *Proxy {
	i := int(crc32.ChecksumIEEE(cmd.Key()))
	i = i % len(config.AppRedisConfig)
	host := config.AppRedisConfig[i]
	p := server.redisProxys[host]
	if p == nil {
		stdlog.Println("new proxy...")
		p = NewProxy(host)
		server.redisProxys[host] = p
	}

	return p
}

func (server *Server) Listen(host string) error {
	listener, err := net.Listen("tcp", host)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		stdlog.Println("accept conn...")
		if err != nil {
			return err
		}
		connection := NewConnection(conn)
		go server.handleConnection(connection)
	}

	return nil
}

func (server *Server) handleConnection(connection *Connection) {
	defer func() {
		stdlog.Println("conn closed...")
		connection.Close()
	}()
	for {
		//SESSION管理
		//解析数据
		command, err := connection.ReadCommand()
		// 常见的error是:
		// 1) io.EOF
		// 2) read tcp 127.0.0.1:51863: connection reset by peer
		if err != nil {
			fmt.Println("conn read err : " + err.Error())
			connection.Close()
			break
		}
		// varify command
		if err := command.verifyCommand(); err != nil {
			connection.WriteReply(ErrorReply(err))
			continue
		}

		// //找到对应得handler处理
		handlerDesc, exist := config.Cmdrules[command.Name()]["handler"]

		if !exist {
			connection.WriteReply(ErrorReply(config.BadCommandError))
		}

		reply, err := DescHandler[handlerDesc[0].(string)](server, command)
		// method := reflect.ValueOf(server).MethodByName(methodNames[0].(string))
		// in := []reflect.Value{reflect.ValueOf(connection), reflect.ValueOf(command)}
		// callResult := method.Call(in)
		// var reply *Reply
		// if callResult[0].Interface() != nil {
		// 	reply = callResult[0].Interface().(*Reply)
		// }

		if err != nil {
			connection.WriteReply(ErrorReply(config.BadCommandError))
		} else if reply != nil {
			err = connection.WriteReply(reply)
			if err != nil {
				stdlog.Println("conn write err : " + err.Error())
				connection.Close()
				break
			}
		}
	}

}
