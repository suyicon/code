package rpc

import (
	"errors"
	"rpc/Codec"
	"sync"
)
type Call struct{
    Seq uint64
    ServiceMethod string
    Args interface{}
    Reply interface{}
    Error error
    Done chan *Call
}
func (call *Call)done(){
    call.Done<-call
}
type Client struct{
    cc Codec.Codec
    option *Option
    sending sync.Mutex
    order sync.Mutex
    header Codec.Header
    seq uint64
    pending map[uint64]*Call
    close bool
    shutdown bool
}

var Errshutdown = errors.New("Error:Client.go:client is shutdown")
func (client *Client)Close()error{
    client.order.Lock()
    defer client.order.Unlock()
    if client.close{
        return Errshutdown
    }
    client.close=true
    return client.cc.Close()
}
func (client *Client)IsValuable()bool{
    client.order.Lock()
    defer client.order.Unlock()
    return !client.close&&!client.shutdown
}
func (client *Client)RegisterCall(call *Call)(uint64,error){
    client.order.Lock()
    defer client.order.Unlock()   
    if client.shutdown||client.close{
        return 0,Errshutdown
    }
    call.Seq=client.seq
    client.pending[call.Seq]=call
    client.seq++
    return call.Seq,nil
}
//这里我感觉临界区没必要加锁
func (client *Client)RemoveCall(seq uint64)*Call{
    call:=client.pending[seq]
    delete(client.pending,seq)
    return call
}
func (client *Client)terminate()(err error){
    client.sending.Lock()
	defer client.sending.Unlock()
	client.order.Lock()
	defer client.order.Unlock()
    client.shutdown=true
    for _,call:=range client.pending{
        call.Error=err
        call.done()
    }
}