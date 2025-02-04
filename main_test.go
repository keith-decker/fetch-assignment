package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keith-decker/fetch-assignment/kvstore"
)

func TestGetPoints(t *testing.T) {
	kv := kvstore.New()
	receiptId := "adb6b560-0eef-42bc-9d16-df48f30e89b2"
	points := rand.Intn(200)
	kv.Set(fmt.Sprintf("receipt-%s", receiptId), fmt.Sprintf("%d", points))

	req, err := http.NewRequest("GET", fmt.Sprintf("/receipts/%s/points", receiptId), nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/receipts/{id}/points", getPoints)

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200; got %d", rec.Code)
	}

	body := rec.Body.String()
	expected := fmt.Sprintf("{\"points\":%d}", points)
	if !strings.Contains(body, expected) {
		t.Errorf("expected body to contain %q, got %q", expected, body)
	}
}

func TestMissingReceipt(t *testing.T) {
	req, err := http.NewRequest("GET", "/receipts/missing/points", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/receipts/{id}/points", getPoints)

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404; got %d", rec.Code)
	}

	body := rec.Body.String()
	expected := "No receipt found for that ID."
	if !strings.Contains(body, expected) {
		t.Errorf("expected body to contain %q, got %q", expected, body)
	}
}
