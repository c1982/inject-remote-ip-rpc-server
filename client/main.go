package main

import (
	"fmt"
	"log"
	"net/rpc/jsonrpc"
)

type Reply struct {
	Status  bool
	Message string
}

type Verify struct {
	Username string `json:"username"`
	RemoteIP string `json:"remoteip"`
}

func main() {

	client, err := jsonrpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal(err)
	}

	r := &Reply{}
	err = client.Call("Verify.Create", []byte(`{"username":"batman"}`), &r)
	if err != nil {
		log.Fatal(err)
	}

	client.Close()
	fmt.Println(r.Message)
}
