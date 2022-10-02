package main

import (
	"encoding/json"
	"fmt"
	"rpc"
	"rpc/Codec"
	"log"
	"net"
	"time"
)

func startServer(addr chan string) {
	// pick a free port
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	rpc.Accept(l)
}

func main() {
	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)

	// in fact, following code is like a simple geerpc client
	conn, err := net.Dial("tcp", <-addr)
	if err!=nil{
		log.Println("error:dial:",err)
	}
	defer func() { _ = conn.Close() }()

	time.Sleep(time.Second)
	// send options
	_ = json.NewEncoder(conn).Encode(rpc.DefaultOption)
	cc := Codec.NewGobCodec(conn)
	// send request & receive response
	for i := 0; i < 5; i++ {
		h := &Codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(i),
		}
		err = cc.Write(fmt.Sprintf("geerpc req %d", h.Seq),h)
		if err!=nil{
		log.Println("error:write:",err)
	    }
		err = cc.ReadHeader(h)
		if err!=nil{
		log.Println("error:readheader:",err)
	    }
		var reply string
		err = cc.Readbody(&reply)
		if err!=nil{
		log.Println("error:readbody:",err)
	    }
		log.Println("reply:", reply)
	}
}
