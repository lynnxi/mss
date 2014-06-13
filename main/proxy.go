package main

import (
	//	"fmt"
	"mss/lib/redigo/redis"
	//"mss/lib/stdlog"
	//"net"
	//"fmt"
	"time"
)

type Proxy struct {
	//conn *Connection
	host string
	pool *redis.Pool
}

func NewProxy(config string) (proxy *Proxy) {
	proxy = &Proxy{host: config}
	// c, err := net.DialTimeout("tcp", config, time.Millisecond*1000)
	// if err != nil {
	// 	return
	// }
	// proxy.conn = NewConnection(c)
	proxy.pool = &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config)
			if err != nil {
				return nil, err
			}
			// if _, err := c.Do("AUTH", ""); err != nil {
			// 	c.Close()
			// 	return nil, err
			// }
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	return
}

func (proxy *Proxy) Dispatch(cmd *Command) (r *Reply, err error) {
	// err = proxy.conn.WriteCommand(cmd)
	// if err != nil {
	// 	stdlog.Println("write cmd error " + err.Error())
	// 	proxy.conn.Close()
	// } else {
	// 	r, err = proxy.conn.ReadReply()
	// }
	conn := proxy.pool.Get()
	defer conn.Close()
	// conn.Do("set", "b", "2")
	// _r, err := conn.Do("get", "b")
	// fmt.Println(string(_r.([]byte)[0]))
	//stdlog.Println("proxy cmd " + proxy.host + " : " + cmd.String())
	_v, err := conn.Do(cmd.Name(), cmd.Args()[1:])
	//fmt.Println(_v)
	r = &Reply{Value: _v}
	//fmt.Println(string(_r.([]byte)))
	return
}
