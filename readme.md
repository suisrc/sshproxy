# 说明

ssh-proxy  
  
通过登录名，进行ssh分流， 中继和代理

server-name[/user-name]@host
其中 user-name 默认值为 user

执行业务应用的分流和监控

## 测试

```sh
# 测试命令：
# 服务: 
TARGET_ADDR={host}:22 go run main.go
# 终端: 
ssh x.x.x.x/root:pass@127.0.0.1
```

## 测试的TARGET_ADDR

```bash
{host}-{port}-{ssvc}{snum}/user:pass@127.0.0.1
# 相当于, 这个是特殊用来解决集群内部 suisrc/webtop | suisrc/vscode 应用 ssh 连接问题
# 但是 {host}:22 是默认值存在
ssh user@{host}-0.vsc-{ssvc}-dev.ws{snum}.svc.cluster.local:{port}
```

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