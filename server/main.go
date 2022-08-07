package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
)

type Reply struct {
	Status  bool
	Message string
}

type Verify struct {
	Username string `json:"username"`
	RemoteIP string `json:"remoteip"`
}

func (c *Verify) Create(payload []byte, reply *Reply) error {
	reply.Status = true
	reply.Message = "Create Algıladım!"
	fmt.Println("execute function:", string(payload))
	// Burada remote ip set etmem lazım Verify içine yada bir üstte farketmez
	return nil
}

func main() {
	s := rpc.NewServer()
	if err := s.Register(new(Verify)); err != nil {
		log.Fatal(err)
	}

	listen, err := net.Listen("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Listen error:", err)
	}

	log.Println("Starting verify server on localhost:1234")
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		//go s.ServeCodec(jsonrpc.NewServerCodec(conn))
		go s.ServeCodec(NewServerCodec(conn, conn.RemoteAddr().String()))
	}
}
