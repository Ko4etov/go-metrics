// Package handler предоставляет обработчики HTTP-запросов для сервера метрик.
package handler

import (
	"context"
	"net/http"
)

// DBPing проверяет доступность базы данных.
func (h *Handler) DBPing(res http.ResponseWriter, req *http.Request) {
	if h.pgx == nil {
		http.Error(res, "Can't connect to DB", http.StatusInternalServerError)
		return
	}
	if err := h.pgx.Ping(context.Background()); err != nil {
		http.Error(res, "Can't connect to DB", http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}