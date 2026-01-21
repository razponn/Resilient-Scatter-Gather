package app

import (
	"net/http"

	"github.com/razponn/Resilient-Scatter-Gather/internal/handlers"
)

type App struct {
	mux *http.ServeMux
}

func New() *App {
	mux := http.NewServeMux()

	// Технический эндпоинт (проверка, что сервис жив)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ок\n"))
	})

	// API: сводка по чату
	h := handlers.New()
	mux.HandleFunc("/chat/summary", h.ChatSummary)

	return &App{mux: mux}
}

func (a *App) Router() http.Handler {
	return a.mux
}
