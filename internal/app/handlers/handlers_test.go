package handlers

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aseptimu/url-shortener/internal/app/service"
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
				code:    http.StatusMethodNotAllowed,
				method:  http.MethodGet,
				body:    "Only POST method allowed\n",
				isError: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.want.method, "/", strings.NewReader("http://example.com"))

			w := httptest.NewRecorder()
			URLCreator(w, r)

			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			if !tt.want.isError {
				require.NoError(t, err)
			}

			require.NotEmpty(t, resBody)
			assert.Equal(t, res.StatusCode, tt.want.code)
			assert.Equal(t, tt.want.body, string(resBody))
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
				code:    http.StatusMethodNotAllowed,
				method:  http.MethodPatch,
				url:     "/abcdef",
				isError: true,
			},
		},
		{
			name: "empty path",
			want: want{
				code:    http.StatusMethodNotAllowed,
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

			GetURL(w, r)

			res := w.Result()

			defer res.Body.Close()

			require.NotEmpty(t, res.Body)
			assert.Equal(t, res.StatusCode, tt.want.code)
		})
	}
}
