package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/rpc"
	"sync"
)

var errMissingParams = errors.New("jsonrpc: request body missing params")

type serverCodec struct {
	dec      *json.Decoder // for reading JSON values
	enc      *json.Encoder // for writing JSON values
	c        io.Closer
	req      serverRequest
	seq      uint64
	mutex    sync.Mutex // protects seq, pending
	pending  map[uint64]*json.RawMessage
	remoteip string
}

// NewServerCodec returns a new rpc.ServerCodec using JSON-RPC on conn.
func NewServerCodec(conn io.ReadWriteCloser, remoteip string) rpc.ServerCodec {
	return &serverCodec{
		dec:      json.NewDecoder(conn),
		enc:      json.NewEncoder(conn),
		c:        conn,
		remoteip: remoteip,
		pending:  make(map[uint64]*json.RawMessage),
	}
}

type serverRequest struct {
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params"`
	Id     *json.RawMessage `json:"id"`
}

func (r *serverRequest) reset() {
	r.Method = ""
	r.Params = nil
	r.Id = nil
}

type serverResponse struct {
	Id     *json.RawMessage `json:"id"`
	Result any              `json:"result"`
	Error  any              `json:"error"`
}

func (c *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	c.req.reset()
	if err := c.dec.Decode(&c.req); err != nil {
		return err
	}
	r.ServiceMethod = c.req.Method
	c.mutex.Lock()
	c.seq++
	c.pending[c.seq] = c.req.Id
	c.req.Id = nil
	r.Seq = c.seq
	c.mutex.Unlock()

	return nil
}

func (c *serverCodec) ReadRequestBody(x any) error {
	if x == nil {
		return nil
	}
	if c.req.Params == nil {
		return errMissingParams
	}

	var params [1]any
	params[0] = x

	newArgv, err := c.injectRemoteIP(*c.req.Params)
	if err == nil {
		*c.req.Params = newArgv
	}

	return json.Unmarshal(*c.req.Params, &params)
}

func (c *serverCodec) injectRemoteIP(x []byte) (json.RawMessage, error) {
	var params [1]any
	params[0] = nil

	err := json.Unmarshal(x, &params)
	if err != nil {
		return nil, err
	}

	messageArgs, ok := params[0].(string)
	if !ok {
		return nil, errors.New("params cannot cast to string")
	}

	argv, err := base64.StdEncoding.DecodeString(messageArgs)
	if err != nil {
		return x, err
	}

	//TODO: move it into a new function
	remoteip := []byte(`,"remoteip":"` + c.remoteip + `"}`)

	closeIndex := bytes.LastIndexByte(argv, '}')
	if closeIndex == -1 {
		return x, nil
	}

	argv = append(argv[:closeIndex], remoteip...)
	params[0] = argv

	xx, _ := json.Marshal(params)
	return xx, nil
}

var null = json.RawMessage([]byte("null"))

func (c *serverCodec) WriteResponse(r *rpc.Response, x any) error {
	c.mutex.Lock()
	b, ok := c.pending[r.Seq]
	if !ok {
		c.mutex.Unlock()
		return errors.New("invalid sequence number in response")
	}
	delete(c.pending, r.Seq)
	c.mutex.Unlock()

	if b == nil {
		b = &null
	}
	resp := serverResponse{Id: b}
	if r.Error == "" {
		resp.Result = x
	} else {
		resp.Error = r.Error
	}
	return c.enc.Encode(resp)
}

func (c *serverCodec) Close() error {
	return c.c.Close()
}
