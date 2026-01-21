package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/razponn/Resilient-Scatter-Gather/internal/mocks"
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

	// Локальные "клиенты" на один запрос.
	// Если в query пришли параметры, то (только для моков) делаем копию и переопределяем Delay/Fail.
	usersClient := h.users
	permsClient := h.perms
	vmClient := h.vm

	q := r.URL.Query()

	// user_delay, user_fail
	if base, ok := h.users.(mocks.UserServiceMock); ok {
		if d, ok := parseDuration(q.Get("user_delay")); ok {
			base.Delay = d
		}
		if b, ok := parseBool(q.Get("user_fail")); ok {
			base.Fail = b
		}
		usersClient = base
	}

	// perms_delay, perms_fail, perms_allowed
	if base, ok := h.perms.(mocks.PermissionsServiceMock); ok {
		if d, ok := parseDuration(q.Get("perms_delay")); ok {
			base.Delay = d
		}
		if b, ok := parseBool(q.Get("perms_fail")); ok {
			base.Fail = b
		}
		if b, ok := parseBool(q.Get("perms_allowed")); ok {
			base.Allowed = b
		}
		permsClient = base
	}

	// vm_delay, vm_fail
	if base, ok := h.vm.(mocks.VectorMemoryMock); ok {
		if d, ok := parseDuration(q.Get("vm_delay")); ok {
			base.Delay = d
		}
		if b, ok := parseBool(q.Get("vm_fail")); ok {
			base.Fail = b
		}
		vmClient = base
	}

	// Fan-out: запускаем 3 запроса параллельно
	userCh := make(chan userResult, 1)
	permsCh := make(chan permsResult, 1)
	vmCh := make(chan ctxResult, 1)

	go func() {
		u, err := usersClient.GetUser(ctx, userID)
		userCh <- userResult{user: u, err: err}
	}()

	go func() {
		p, err := permsClient.CheckAccess(ctx, userID, chatID)
		permsCh <- permsResult{perms: p, err: err}
	}()

	go func() {
		c, err := vmClient.GetContext(ctx, chatID)
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

func parseDuration(s string) (time.Duration, bool) {
	if s == "" {
		return 0, false
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, false
	}
	return d, true
}

func parseBool(s string) (bool, bool) {
	if s == "" {
		return false, false
	}
	switch s {
	case "1", "true", "TRUE", "yes", "YES", "y", "Y":
		return true, true
	case "0", "false", "FALSE", "no", "NO", "n", "N":
		return false, true
	default:
		return false, false
	}
}
