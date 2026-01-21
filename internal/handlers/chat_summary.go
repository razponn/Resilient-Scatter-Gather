package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/razponn/Resilient-Scatter-Gather/internal/models"
)

type userResult struct {
	user models.User
	err  error
}

type permsResult struct {
	perms models.Permissions
	err   error
}

type ctxResult struct {
	ctx models.VectorContext
	err error
}

func (h *Handlers) ChatSummary(w http.ResponseWriter, r *http.Request) {
	// Жёсткий SLA на весь запрос
	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancel()

	// Для простоты берём user_id и chat_id из query:
	// /chat/summary?user_id=1&chat_id=42
	userID := r.URL.Query().Get("user_id")
	chatID := r.URL.Query().Get("chat_id")

	if userID == "" || chatID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"ошибка": "нужны параметры user_id и chat_id",
		})
		return
	}

	// Fan-out: запускаем 3 запроса параллельно
	userCh := make(chan userResult, 1)
	permsCh := make(chan permsResult, 1)
	vmCh := make(chan ctxResult, 1)

	go func() {
		u, err := h.users.GetUser(ctx, userID)
		userCh <- userResult{user: u, err: err}
	}()

	go func() {
		p, err := h.perms.CheckAccess(ctx, userID, chatID)
		permsCh <- permsResult{perms: p, err: err}
	}()

	go func() {
		c, err := h.vm.GetContext(ctx, chatID)
		vmCh <- ctxResult{ctx: c, err: err}
	}()

	// Fan-in: критичные результаты обязаны успеть
	var (
		userRes  userResult
		permsRes permsResult
		gotUser  bool
		gotPerms bool
	)

	for !(gotUser && gotPerms) {
		select {
		case <-ctx.Done():
			// SLA истёк (или клиент отменил запрос). Критичные не успели -> 500
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"ошибка": "критичный сервис не успел за SLA",
			})
			return

		case ur := <-userCh:
			gotUser = true
			userRes = ur

		case pr := <-permsCh:
			gotPerms = true
			permsRes = pr
		}
	}

	// Если критичные сервисы вернули ошибку -> 500
	if userRes.err != nil || permsRes.err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"ошибка": "критичный сервис вернул ошибку",
		})
		return
	}

	// Некритичный результат (VectorMemory): не блокируем ответ.
	// Если уже успел — берём. Если нет/ошибка — деградация.
	var (
		contextData *models.VectorContext
		degraded    bool
	)

	select {
	case vr := <-vmCh:
		if vr.err == nil {
			tmp := vr.ctx
			contextData = &tmp
			degraded = false
		} else {
			degraded = true
		}
	default:
		// VectorMemory ещё не успел — деградация
		degraded = true
	}

	resp := models.ChatSummaryResponse{
		User:        userRes.user,
		Permissions: permsRes.perms,
		Context:     contextData,
		Degraded:    degraded,
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
