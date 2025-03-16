package model

import (
	"gorm.io/gorm"
)

// 客户端模型（对应 clients 表）
type Client struct {
	gorm.Model
	Name string `gorm:"type:varchar(255);not null;index" json:"name"`
}

func GetClient(id uint) (*Client, error) {
	res := &Client{}
	result := DB.First(res, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return res, nil
}

func RegisterClient(name string) (*Client, error) {
	client := Client{Name: name}
	result := DB.Create(&client)
	if result.Error != nil {
		return nil, result.Error
	}
	return &client, nil
}

// 消息模型（对应 messages 表）
type Message struct {
	gorm.Model
	ClientID  uint   `gorm:"not null;index"`
	Content   string `gorm:"type:text;not null"`
	Client    Client `gorm:"foreignKey:ClientID"` // 外键关联
	SentByBot bool
}

type Feedback struct {
	gorm.Model
	ClientID uint   `gorm:"not null;index"`
	Content  string `gorm:"type:text;not null"`
	Client   Client `gorm:"foreignKey:ClientID"` // 外键关联
}

func SaveFeedback(clientID uint, content string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 验证客户端存在性
		if err := tx.First(&Client{}, clientID).Error; err != nil {
			return err
		}

		// 创建消息记录
		feedback := Feedback{
			ClientID: clientID,
			Content:  content,
		}
		return tx.Create(&feedback).Error
	})
}

func SaveMessage(clientID uint, content string, isBot bool) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 验证客户端存在性
		if err := tx.First(&Client{}, clientID).Error; err != nil {
			return err
		}

		// 创建消息记录
		message := Message{
			ClientID:  clientID,
			Content:   content,
			SentByBot: isBot,
		}
		return tx.Create(&message).Error
	})
}
