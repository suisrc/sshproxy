package sshd

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// SshProxy 接口
type Proxy struct {
	KeysFolder string
	ListenAddr string
	TargetAddr string
	Listener   net.Listener
	ConnCache  sync.Map // 连接缓存
	wchan      chan int
}

func (aa *Proxy) Wait() {
	if aa.wchan != nil {
		<-aa.wchan
	}
}

func (aa *Proxy) Stop() error {
	if aa.Listener != nil {
		return aa.Listener.Close()
	}
	return nil
}

// 启动服务
func (aa *Proxy) Start() error {
	if aa.Listener != nil {
		return errors.New("ssh proxy server is already running")
	}
	conf, err := aa.InitConf()
	if err != nil {
		return err
	}
	// 启动服务
	aa.Listener, err = net.Listen("tcp", aa.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to start ssh  proxy server: %s", err.Error())
	}
	// 重置通知
	if aa.wchan != nil {
		close(aa.wchan)
	}
	aa.wchan = make(chan int, 1)
	// 监听连接
	go func() {
		defer aa.Listener.Close()
		logrus.Info("ssh proxy server is listening on: ", aa.ListenAddr)
		for {
			// 接受客户端连接
			conn, err := aa.Listener.Accept()
			if err != nil {
				if aa.Listener == nil || errors.Is(err, net.ErrClosed) {
					if aa.Listener != nil {
						aa.Listener = nil
					}
					logrus.Info("ssh proxy server is closed")
					close(aa.wchan)
					aa.wchan = nil
					return // 服务器已关闭
				}
				logrus.Error("failed to accept incoming connection:", err)
				continue
			}
			go aa.HandleSshConn(conn, conf)
		}
	}()
	return nil
}

// 启动服务
func (aa *Proxy) InitConf() (*ssh.ServerConfig, error) {
	// 服务配置
	conf := &ssh.ServerConfig{NoClientAuth: true}
	// 加载密钥, DSA 密钥已被弃用, ssh-keygen -A -f "__keys/"
	if aa.KeysFolder == "" {
		aa.KeysFolder = "/etc/ssh"
	}
	if err := aa.AddHotKey(conf, aa.KeysFolder+"/ssh_host_rsa_key"); err != nil {
		return nil, err
	}
	if err := aa.AddHotKey(conf, aa.KeysFolder+"/ssh_host_ecdsa_key"); err != nil {
		return nil, err
	}
	if err := aa.AddHotKey(conf, aa.KeysFolder+"/ssh_host_ed25519_key"); err != nil {
		return nil, err
	}
	// 服务监听地址
	if aa.ListenAddr == "" {
		aa.ListenAddr = ":22"
	}
	if aa.TargetAddr == "" {
		aa.TargetAddr = "{host}:22"
		// aa.TargetAddr = "{host}-0.vsc-{ssvc}-dev.ws{snum}.svc.cluster.local:22"
	}
	return conf, nil
}

// 增加密钥
func (aa *Proxy) AddHotKey(conf *ssh.ServerConfig, file string) error {
	bts, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %v", err.Error())
	}
	key, err := ssh.ParsePrivateKey(bts)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err.Error())
	}
	conf.AddHostKey(key)
	return nil
}
