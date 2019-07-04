package main

import (
	"github.com/soluty/link"
	"github.com/soluty/link/codec"
	"testing"
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