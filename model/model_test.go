package model

import (
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	// 引入logrus
	"github.com/sirupsen/logrus"
)

var testDB *gorm.DB

func setupTestDB() {
	var err error
	testDB, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		logrus.Fatal("failed to connect database: ", err)
	}
	// 自动迁移模型
	testDB.AutoMigrate(&Client{}, &Message{}, &Feedback{})
	// 用 testDB 替换全局变量 DB
	DB = testDB
}

func teardownTestDB() {
	// 删除测试数据库文件
	os.Remove("test.db")
}

func TestMain(m *testing.M) {
	setupTestDB()
	code := m.Run()
	teardownTestDB()
	os.Exit(code)
}

func TestGetClient(t *testing.T) {
	// 注册一个客户端
	client, err := RegisterClient("test_user")
	if err != nil {
		t.Fatalf("RegisterClient failed: %v", err)
	}

	// 获取客户端
	fetchedClient, err := GetClient(client.ID)
	if err != nil {
		t.Fatalf("GetClient failed: %v", err)
	}

	if fetchedClient.ID != client.ID {
		t.Errorf("Expected client ID %d, got %d", client.ID, fetchedClient.ID)
	}
}

func TestRegisterClient(t *testing.T) {
	client, err := RegisterClient("test_user")
	if err != nil {
		t.Fatalf("RegisterClient failed: %v", err)
	}

	if client.Name != "test_user" {
		t.Errorf("Expected client name %s, got %s", "test_user", client.Name)
	}
}

func TestSaveMessage(t *testing.T) {
	// 注册一个客户端
	client, err := RegisterClient("test_user")
	if err != nil {
		t.Fatalf("RegisterClient failed: %v", err)
	}

	// 保存消息
	err = SaveMessage(client.ID, "test message", false)
	if err != nil {
		t.Fatalf("SaveMessage failed: %v", err)
	}
}
