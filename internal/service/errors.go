package service

import "errors"

var (
	// ErrValidation 表示输入验证错误
	ErrValidation = errors.New("validation error")

	// ErrNotFound 表示资源未找到
	ErrNotFound = errors.New("not found")
)

// fieldError 用于创建字段验证错误
func fieldError(message string) error {
	return errors.Join(ErrValidation, errors.New(message))
}
