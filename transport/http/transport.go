package httpt

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/advanderveer/brahms"
)

// Transport is a transport that uses an http client
type Transport struct {
	client *http.Client
	logs   *log.Logger
}

// New initializes the transport
func New(logw io.Writer) (tr *Transport) {
	tr = &Transport{client: &http.Client{}, logs: log.New(logw, "httpt/transport: ", 0)}
	return
}

//Request performs a http request on the provided node an decodes the response into msg
func (tr *Transport) Request(ctx context.Context, method string, n brahms.Node, path string, body io.Reader, msg interface{}) (err error) {
	loc := "http://" + n.IP.String() + ":" + strconv.Itoa(int(n.Port)) + path
	req, err := http.NewRequest(method, loc, body)
	if err != nil {
		return TransportErr{err, "request_creation"}
	}

	req = req.WithContext(ctx)
	resp, err := tr.client.Do(req)
	if err != nil {
		return TransportErr{err, "request_execution"}
	}

	if msg != nil {
		defer resp.Body.Close()
		dec := json.NewDecoder(resp.Body)
		err = dec.Decode(msg)
		if err != nil {
			return TransportErr{err, "response_decoding"}
		}
	}

	return nil
}

// RequestOrLog will perform the request and log the error if anything fails
func (tr *Transport) RequestOrLog(ctx context.Context, method string, n brahms.Node, path string, body io.Reader, msg interface{}) {
	err := tr.Request(ctx, method, n, path, body, msg)
	if err != nil {
		tr.logs.Printf("failed to perform request: %v", err)
	}
}

// Push implements node information pushing
func (tr *Transport) Push(ctx context.Context, self brahms.Node, to brahms.Node) {
	msg := MsgNode{IP: self.IP, Port: self.Port}
	data, _ := json.Marshal(MsgPushReq{msg})
	tr.RequestOrLog(ctx, http.MethodGet, to, "/push", bytes.NewReader(data), nil)
}

// Pull impelents node information pulling
func (tr *Transport) Pull(ctx context.Context, c chan<- brahms.View, from brahms.Node) {
	var msg MsgPullResp
	tr.RequestOrLog(ctx, http.MethodGet, from, "/pull", nil, &msg)

	v := make(brahms.View)
	for _, m := range msg {
		n := brahms.Node{IP: m.IP, Port: m.Port}
		v[n.Hash()] = n
	}

	c <- v
}

// Probe implements node status probing
func (tr *Transport) Probe(ctx context.Context, c chan<- int, idx int, n brahms.Node) {
	msg := new(MsgProbeResp)
	tr.RequestOrLog(ctx, http.MethodGet, n, "/probe", nil, msg)
	if msg.Active {
		c <- idx
	}
}
