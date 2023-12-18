package main

import (
	"os"
	"sshproxy/sshd"

	"github.com/sirupsen/logrus"
)

// 启动 ssh 代理服务，
// 测试命令：
// 服务: TARGET_ADDR={uname}:22 go run main.go
// 终端: ssh x.x.x.x/root:pass@127.0.0.1
func main() {
	sshp := &sshd.Proxy{
		CertsPath:  os.Getenv("CERTS_PATH"),
		ListenAddr: os.Getenv("LISTEN_ADDR"),
		TargetAddr: os.Getenv("TARGET_ADDR"),
	}
	logrus.Info("ssh proxy server is starting...")
	logrus.Info("certs path: ", sshp.CertsPath)
	logrus.Info("listen addr: ", sshp.ListenAddr)
	logrus.Info("target addr: ", sshp.TargetAddr)

	// 启动服务
	if err := sshp.Start(); err != nil {
		logrus.Error("ssh proxy server start failed: ", err.Error())
		return
	}
	// 等待服务停止
	sshp.Wait()

	logrus.Info("ssh proxy server is stopped")

}
