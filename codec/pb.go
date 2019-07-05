package codec

import (
	"encoding/binary"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/soluty/link"
	"io"
	"reflect"
)

// len, uint(cmd), pb

type ProtobufProtocol struct {
	types map[uint16]reflect.Type
	names map[reflect.Type]uint16
	order binary.ByteOrder
}

type protobufCodec struct {
	p      *ProtobufProtocol
	rw     io.ReadWriter
	closer io.Closer
}

func (p *ProtobufProtocol) NewCodec(rw io.ReadWriter) (link.Codec, error) {
	codec := &protobufCodec{
		p:  p,
		rw: rw,
	}
	codec.closer, _ = rw.(io.Closer)
	return codec, nil
}

func Protobuf() *ProtobufProtocol {
	return &ProtobufProtocol{
		types: make(map[uint16]reflect.Type),
		names: make(map[reflect.Type]uint16),
		order: binary.LittleEndian,
	}
}

func (p *ProtobufProtocol) Register(cmd uint16, msg proto.Message) {
	name := proto.MessageName(msg)
	rt := proto.MessageType(name)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	p.types[cmd] = rt
	p.names[rt] = cmd
}

func (c *protobufCodec) Receive() (interface{}, error) {
	var lenBytes = make([]byte, 2)
	n, err := c.rw.Read(lenBytes)
	if err != nil || n != 2 {
		return nil, errors.New("read head error")
	}
	var cmdBytes = make([]byte, 2)
	n, err = c.rw.Read(cmdBytes)
	if err != nil || n != 2 {
		return nil, errors.New("read head error")
	}
	length := c.p.order.Uint16(lenBytes)
	cmd := c.p.order.Uint16(cmdBytes)
	var body interface{}
	if t, exists := c.p.types[cmd]; exists {
		body = reflect.New(t).Interface()
	} else {
		return nil, nil
	}
	var pbBytes = make([]byte, length)
	n, err = io.ReadFull(c.rw, pbBytes)
	if err != nil || n != int(length) {
		return nil, errors.New("read body error")
	}
	err = proto.Unmarshal(pbBytes, body.(proto.Message))
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (c *protobufCodec) Send(msg interface{}) error {
	rt := reflect.TypeOf(msg)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if cmd, ok := c.p.names[rt]; !ok {
		return errors.New("msg type not register")
	} else {
		bs, err := proto.Marshal(msg.(proto.Message))
		if err != nil {
			return err
		}
		err = binary.Write(c.rw, c.p.order, uint16(len(bs)))
		if err != nil {
			return err
		}
		err = binary.Write(c.rw, c.p.order, cmd)
		if err != nil {
			return err
		}
		_, err = c.rw.Write(bs)
		if err != nil {
			return err
		}
		return nil
	}
}

func (c *protobufCodec) Close() error {
	return c.closer.Close()
}
