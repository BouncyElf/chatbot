package main

import (
	"chatbot_server/handler"
	"chatbot_server/model"

	"github.com/sirupsen/logrus"
)

func main() {
	// 初始化数据库
	dsn := "user:password@tcp(db:3306)/chatdb?charset=utf8mb4&parseTime=True&loc=Local"
	if err := model.InitDB(dsn); err != nil {
		logrus.Fatal("数据库初始化失败:", err)
	}

	handler.Run(":8080")
}
