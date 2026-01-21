package app

import (
	"net/http"
	"time"

	"github.com/razponn/Resilient-Scatter-Gather/internal/handlers"
	"github.com/razponn/Resilient-Scatter-Gather/internal/mocks"
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

	// Моки сервисов (поведение по умолчанию близко к ТЗ)
	usersMock := mocks.UserServiceMock{
		Delay: 10 * time.Millisecond,
		Fail:  false,
	}
	permsMock := mocks.PermissionsServiceMock{
		Delay:   50 * time.Millisecond,
		Fail:    false,
		Allowed: true,
	}
	vmMock := mocks.VectorMemoryMock{
		Delay: 100 * time.Millisecond,
		Fail:  false,
	}

	// API: сводка по чату
	h := handlers.New(usersMock, permsMock, vmMock)
	mux.HandleFunc("/chat/summary", h.ChatSummary)

	return &App{mux: mux}
}

func (a *App) Router() http.Handler {
	return a.mux
}
