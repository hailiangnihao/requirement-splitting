package service

import (
	"fmt"

	"github.com/google/uuid"
)

// generateID 生成标准 UUID
func generateID() string {
	return uuid.New().String()
}

// generateSimpleID 生成带前缀的 UUID（用于兼容现有代码）
func generateSimpleID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, uuid.New().String())
}
