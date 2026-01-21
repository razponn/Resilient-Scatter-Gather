package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/razponn/Resilient-Scatter-Gather/internal/app"
)

func main() {
	приложение := app.New()

	сервер := &http.Server{
		Addr:              ":8080",
		Handler:           приложение.Router(),
		ReadHeaderTimeout: 3 * time.Second,
	}

	// Запуск HTTP-сервера
	go func() {
		log.Printf("сервер запущен: %s", сервер.Addr)
		if err := сервер.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ошибка запуска сервера: %v", err)
		}
	}()

	// Корректная остановка по SIGINT/SIGTERM
	останов := make(chan os.Signal, 1)
	signal.Notify(останов, syscall.SIGINT, syscall.SIGTERM)
	<-останов

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("остановка сервера...")
	if err := сервер.Shutdown(ctx); err != nil {
		log.Fatalf("ошибка остановки сервера: %v", err)
	}
	log.Println("сервер остановлен")
}
