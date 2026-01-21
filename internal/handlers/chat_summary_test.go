package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/razponn/Resilient-Scatter-Gather/internal/handlers"
	"github.com/razponn/Resilient-Scatter-Gather/internal/mocks"
)

func newTestHandler() *handlers.Handlers {
	return handlers.New(
		mocks.UserServiceMock{Delay: 10 * time.Millisecond, Fail: false},
		mocks.PermissionsServiceMock{Delay: 50 * time.Millisecond, Fail: false, Allowed: true},
		mocks.VectorMemoryMock{Delay: 100 * time.Millisecond, Fail: false},
	)
}

func decodeBody(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &m); err != nil {
		t.Fatalf("не удалось распарсить json: %v; body=%s", err, rr.Body.String())
	}
	return m
}

func TestChatSummary_Успех_КонтекстУспевает(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/chat/summary?user_id=1&chat_id=42", nil)
	rr := httptest.NewRecorder()

	start := time.Now()
	h.ChatSummary(rr, req)
	elapsed := time.Since(start)

	if rr.Code != http.StatusOK {
		t.Fatalf("ожидали 200, получили %d, body=%s", rr.Code, rr.Body.String())
	}
	if elapsed > 250*time.Millisecond {
		t.Fatalf("ответ слишком долгий: %s (ожидали укладываться в SLA)", elapsed)
	}

	body := decodeBody(t, rr)

	// context должен быть, degraded = false
	if _, ok := body["context"]; !ok {
		t.Fatalf("ожидали поле context в ответе, body=%s", rr.Body.String())
	}
	if v, ok := body["degraded"]; !ok || v.(bool) != false {
		t.Fatalf("ожидали degraded=false, body=%s", rr.Body.String())
	}
}

func TestChatSummary_Деградация_VectorMemoryМедленный(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/chat/summary?user_id=1&chat_id=42&vm_delay=3s", nil)
	rr := httptest.NewRecorder()

	start := time.Now()
	h.ChatSummary(rr, req)
	elapsed := time.Since(start)

	if rr.Code != http.StatusOK {
		t.Fatalf("ожидали 200, получили %d, body=%s", rr.Code, rr.Body.String())
	}
	if elapsed > 250*time.Millisecond {
		t.Fatalf("ответ слишком долгий: %s (ожидали деградацию и ответ в SLA)", elapsed)
	}

	body := decodeBody(t, rr)

	// context не должен прийти, degraded = true
	if _, ok := body["context"]; ok {
		t.Fatalf("не ожидали поле context при деградации, body=%s", rr.Body.String())
	}
	if v, ok := body["degraded"]; !ok || v.(bool) != true {
		t.Fatalf("ожидали degraded=true, body=%s", rr.Body.String())
	}
}

func TestChatSummary_500_ЕслиUserПадает(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/chat/summary?user_id=1&chat_id=42&user_fail=1", nil)
	rr := httptest.NewRecorder()

	start := time.Now()
	h.ChatSummary(rr, req)
	elapsed := time.Since(start)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("ожидали 500, получили %d, body=%s", rr.Code, rr.Body.String())
	}
	if elapsed > 250*time.Millisecond {
		t.Fatalf("слишком долгий ответ при ошибке критичного сервиса: %s", elapsed)
	}
}

func TestChatSummary_500_ЕслиPermissionsНеУспевает(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/chat/summary?user_id=1&chat_id=42&perms_delay=250ms", nil)
	rr := httptest.NewRecorder()

	start := time.Now()
	h.ChatSummary(rr, req)
	elapsed := time.Since(start)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("ожидали 500, получили %d, body=%s", rr.Code, rr.Body.String())
	}
	// Должно отвалиться по SLA ~200мс (с небольшим запасом на окружение)
	if elapsed > 300*time.Millisecond {
		t.Fatalf("слишком долгий ответ: %s (ожидали ограничение SLA)", elapsed)
	}
}
