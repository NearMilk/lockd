# tblock
## 已经完成的功能
 - 当多个客户端请求同一个锁（job，简单字符串描述），或者同一个客户端对同一个锁发起多个请求时，
 - 只有一个请求获得锁，其它未获得锁的请求阻塞。拿到锁的请求执行任务，执行完毕向服务端请求释放锁，
 - 支持请求自定义超时时间
 - 支持http协议
 - 锁释放后的广播
 
## 待完成功能
 - Redis协议接口的支持
 - live reload 热升级机制

# lockd
[![Build Status](https://travis-ci.org/teambition/lockd.svg?branch=master)](https://travis-ci.org/teambition/lockd)

An HTTP content negotiator for Go

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


### Done
 - 当多个客户端请求同一个锁（job，简单字符串描述），或者同一个客户端对同一个锁发起多个请求时，
 - 只有一个请求获得锁，其它未获得锁的请求阻塞。拿到锁的请求执行任务，执行完毕向服务端请求释放锁，
 - 支持请求自定义超时时间
 - 支持http协议
 - 锁释放后的广播
### Todo
 - Redis协议接口的支持
 - live reload 热升级机制
