package sshd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// 双向数据拷贝
func (aa *Proxy) HandleTcpConn(ssc net.Conn, addr string) {
	defer ssc.Close()
	// 转发流量到目标服务器
	ttc, err := net.DialTimeout("tcp", addr, time.Second*10)
	if err != nil {
		logrus.Errorf("bridge tcp connection error [%s -> %s]: %s", ssc.RemoteAddr().String(), addr, err.Error())
		return
	}
	defer ttc.Close()
	// 双向数据拷贝
	go io.Copy(ttc, ssc) // c -> s
	io.Copy(ssc, ttc)    // s -> c
}

// 处理SSH连接
func (aa *Proxy) HandleSshConn(ssc net.Conn, config *ssh.ServerConfig) {
	defer ssc.Close()
	// SSH 握手
	cConn, cChans, cReqs, err := ssh.NewServerConn(ssc, config)
	if err != nil {
		ssc.Close()
		if errors.Is(err, io.EOF) {
			return // 客户端主动断开连接
		}
		logrus.Error("failed to establish ssh proxy to client:", ssc.RemoteAddr().String(), " -> ", err.Error())
		return
	}
	defer cConn.Close()
	//========================================================================
	// 解析登录信息，user, ip
	cUser := cConn.User()
	if cUser == "" {
		logrus.Error("failed to get user name from ssh connection", ssc.RemoteAddr().String())
		return
	}

	// 密码
	attr1 := strings.SplitN(cUser, ":", 2)
	if len(attr1) != 2 {
		logrus.Error("invalid user name: ", cUser)
		return
	}
	tPass := attr1[1]

	// 用户
	tUser := "user"
	attr2 := strings.SplitN(attr1[0], "/", 2)
	if len(attr2) == 2 {
		tUser = attr2[1]
	}
	tName := attr2[0]

	// 地址， uname-port-sname|nname, port会被忽略
	tAddr := aa.TargetAddr
	attr3 := strings.SplitN(attr2[0], "-", 3)
	if len(attr3) == 3 {
		// uname, sname, nname <- attr3[...]
		for i := 0; i < len(attr3[2]); i++ {
			if attr3[2][i] < '0' || attr3[2][i] > '9' {
				continue
			}
			attr3[1] = attr3[2][:i]
			attr3[2] = attr3[2][i:]
			break
		}
		tAddr = strings.ReplaceAll(tAddr, "{uname}", attr3[0])
		tAddr = strings.ReplaceAll(tAddr, "{sname}", attr3[1])
		tAddr = strings.ReplaceAll(tAddr, "{nname}", attr3[2])
	} else {
		// uname <- attr2[0]
		tAddr = strings.ReplaceAll(tAddr, "{uname}", attr2[0])
	}

	// 标签
	tTag := fmt.Sprintf("[%s: %s -> %s]", tName, ssc.RemoteAddr().String(), tAddr)
	//========================================================================

	ttc, err := net.Dial("tcp", tAddr)
	if err != nil {
		logrus.Error(tTag, " dial failed ")
		return
	}
	defer ttc.Close()
	// 执行代理转发
	logrus.Info(tTag, " ssh proxy >>> begin, username: ", tUser)
	defer func() {
		logrus.Info(tTag, " ssh proxy <<< final, username: ", tUser)
	}()
	tConn, tChans, tReqs, err := ssh.NewClientConn(ttc, tAddr, &ssh.ClientConfig{
		User:            tUser,
		Auth:            []ssh.AuthMethod{ssh.Password(tPass)},
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		// cConn.SendRequest("exit-status", false, []byte(err.Error()))
		logrus.Error(tTag, " failed to establish ssh proxy to target: ", err.Error())
		return
	}
	defer tConn.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// client -> workspace global request forward
	go aa.ForwardRequest(cReqs, tConn, tTag)
	// target -> client global request forward
	go aa.ForwardRequest(tReqs, cConn, tTag)

	// client -> target channel forward
	go func() {
		for nChan := range tChans {
			go aa.ForwardChannel(ctx, cConn, nChan, tTag)
		}
	}()
	// target -> client channel forward
	go func() {
		for nChan := range cChans {
			go aa.ForwardChannel(ctx, tConn, nChan, tTag)
		}
	}()

	// wait for client or target connection close
	go func() {
		cConn.Wait()
		cancel()
	}()
	// wait for client or target connection close
	go func() {
		tConn.Wait()
		cancel()
	}()

	// // cache connection
	// cc_key := fmt.Sprintf("%s <- %s", tTag, ssc.RemoteAddr().String())
	// aa.ConnCache.Store(cc_key, ssc) // 可用于中断连接
	// defer aa.ConnCache.Delete(cc_key)

	// wait for client or target connection close
	<-ctx.Done()
	cancel()
}
