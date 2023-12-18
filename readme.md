# 说明

ssh-proxy  
  
通过登录名，进行ssh分流， 中继和代理

server-name[/user-name]@host
其中 user-name 默认值为 user

执行业务应用的分流和监控

## golang

``` bash
go mod init sshproxy
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

# 生成 ssh 私钥
ssh-kekgen -A
```