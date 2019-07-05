package main

import (
	"encoding/binary"
	"io"
	"log"

	"github.com/soluty/link"
	"github.com/soluty/link/codec"
)

type AddReq struct {
	A, B int
}

type AddRsp struct {
	C int
}

func main() {
	json := codec.Json()
	json.Register(AddReq{})
	json.Register(AddRsp{})

	fix := codec.FixLen(json,2 , binary.LittleEndian, 1000,1000)

	server, err := link.Listen("tcp", "0.0.0.0:0", fix, 0 /* sync send */, link.HandlerFunc(serverSessionLoop))
	checkErr(err)
	addr := server.Listener().Addr().String()
	go server.Serve()

	client, err := link.Dial("tcp", addr, fix, 0)
	checkErr(err)
	clientSessionLoop(client)
}

func serverSessionLoop(session *link.Session) {
	for {
		req, err := session.Receive()
		if err == io.ErrClosedPipe {
			return
		}
		checkErr(err)

		err = session.Send(&AddRsp{
			req.(*AddReq).A + req.(*AddReq).B,
		})
		checkErr(err)
	}
}

func clientSessionLoop(session *link.Session) {
	for i := 0; i < 10; i++ {
		err := session.Send(&AddReq{
			i, i,
		})
		checkErr(err)
		log.Printf("Send: %d + %d", i, i)

		rsp, err := session.Receive()
		checkErr(err)
		log.Printf("Receive: %d", rsp.(*AddRsp).C)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
