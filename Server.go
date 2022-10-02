package rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"rpc/Codec"
	"sync"
)
const MagicNumber = 0x3bef5c
type Option struct {
	MagicNumber int        // MagicNumber marks this's a geerpc request
	CodecType   Codec.Type // client may choose different Codec to encode body
}
var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   Codec.GobType,
}
type Server struct{
	//Ip string
	//Port int
}
func NewServer() *Server {
	return &Server{
		//Ip:ip,
		//Port:port,
	}
}
var DefaultServer = NewServer()
//accept连接
func (server *Server) Accept(listener net.Listener) {
	for {
		log.Println("start to accept")
		conn, err := listener.Accept()
		if err != nil {
			log.Println("error:accept:serbver.go line 38:", err)
			return
		}
		log.Println("success to accept a connection")
		go server.ServeConn(conn)
	}
}
//verify the connection
//conn的值解码放到opt，opt的type信息再写进map的type里，得到这个conn携带的信息并返回codec结构体
func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("error:option:serbver.go line 50:", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("error: invalid magic number %x:server.go line 54", opt.MagicNumber)
		return
	}
	f:=Codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("error: invalid codec type %s::server.go line 59", opt.CodecType)
		return
	}
	cc:=f(conn)
	server.serveCodec(cc)
}
var invalidRequest=struct{}{}
//先读请求，发送响应，再起n个协程处理请求直到请求被处理完毕（这里主协程要等到所有协程执行完退出）
func (server *Server) serveCodec(cc Codec.Codec) {
	sending := new(sync.Mutex) // make sure to send a complete response
	wg := new(sync.WaitGroup)  // wait until all request are handled
	for {
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break // it's not possible to recover, so close the connection
			}
			req.Header.Error = err.Error()
			server.sendResponse(cc, req.Header, invalidRequest, sending)//接收到错误信息，发个空body过去
			continue
		}
	    wg.Add(1)
		go server.handleRequest(cc, req, sending, wg)
	}
	wg.Wait()
	_ = cc.Close()
}
type Request struct {
	Header            *Codec.Header // header of request
	Argv, Replyv reflect.Value // argv and replyv of request
}
//获取header的值，传出到h
func (server *Server) readRequestHeader(cc Codec.Codec) (*Codec.Header, error) {
	var h Codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("error:header:server.go line 95", err)
		}
		return nil, err
	}
	return &h, nil
}
//获取body的值，传出到argv
func (server *Server) readRequest(cc Codec.Codec) (*Request, error) {
	header, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &Request{Header: header}
	// TODO: now we don't know the type of request argv
	// day 1, just suppose it's string
	req.Argv = reflect.New(reflect.TypeOf(""))
	if err = cc.Readbody(req.Argv.Interface()); err != nil {
		log.Println("error:body :server.go line 112", err)
	}
	return req, nil
}
//调用Codec.wirte函数发送响应，需要加锁
func (server* Server)sendResponse(cc Codec.Codec,header *Codec.Header,body interface{},sending *sync.Mutex){
	sending.Lock()
	defer sending.Unlock()
	if err:=cc.Write(body,header);err!=nil{
		log.Println("error:Codec:Codec.write:server.go line 120:",err)
	}
}	
//收到请求验证成功，调用sendresponse发送head和body过去
func (server* Server)handleRequest(cc Codec.Codec,req *Request,sending *sync.Mutex,wg *sync.WaitGroup){
	defer wg.Done()
	req.Replyv=reflect.ValueOf(fmt.Sprintf("success to accpet your request,response:%d",req.Header.Seq))
	server.sendResponse(cc, req.Header, req.Replyv.Interface(), sending)
}
/*func (this Server)Startserver(){    
    listener,err:=net.Listen("tcp",fmt.Sprintf("%s:%d",this.Ip,this.Port))
	log.Println("server listening,ip:%s port:%d",this.Ip,this.Port)
    if err!= nil{
        fmt.Println("error:net listener:",err)
        return
    }
   for{ 
       rpc.Accept(listener)
       if err!=nil{
           fmt.Println("error:net accept:",err)
           continue
       }
   }
}*/
func Accept(lis net.Listener) { DefaultServer.Accept(lis) }




