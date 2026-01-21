package models

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Permissions struct {
	ChatID  string `json:"chat_id"`
	UserID  string `json:"user_id"`
	Allowed bool   `json:"allowed"`
}

type VectorContext struct {
	ChatID    string `json:"chat_id"`
	Snippet   string `json:"snippet"`
	Source    string `json:"source"`
	LatencyMs int64  `json:"latency_ms"`
}

// ChatSummaryResponse — ответ, который отдаёт Gateway клиенту
// Context — опциональный (может быть nil при деградации VectorMemory
type ChatSummaryResponse struct {
	User        User           `json:"user"`
	Permissions Permissions    `json:"permissions"`
	Context     *VectorContext `json:"context,omitempty"`

	// Для наглядности деградации (не обязательно, но удобно)
	Degraded bool `json:"degraded"`
}
