package sshd

import (
	"context"
	"io"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// 请求转发
func (aa *Proxy) ForwardRequest(reqs <-chan *ssh.Request, targetConn ssh.Conn, tTag string) {
	for req := range reqs {
		result, payload, err := targetConn.SendRequest(req.Type, req.WantReply, req.Payload)
		if err != nil {
			continue
		}
		_ = req.Reply(result, payload)
	}
}

// 通道转发
func (aa *Proxy) ForwardChannel(ctx context.Context, targetConn ssh.Conn, originChannel ssh.NewChannel, tTag string) {
	targetChan, targetReqs, err := targetConn.OpenChannel(originChannel.ChannelType(), originChannel.ExtraData())
	if err != nil {
		logrus.Error(tTag, "open target channel error")
		_ = originChannel.Reject(ssh.ConnectionFailed, "open target channel error")
		return
	}
	defer targetChan.Close()

	originChan, originReqs, err := originChannel.Accept()
	if err != nil {
		logrus.Error(tTag, "accept origin channel failed")
		return
	}
	defer originChan.Close()

	maskedReqs := make(chan *ssh.Request, 1)

	go func() {
		for req := range originReqs {
			maskedReqs <- req
		}
		close(maskedReqs)
	}()

	originChannelWg := sync.WaitGroup{}
	originChannelWg.Add(3)
	targetChannelWg := sync.WaitGroup{}
	targetChannelWg.Add(3)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = io.Copy(targetChan, originChan)
		_ = targetChan.CloseWrite()
		targetChannelWg.Done()
		targetChannelWg.Wait()
		_ = targetChan.Close()
	}()

	go func() {
		defer wg.Done()
		_, _ = io.Copy(originChan, targetChan)
		_ = originChan.CloseWrite()
		originChannelWg.Done()
		originChannelWg.Wait()
		_ = originChan.Close()
	}()

	go func() {
		_, _ = io.Copy(targetChan.Stderr(), originChan.Stderr())
		targetChannelWg.Done()
	}()

	go func() {
		_, _ = io.Copy(originChan.Stderr(), targetChan.Stderr())
		originChannelWg.Done()
	}()

	forward := func(sourceReqs <-chan *ssh.Request, targetChan ssh.Channel, channelWg *sync.WaitGroup) {
		defer channelWg.Done()
		for ctx.Err() == nil {
			select {
			case req, ok := <-sourceReqs:
				if !ok {
					return
				}
				b, err := targetChan.SendRequest(req.Type, req.WantReply, req.Payload)
				_ = req.Reply(b, nil)
				if err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}

	go forward(maskedReqs, targetChan, &targetChannelWg)
	go forward(targetReqs, originChan, &originChannelWg)

	wg.Wait()
	logrus.Debug("session forward stop")
}
