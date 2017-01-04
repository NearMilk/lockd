# lockd
[![Build Status](https://travis-ci.org/teambition/lockd.svg?branch=master)](https://travis-ci.org/teambition/lockd)

lock service for distributed system.

## Installation

```sh
go get -u github.com/teambition/lockd
```

## Documentation

API documentation can be found here:
## Usage

```go
go run lockd/main.go 
```

## Lock key

```sh
curl -d "key=aaaaaaa&timeout=10" http://127.0.0.1:14000/lock 
```


## Unlock key
```sh
curl -X DELETE http://127.0.0.1:14000/lock?key=aaaaaaa
```


### Done
 - 当多个客户端请求同一个锁（job，简单字符串描述），或者同一个客户端对同一个锁发起多个请求时，
 - 只有一个请求获得锁，其它未获得锁的请求阻塞。拿到锁的请求执行任务，执行完毕向服务端请求释放锁，
 - 支持请求自定义超时时间
 - 支持http协议
 - 锁释放后的广播
 
### Todo
 - Redis协议接口的支持
 - live reload 热升级机制
