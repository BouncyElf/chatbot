package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

func TestReadMessages(t *testing.T) {
	// 创建一个 WebSocket 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logrus.Fatalf("Failed to upgrade connection: %v", err)
		}
		defer conn.Close()

		// 发送消息
		err = conn.WriteMessage(websocket.TextMessage, []byte("test message"))
		if err != nil {
			logrus.Fatalf("Failed to send message: %v", err)
		}
	}))
	defer server.Close()

	// 解析服务器地址
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Failed to parse server URL: %v", err)
	}

	// 连接到服务器
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(u.String(), "http"), nil)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// 启动读取消息的 goroutine
	go readMessages(conn)

	// 等待一段时间
	time.Sleep(1 * time.Second)
}

func TestSendMessages(t *testing.T) {
	// 创建一个 WebSocket 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logrus.Fatalf("Failed to upgrade connection: %v", err)
		}
		defer conn.Close()

		// 读取消息
		_, _, err = conn.ReadMessage()
		if err != nil {
			logrus.Fatalf("Failed to read message: %v", err)
		}
	}))
	defer server.Close()

	// 解析服务器地址
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Failed to parse server URL: %v", err)
	}

	// 连接到服务器
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(u.String(), "http"), nil)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// 模拟输入
	go func() {
		time.Sleep(1 * time.Second)
		conn.WriteMessage(websocket.TextMessage, []byte("test message"))
	}()

	// 启动发送消息的 goroutine
	sendMessages(conn)
}