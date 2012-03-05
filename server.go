/*
Copyright 2012, Google Inc.
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

    * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
    * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
    * Neither the name of Google Inc. nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,           
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY           
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package bsonrpc

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/rpc"
    "log"
    "strings"
)

const (
	connected = "200 Connected to Go RPC"
)

type ClientCodecFactory func(conn io.ReadWriteCloser) rpc.ClientCodec

type BufferedConnection struct {
	*bufio.Reader
	io.WriteCloser
}

func NewBufferedConnection(conn io.ReadWriteCloser) *BufferedConnection {
	return &BufferedConnection{bufio.NewReader(conn), conn}
}


type ServerCodecFactory func(conn io.ReadWriteCloser) rpc.ServerCodec

// ServeRPC handles rpc requests using the hijack scheme of rpc
func ServeRPC() {
	http.Handle(GetRpcPath(codecName), &rpcHandler{NewServerCodec})
}

// ServeHTTP handles rpc requests in HTTP compliant POST form
func ServeHTTP() {
	http.Handle(GetHttpPath(codecName), &httpHandler{NewServerCodec})
}

// listen and run
func ListenAndServe(addr string) error {
    var netaddr net.Addr
    var err error
    if strings.Contains(addr,"/") {
        netaddr,err = net.ResolveUnixAddr("unix",addr)
        if err != nil {
            return err
        }
    } else {
        netaddr,err = net.ResolveTCPAddr("tcp",addr)
        if err != nil {
            return err
        }
    }
    // listen
    l,err := net.Listen(netaddr.Network(),netaddr.String())
    if err != nil {
        return err
    }
    // same with ServeRpc
    http.Handle(GetRpcPath(codecName), &rpcHandler{NewServerCodec})

    err = http.Serve(l,nil)
    return err
}

type rpcHandler struct {
	cFactory ServerCodecFactory
}

func (self *rpcHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    if req.Method != "CONNECT" {
        w.Header().Set("Content-Type","text/plain; charset=utf-8")
        w.WriteHeader(http.StatusMethodNotAllowed)
        io.WriteString(w,"405 must CONNECT\n")
        return
    }
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Printf("rpc hijacking %s: %v", req.RemoteAddr, err)
		return
	}
	io.WriteString(conn, "HTTP/1.0 "+connected+"\n\n")
	rpc.ServeCodec(self.cFactory(NewBufferedConnection(conn)))
}

func GetRpcPath(codecName string) string {
	return "/_" + codecName + "_rpc_"
}

type httpHandler struct {
	cFactory ServerCodecFactory
}

func (self *httpHandler) ServeHTTP(c http.ResponseWriter, req *http.Request) {
	conn := &httpConnectionBroker{c, req.Body}
	codec := self.cFactory(conn)
	if err := rpc.ServeRequest(codec); err != nil {
		log.Printf("rpcwrap: %v", err)
	}
}

func GetHttpPath(codecName string) string {
	return "/_" + codecName + "_http_"
}

type httpConnectionBroker struct {
    http.ResponseWriter
    io.Reader
}

func (self *httpConnectionBroker) Close() error {
    return nil
}
