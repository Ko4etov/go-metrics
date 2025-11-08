package handler

import (
	"context"
	"net/http"
)

func (h *Handler) DbPing(res http.ResponseWriter, req *http.Request) {
	if err := h.pgx.Ping(context.Background()); err != nil {
		http.Error(res, "Can't connect to DB", http.StatusInternalServerError)
	}

	res.WriteHeader(http.StatusOK)
}
