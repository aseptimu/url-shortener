// example_test.go демонстрирует, как пользоваться URLService из пакета service.
package service_test

import (
	"context"
	"fmt"

	"github.com/aseptimu/url-shortener/internal/app/service"
)

// stubTestStore — простая реализация StoreURLSetter для примера.
type stubTestStore struct{}

// Set просто возвращает переданный shortURL без ошибок.
func (s *stubTestStore) Set(_ context.Context, shortURL, _, _ string) (string, error) {
	return shortURL, nil
}

// BatchSet возвращает тот же мапинг, что и получили на вход.
func (s *stubTestStore) BatchSet(_ context.Context, urls map[string]string, _ string) (map[string]string, error) {
	return urls, nil
}

// ExampleURLService показывает, как создать URLService и сократить один и несколько URL.
//
// Этот пример будет запускаться в GoDoc Play и через go test (если вы добавите блок // Output:).
func ExampleURLService() {
	// Инициализируем сервис на нашем stub-хранилище
	svc := service.NewURLService(&stubTestStore{})
	ctx := context.Background()

	// Сокращаем один URL
	short, err := svc.ShortenURL(ctx, "http://example.com", "user123")
	fmt.Println("short:", short, "err:", err)

	// Сокращаем сразу несколько URL
	batch, err := svc.ShortenURLs(ctx,
		[]string{"http://foo.com", "http://bar.com"},
		"user123",
	)
	fmt.Println("batch result:", batch, "err:", err)

	// Если вы хотите, чтобы go test проверял вывод, добавьте сюда:
	// Output:
	// short: XyZ123 err: <nil>
	// batch result: map[XyZ123:http://foo.com AbC456:http://bar.com] err: <nil>
}
