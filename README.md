# bsonrpc

a rpc from http://code.google.com/p/vitess/

# go rpc server:

```go

package main

import (
    "net/rpc"
    "bsonrpc"
)


type Args struct {
    A, B int 
}

type Reply struct {
    C int 
}

type Arith int 


func (t *Arith) Add(args *Args, reply *Reply) error {
    reply.C = args.A + args.B
    return nil 
}

func (t *Arith) Mul(args *Args, reply *Reply) error {
    reply.C = args.A * args.B
    return nil 
}

func (t *Arith) NError(args *Args, reply *Reply) error {
    return errors.New("normalerror")
}

func main(){
    rpc.Register(new(Arith))
    bsonrpc.ListenAndServe("localhost:9999")
}

```

# python rpc client

```python

import bsonrpc                                                                  

addr = 'localhost:9998'
uri = 'http://%s/_bson_rpc_' % addr

client = bsonrpc.BsonRpcClient(uri,5.0)

reply = client.call('Arith.Add',{'A':7,'B':8})

print reply.sequence_id,reply.reply

reply = client.call('Arith.Mul',{'A':4,'B':8})

print reply.sequence_id,reply.reply
```
