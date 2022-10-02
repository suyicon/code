package Codec

import  "io"

type Header struct{
    ServiceMethod string
    Seq uint64
    Error string
}
type Codec interface
{
    io.Closer
    ReadHeader(*Header)error
    Readbody(interface{})error
    Write(interface{},*Header)error
}
type Type string
const(
    GobType  Type = "application/gob"
    JsonType Type = "application/json" // not implemented
)
type NewCodecFunc func(io.ReadWriteCloser) Codec
var NewCodecFuncMap map[Type]NewCodecFunc
func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)//need to verify
	NewCodecFuncMap[GobType] = NewGobCodec
}


