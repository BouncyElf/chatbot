package handler

import (
	"os"
	"testing"

	"gopkg.in/yaml.v2"
)

// 测试 LoadBotConfig 函数
func TestLoadBotConfig(t *testing.T) {
	// 创建一个临时配置文件
	tempFile, err := os.CreateTemp("", "bot-config.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	config := BotMsgConfig{
		Openning:        "Test Openning",
		Ending:          "Test Ending",
		Default:         "Test Default",
		GuideToFeedback: "Test GuideToFeedback",
		StartReview:     "Test StartReview",
		EndKeywords:     []string{"bye"},
		ReviewKeywords:  []string{"feedback"},
		Keywords:        map[string]string{"test": "test response"},
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	_, err = tempFile.Write(data)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// 修改工作目录以加载临时配置文件
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldDir)

	err = os.Chdir(os.TempDir())
	if err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}

	LoadBotConfig(tempFile.Name())

	if BotMsgTemp.Openning != "Test Openning" {
		t.Errorf("Expected Openning %s, got %s", "Test Openning", BotMsgTemp.Openning)
	}
}

// 测试 Client.handleMessage 函数
func TestClientHandleMessage(t *testing.T) {
	// 初始化 BotMsgTemp
	config := BotMsgConfig{
		Openning:        "Test Openning",
		Ending:          "Test Ending",
		Default:         "Test Default",
		GuideToFeedback: "Test GuideToFeedback",
		StartReview:     "Test StartReview",
		EndKeywords:     []string{"bye"},
		ReviewKeywords:  []string{"feedback"},
		Keywords:        map[string]string{"test": "test response"},
	}
	BotMsgTemp = &config

	client := &Client{
		conn:   nil,
		send:   make(chan []byte, 256),
		client: nil,
		state:  state_init,
	}

	// 测试包含 ReviewKeywords 的消息
	message1 := "I want to give feedback"
	reply1 := client.handleMessage(message1)
	if reply1 != "Test StartReview" {
		t.Errorf("Expected reply %s, got %s", "Test StartReview", reply1)
	}

	// 测试包含 EndKeywords 的消息
	message2 := "bye"
	reply2 := client.handleMessage(message2)
	if reply2 != "Test Ending" {
		t.Errorf("Expected reply %s, got %s", "Test Ending", reply2)
	}

	// 测试包含 Keywords 的消息
	message3 := "This is a test"
	reply3 := client.handleMessage(message3)
	if reply3 != "test response" {
		t.Errorf("Expected reply %s, got %s", "test response", reply3)
	}

	// 测试 state_wait_feedback 状态下的默认消息
	client.state = state_wait_feedback
	message4 := "Some other message"
	reply4 := client.handleMessage(message4)
	if reply4 != "Test GuideToFeedback" {
		t.Errorf("Expected reply %s, got %s", "Test GuideToFeedback", reply4)
	}

	// 测试 state_handle_feedback 状态下的默认消息
	client.state = state_handle_feedback
	message5 := "Another message"
	reply5 := client.handleMessage(message5)
	if reply5 != "Test StartReview" {
		t.Errorf("Expected reply %s, got %s", "Test StartReview", reply5)
	}

	// 测试其他情况的默认消息
	client.state = state_init
	message6 := "No keyword message"
	reply6 := client.handleMessage(message6)
	if reply6 != "Test Default" {
		t.Errorf("Expected reply %s, got %s", "Test Default", reply6)
	}
}
