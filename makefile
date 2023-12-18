.PHONY: start build

NOW = $(shell date -u '+%Y%m%d%I%M%S')

APP = sshproxy
SERVER_BIN = ./docker/app

dev: start

# 初始化mod
init:
	go mod init ${APP}
#go mod init github.com/suisrc/${APP}

# 修正依赖
tidy:
	go mod tidy

# 打包应用 go build -gcflags=-G=3 -o $(SERVER_BIN) ./cmd
build:
	go build -ldflags "-w -s" -o $(SERVER_BIN) ./
