package handlers

// Handlers — контейнер для HTTP-хендлеров.
// Позже сюда добавим зависимости (клиенты сервисов), чтобы хендлер собирал ответ.
type Handlers struct{}

func New() *Handlers {
	return &Handlers{}
}
