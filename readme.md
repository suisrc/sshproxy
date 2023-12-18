# 说明

ssh-proxy  
  
通过登录名，进行ssh分流， 中继和代理

server-name[/user-name]@host
其中 user-name 默认值为 user

执行业务应用的分流和监控


## kratos

致力于快速搭建基于 gin + wire 服务框架，应用于应用的快速开发

如果需要完成的鉴权体系，推荐使用kratos-golang版本， 该 vkc 版本只是一个简单的框架，不包含鉴权体系
同时， vkc 是为效率而生， 因此唯一推荐持久化数据库为 mongodb。如果使用 mysql 推荐使用 kratos-golang 版本

## golang

``` bash
go mod init vkc
go mod tidy
go run main.go
go build


```

## 自签名证书

``` bash
cd shconf/cert

# 生成私钥
openssl genrsa -out default.key.pem 2048

# 生成自签名证书
openssl req -new -x509 -key default.key -out default.crt.pem -days 36500
```