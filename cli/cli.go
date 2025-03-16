// client/main.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"

	// 引入 logrus
	"github.com/sirupsen/logrus"
)

var (
	serverURL string
	id        int
	name      string
)

type client struct {
	Name string `json:"name"`
}

var rootCmd = &cobra.Command{
	Use:   "chat-client",
	Short: "WebSocket Chat CLI",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%s/login?id=%v", serverURL, id))
		if err != nil {
			// 使用 logrus 记录错误日志
			logrus.Fatalf("login failed: %v", err)
		}
		bs, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			// 使用 logrus 记录错误日志
			logrus.Fatalf("login failed: %v", err)
		}
		c := &client{}
		err = json.Unmarshal(bs, c)
		if err != nil {
			// 使用 logrus 记录错误日志
			logrus.Fatalf("login failed: %v", err)
		}
		name = c.Name
		url := fmt.Sprintf("ws://localhost:%s/ws?id=%v", serverURL, id)
		// 使用 logrus 记录信息日志
		logrus.Info(url)
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			// 使用 logrus 记录错误日志
			logrus.Fatalf("Connection error: %v", err)
		}
		defer conn.Close()

		go readMessages(conn)
		sendMessages(conn)
	},
}

func readMessages(conn *websocket.Conn) {
	for {
		typ, message, err := conn.ReadMessage()
		if err != nil {
			defer func() {
				os.Exit(0)
			}()
			if typ == websocket.CloseMessage {
				logrus.Info("Server closed")
				return
			}
			// 使用 logrus 记录错误日志
			logrus.Errorf("Read error: %v, type: %v", err, typ)
			return
		}
		fmt.Printf("\r%s\n> ", string(message))
	}
}

func sendMessages(conn *websocket.Conn) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(scanner.Text())); err != nil {
			// 使用 logrus 记录错误日志
			logrus.Errorf("Send error: %v", err)
			return
		}
		fmt.Print("> ")
	}
}

func main() {
	rootCmd.Flags().StringVarP(&serverURL, "server", "s", "ws://localhost:8080/ws", "WebSocket server URL")
	rootCmd.Flags().IntVarP(&id, "id", "i", 0, "user id")

	if err := rootCmd.Execute(); err != nil {
		// 使用 logrus 记录错误日志
		logrus.Fatalf("Command execution error: %v", err)
	}
}
