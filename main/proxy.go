package main

import (
	"mss/lib/redigo/redis"
	"mss/lib/stdlog"
	//"net"
	//"fmt"
	"time"
)

type Proxy struct {
	//conn *Connection
	host        string
	pool        *redis.Pool
	maxPoolSize int

	poolIn  chan redis.Conn
	poolOut chan redis.Conn
}

func NewProxy(config string) (proxy *Proxy) {
	proxy = &Proxy{host: config}
	// c, err := net.DialTimeout("tcp", config, time.Millisecond*1000)
	// if err != nil {
	// 	return
	// }
	// proxy.conn = NewConnection(c)
	proxy.maxPoolSize = 200
	proxy.poolOut = make(chan redis.Conn, proxy.maxPoolSize)
	proxy.poolIn = make(chan redis.Conn, proxy.maxPoolSize)

	proxy.pool = &redis.Pool{
		MaxIdle:     200,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			stdlog.Println("new conn...")
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
		// TestOnBorrow: func(c redis.Conn, t time.Time) error {
		// 	_, err := c.Do("PING")
		// 	return err
		// },
	}
	go proxy.producer()
	go proxy.consumer()
	return
}

func (proxy *Proxy) consumer() {
	var conn redis.Conn
	for {
		conn = <-proxy.poolIn
		conn.Close()
	}
}

func (proxy *Proxy) producer() {
	var conn redis.Conn
	for {
		conn = proxy.pool.Get()
		proxy.poolOut <- conn
	}
}

func (proxy *Proxy) GetConn() redis.Conn {
	return <-proxy.poolOut
}

func (proxy *Proxy) CloseConn(conn redis.Conn) {
	proxy.poolIn <- conn
}

func (proxy *Proxy) Dispatch(cmd *Command) (r *Reply, err error) {
	// err = proxy.conn.WriteCommand(cmd)
	// if err != nil {
	// 	stdlog.Println("write cmd error " + err.Error())
	// 	proxy.conn.Close()
	// } else {
	// 	r, err = proxy.conn.ReadReply()
	// }
	/**
	conn := proxy.pool.Get()
	defer conn.Close()
	**/

	conn := proxy.GetConn()
	defer proxy.CloseConn(conn)
	// conn.Do("set", "b", "2")
	// _r, err := conn.Do("get", "b")
	// fmt.Println(string(_r.([]byte)[0]))
	//stdlog.Println("proxy cmd " + proxy.host + " : " + cmd.String())
	_v, err := conn.Do(cmd.Name(), cmd.Args()[1:])

	r = &Reply{Value: _v}
	return
}
