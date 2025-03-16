package handler

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"chatbot_server/model"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type client_state int

const (
	state_init client_state = iota
	state_wait_feedback
	state_handle_feedback
	state_over
)

var stateMap = map[client_state]client_state{
	state_init:            state_wait_feedback,
	state_wait_feedback:   state_handle_feedback,
	state_handle_feedback: state_over,
}

type BotMsgConfig struct {
	Openning        string            `yaml:"openning"`
	Ending          string            `yaml:"ending"`
	Default         string            `yaml:"default"`
	GuideToFeedback string            `yaml:"guide_to_feedback"`
	StartReview     string            `yaml:"start_review"`
	EndKeywords     []string          `yaml:"end_keywords"`
	ReviewKeywords  []string          `yaml:"review_keywords"`
	Keywords        map[string]string `yaml:"keywords"`
}

func (b *BotMsgConfig) show() {
	if b == nil {
		return
	}
	// 使用logrus记录日志
	logrus.Info(*b)
	for k, v := range b.Keywords {
		logrus.Info(k, v)
	}
}

var BotMsgTemp *BotMsgConfig

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	client *model.Client
	state  client_state
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	// 从URL参数获取昵称
	id := r.URL.Query().Get("id")
	clientID, err := strconv.Atoi(id)
	if err != nil || clientID < 0 {
		// 使用logrus记录错误日志
		logrus.Error("invalid id: ", id)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid format of client id"))
		return
	}

	client, err := model.GetClient(uint(clientID))
	if err != nil {
		// 使用logrus记录错误日志
		logrus.Error("get client failed: ", clientID, "err: ", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid client id, client not exists"))
		return
	}

	// 建立WebSocket连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// 使用logrus记录错误日志
		logrus.Error("WebSocket upgrade error:", err)
		return
	}

	c := &Client{
		conn:   conn,
		send:   make(chan []byte, 256),
		client: client,
	}

	go c.writePump()
	go c.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
	}()

	state := state_init

	if c.state == state_init {
		reply := BotMsgTemp.Openning
		c.send <- []byte(reply)
		c.state = stateMap[state]
		go func() {
			if err := model.SaveMessage(c.client.ID, string(reply), true); err != nil {
				// 使用logrus记录错误日志
				logrus.Error("save bot message error: ", err)
			} else {
				// 使用logrus记录信息日志
				logrus.Info("save bot message success: ", string(reply))
			}
		}()
	}

	for {
		if c.state == state_over {
			close(c.send)
			time.Sleep(1 * time.Second)
			break
		}
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		reply := c.handleMessage(string(message))
		c.send <- []byte(reply)

		go func() {
			if c.state == state_handle_feedback {
				if err := model.SaveFeedback(c.client.ID, string(message)); err != nil {
					logrus.Error("save feedback error: ", err)
				} else {
					logrus.Info("save feedback success: ", string(message))
				}
			}
			if err := model.SaveMessage(c.client.ID, string(reply), true); err != nil {
				// 使用logrus记录错误日志
				logrus.Error("save bot message error: ", err)
			} else {
				// 使用logrus记录信息日志
				logrus.Info("save bot message success: ", string(reply))
			}
			if err := model.SaveMessage(c.client.ID, string(message), false); err != nil {
				// 使用logrus记录错误日志
				logrus.Error("save message error: ", err)
			} else {
				// 使用logrus记录信息日志
				logrus.Info("save message success: ", string(message))
			}
		}()
	}
}

func (c *Client) handleMessage(message string) string {
	// 使用logrus记录信息日志
	logrus.Info(message)
	for _, v := range BotMsgTemp.ReviewKeywords {
		if strings.Contains(message, v) {
			c.state = state_handle_feedback
			return BotMsgTemp.StartReview
		}
	}
	for _, v := range BotMsgTemp.EndKeywords {
		if strings.Contains(message, v) {
			c.state = state_over
			return BotMsgTemp.Ending
		}
	}
	for k, v := range BotMsgTemp.Keywords {
		if strings.Contains(message, k) {
			// 使用logrus记录信息日志
			logrus.Info(message, k, v, strings.Contains(message, k))
			return v
		}
	}
	if c.state == state_wait_feedback {
		return BotMsgTemp.GuideToFeedback
	} else if c.state == state_handle_feedback {
		return BotMsgTemp.StartReview
	}
	return BotMsgTemp.Default
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func LoadBotConfig(path string) {
	config_path := path
	if config_path == "" {
		config_path = "./bot-config.yaml"
	}
	data, err := os.ReadFile(config_path)
	if err != nil {
		// 使用logrus记录错误日志
		logrus.Fatal("read fail: ", err)
	}
	BotMsgTemp = &BotMsgConfig{}
	if err := yaml.Unmarshal(data, BotMsgTemp); err != nil {
		// 使用logrus记录错误日志
		logrus.Fatal("unmarshal bot config error: ", err)
	}
	BotMsgTemp.show()
}

func Run(addr string) {
	LoadBotConfig("")
	// 创建一个默认的Gin引擎
	r := gin.Default()

	// 定义WebSocket处理路由
	r.GET("/ws", func(c *gin.Context) {
		handleWS(c.Writer, c.Request)
	})

	// 定义登录处理路由
	r.GET("/login", func(c *gin.Context) {
		id := c.Query("id")
		clientID, err := strconv.Atoi(id)
		if err != nil || clientID < 0 {
			logrus.Error("invalid id: ", id)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format of client id"})
			return
		}

		client, err := model.GetClient(uint(clientID))
		if err != nil {
			logrus.Error("get client failed: ", clientID, " err: ", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client id, client not exists"})
			return
		}
		c.JSON(http.StatusOK, client)
	})

	// 定义注册处理路由
	r.GET("/register", func(c *gin.Context) {
		name := c.Query("name")
		client, err := model.RegisterClient(name)
		if err != nil {
			logrus.Error("register client err: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "register client error"})
			return
		}
		c.JSON(http.StatusOK, client)
	})

	// 启动Gin服务器
	logrus.Info("server start")
	if err := r.Run(addr); err != nil {
		// 使用logrus记录错误日志
		logrus.Fatal("server start error: ", err)
	}
}
