package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/soluty/link"
	"github.com/soluty/link/codec"
	"testing"
	"time"
)

func TestJsonToy(t *testing.T) {
	json := codec.Json()
	json.Register(AddReq{})
	json.Register(AddRsp{})

	server ,_ := link.Listen("test", "", json, 0 /* sync send */, link.HandlerFunc(serverSessionLoop))
	client, _ := link.Dial("test", "", json, 0)
	client.Send(&AddReq{1,2})
	rsp, _ := client.Receive()
	if rsp.(*AddRsp).C != 3 {
		t.Fail()
	}
	server.Stop()
}

func TestJsonToy2(t *testing.T) {
	json := codec.Json()
	json.Register(AddReq{})
	json.Register(AddRsp{})

	server ,_ :=link.Listen("test", "", json, 0 /* sync send */, link.HandlerFunc(serverSessionLoop))
	client, _ := link.Dial("test", "", json, 0)
	client.Send(&AddReq{10,-2})
	rsp, _ := client.Receive()
	if rsp.(*AddRsp).C != 8 {
		t.Fail()
	}
	server.Stop()
}

func TestPb(t *testing.T) {
	pb := codec.Protobuf()
	pb.Register(1, &C2SLogin{})

	login := &C2SLogin{}
	login.Version = proto.Int32(21)
	server ,_ :=link.Listen("test", "", pb, 0 /* sync send */, link.HandlerFunc(func(session *link.Session) {
		for {
			req, err := session.Receive()
			checkErr(err)
			l , ok := req.(*C2SLogin)
			if !ok {
				t.Failed()
			}
			fmt.Println(l.GetVersion())
		}
	}))
	client, _ := link.Dial("test", "", pb, 0)
	client.Send(login)
	time.Sleep(time.Millisecond*10)
	server.Stop()
}