package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *Handlers) ChatSummary(w http.ResponseWriter, r *http.Request) {
	// Заглушка
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNotImplemented)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"ошибка": "TODO: реализовать scatter-gather хендлер для /chat/summary",
	})
}
