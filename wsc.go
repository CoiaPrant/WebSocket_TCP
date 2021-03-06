package main

import (
	"crypto/tls"
	"net"
	"strings"

	"github.com/CoiaPrant/zlog"
	"golang.org/x/net/websocket"
)

func LoadWSCRules(i string, rule Rule) {
	tcpaddress, _ := net.ResolveTCPAddr("tcp", ":"+rule.Port)

	ln, err := net.ListenTCP("tcp", tcpaddress)
	if err == nil {
		zlog.Info("Loaded [", i, "] (WebSocket Client) ", rule.Port, " => ", rule.Address)
	} else {
		zlog.Error("Load failed [", i, "] (WebSocket Client) Error:", err)
		return
	}

	for {
		conn, err := ln.Accept()

		if err != nil {
			if err, ok := err.(net.Error); ok && err.Temporary() {
				continue
			}
			break
		}
		go wsc_handleRequest(conn, rule)
	}
}

func wsc_handleRequest(conn net.Conn, r Rule) {
	var ws_config *websocket.Config
	var err error

	if r.TLS {
		ws_config, err = websocket.NewConfig("wss://"+r.Address+"/ws/", "https://"+r.Address+"/ws/")
		ws_config.TlsConfig = &tls.Config{InsecureSkipVerify: true}
	} else {
		ws_config, err = websocket.NewConfig("ws://"+r.Address+"/ws/", "http://"+r.Address+"/ws/")
	}

	if err != nil {
		_ = conn.Close()
		return
	}
	ws_config.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36")
	ws_config.Header.Set("X-Forward-For", ParseAddrToIP(conn.RemoteAddr().String()))
	ws_config.Header.Set("X-Forward-Protocol", conn.RemoteAddr().Network())
	ws_config.Header.Set("X-Forward-Address", conn.RemoteAddr().String())
	proxy, err := websocket.DialConfig(ws_config)
	if err != nil {
		_ = conn.Close()
		return
	}
	proxy.PayloadType = websocket.BinaryFrame

	go copyIO(conn, proxy)
	go copyIO(proxy, conn)
}

func ParseAddrToIP(addr string) string {
	var str string
	arr := strings.Split(addr, ":")
	for i := 0; i < (len(arr) - 1); i++ {
		if i != 0 {
			str = str + ":" + arr[i]
		} else {
			str = str + arr[i]
		}
	}
	return str
}
