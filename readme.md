# bh-go-grpc-utils

捕获科技 grpc 工具

## config

see ./testdata/config.go

## dev environment

go version

> go version go1.12 darwin/amd64 & go module enable

etcd 

> etcdctl version: 3.3.12 & API version: 3.3


## test 

run server & run client

```bash

go run ./testdata/executable/server.go
# INFO[0000] server addr : 192.168.0.145:50051            
# INFO[0000] [info] etcd key : /bh_ikaigunag/bh_ikaigunag_server/192.168.0.145:50051 

go run ./testdata/executable/client.go
# INFO[0000] dial : bh_ikaigunag://ikaiguang/bh_ikaigunag_server 
# INFO[0000] client.UnaryEcho resp : message:"this is client request msg" 
# INFO[0000] client.UnaryEcho resp : message:"this is client request msg" 
# ...

```