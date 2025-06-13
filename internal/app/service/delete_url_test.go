package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type stubStoreDelete struct {
	called    bool
	gotURLs   []string
	gotUserID string
	errToRet  error
}

func (s *stubStoreDelete) BatchDelete(ctx context.Context, shortURLs []string, userID string) error {
	s.called = true
	s.gotURLs = append([]string(nil), shortURLs...)
	s.gotUserID = userID
	return s.errToRet
}

func TestDeleteURLs_Success(t *testing.T) {
	stub := &stubStoreDelete{errToRet: nil}
	svc := NewURLDeleter(stub)

	urls := []string{"u1", "u2"}
	user := "user42"

	err := svc.DeleteURLs(context.Background(), urls, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stub.called {
		t.Error("expected BatchDelete to be called")
	}
	if !reflect.DeepEqual(stub.gotURLs, urls) {
		t.Errorf("expected URLs %v, got %v", urls, stub.gotURLs)
	}
	if stub.gotUserID != user {
		t.Errorf("expected userID %q, got %q", user, stub.gotUserID)
	}
}

func TestDeleteURLs_Error(t *testing.T) {
	expectedErr := errors.New("delete failed")
	stub := &stubStoreDelete{errToRet: expectedErr}
	svc := NewURLDeleter(stub)

	urls := []string{"only-one"}
	user := "uXYZ"

	err := svc.DeleteURLs(context.Background(), urls, user)
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if !stub.called {
		t.Error("expected BatchDelete to be called even on error")
	}
}
