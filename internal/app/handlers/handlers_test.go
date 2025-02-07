package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aseptimu/url-shortener/internal/app/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockShortenURL(url string) (string, error) {
	if url == "http://example.com" {
		return "abcdef", nil
	}
	return "", errors.New("invalid URL format")
}

func TestURLCreator(t *testing.T) {
	service.ShortenURL = mockShortenURL

	router := gin.Default()
	router.POST("/", URLCreator)

	type want struct {
		code    int
		method  string
		body    string
		isError bool
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "successful URL creation",
			want: want{
				code:    http.StatusCreated,
				method:  http.MethodPost,
				body:    "http://localhost:8080/abcdef",
				isError: false,
			},
		},
		{
			name: "method not allowed",
			want: want{
				code:    http.StatusNotFound,
				method:  http.MethodGet,
				body:    "Only POST method allowed\n",
				isError: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.want.method, "/", strings.NewReader("http://example.com"))
			router.ServeHTTP(w, r)

			res := w.Result()
			require.NotNil(t, res)
			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}

func mockGetOriginalURL(input string) (string, bool) {
	if input != "/abcdef" {
		return "http://example.com", true
	}
	return "", false
}

func TestGetURL(t *testing.T) {

	router := gin.Default()
	router.GET("/:url", GetURL)

	service.GetOriginalURL = mockGetOriginalURL

	type want struct {
		code    int
		method  string
		url     string
		isError bool
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "successful response",
			want: want{
				code:    http.StatusTemporaryRedirect,
				method:  http.MethodGet,
				url:     "/abcdef",
				isError: false,
			},
		},
		{
			name: "wrong method",
			want: want{
				code:    http.StatusNotFound,
				method:  http.MethodPatch,
				url:     "/abcdef",
				isError: true,
			},
		},
		{
			name: "empty path",
			want: want{
				code:    http.StatusNotFound,
				method:  http.MethodPatch,
				url:     "/",
				isError: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.want.method, tt.want.url, nil)

			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			res := w.Result()
			require.NotNil(t, res)
			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}
