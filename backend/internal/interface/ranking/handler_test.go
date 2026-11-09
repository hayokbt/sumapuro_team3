package ranking

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	domainranking "gomoku-backend/internal/domain/ranking"
)

type repositoryStub struct {
	entries []domainranking.Entry
	err     error
}

func (r repositoryStub) List(context.Context) ([]domainranking.Entry, error) {
	return r.entries, r.err
}

func TestHandlerReturnsRankingEntries(t *testing.T) {
	entries := []domainranking.Entry{
		{Position: 1, AccountName: "first"},
		{Position: 2, AccountName: "second"},
	}
	request := httptest.NewRequest(http.MethodGet, "/ranking", nil)
	response := httptest.NewRecorder()

	NewHandler(repositoryStub{entries: entries}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var got []domainranking.Entry
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, entries) {
		t.Fatalf("ranking = %#v, want %#v", got, entries)
	}
}

func TestHandlerRejectsNonGet(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/ranking", nil)
	response := httptest.NewRecorder()

	NewHandler(repositoryStub{}).ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
	if response.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("Allow = %q, want %q", response.Header().Get("Allow"), http.MethodGet)
	}
}

func TestHandlerReturnsErrorWhenRepositoryFails(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/ranking", nil)
	response := httptest.NewRecorder()

	NewHandler(repositoryStub{err: errors.New("database unavailable")}).ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
}
