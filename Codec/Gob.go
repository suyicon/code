package Codec
import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
    
)
type GobCodec struct {
	conn io.ReadWriteCloser//连接（文件）
	buf  *bufio.Writer//内存缓冲区
	dec  *gob.Decoder//解码器
	enc  *gob.Encoder//编码器
}
var _ Codec = (*GobCodec)(nil)
//创建一个Gob编解码器
//读：之间从conn读进内存
//写：把内存信息写到buf，再通过buf发送到文件（连接）中去
func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),//解码连接（文件）的内容
		enc:  gob.NewEncoder(buf),//编译内存的内容
	}
}
func (c *GobCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h)
}
func (c *GobCodec) Readbody(body interface{}) error {
	return c.dec.Decode(body)
}
func (c *GobCodec) Close() error {
	return c.conn.Close()
}
func (c *GobCodec) Write(body interface{},h *Header) (err error) {
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()
	if err := c.enc.Encode(h); err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}
	if err := c.enc.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return err
	}
	return nil
}



