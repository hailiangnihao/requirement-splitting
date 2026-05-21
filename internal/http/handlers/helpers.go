package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"requirement-splitting/internal/service"
)

// writeJSON 写入 JSON 响应
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// writeError 写入错误响应
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// writeServiceError 根据 service 层错误类型返回合适的 HTTP 状态码
func writeServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrValidation) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}

// fieldError 用于 service 层返回字段验证错误
func fieldError(message string) error {
	return errors.Join(service.ErrValidation, errors.New(message))
}
