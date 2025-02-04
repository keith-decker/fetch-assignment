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

type ReceiptResponse struct {
	ID string `json:"id"`
}

type PointsResponse struct {
	Points int `json:"points"`
}

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
	mux := buildRouter()
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
	mux := buildRouter()
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

func TestProcessReceiptEmptyReceipt(t *testing.T) {
	req, err := http.NewRequest("POST", "/receipts/process", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rec := httptest.NewRecorder()
	mux := buildRouter()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400; got %d", rec.Code)
	}

	body := rec.Body.String()
	expected := "The receipt is invalid."
	if !strings.Contains(body, expected) {
		t.Errorf("expected body to contain %q, got %q", expected, body)
	}
}

func TestProcessReceiptInvalidReceipt(t *testing.T) {
	invalidReceipt := `{"invalid": "data"}`
	req, err := http.NewRequest("POST", "/receipts/process", strings.NewReader(invalidReceipt))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	mux := buildRouter()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400; got %d", rec.Code)
	}

	body := rec.Body.String()
	expected := "The receipt is invalid."
	if !strings.Contains(body, expected) {
		t.Errorf("expected body to contain %q, got %q", expected, body)
	}
}

func TestProcessReceipt(t *testing.T) {
	receipt := `{
  "retailer": "M&M Corner Market",
  "purchaseDate": "2022-03-20",
  "purchaseTime": "14:33",
  "items": [
    {
      "shortDescription": "Gatorade",
      "price": "2.25"
    },{
      "shortDescription": "Gatorade",
      "price": "2.25"
    },{
      "shortDescription": "Gatorade",
      "price": "2.25"
    },{
      "shortDescription": "Gatorade",
      "price": "2.25"
    }
  ],
  "total": "9.00"
}`
	req, err := http.NewRequest("POST", "/receipts/process", strings.NewReader(receipt))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	mux := buildRouter()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200; got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "id") {
		t.Errorf("expected body to contain an ID, got %q", body)
	}
}

func TestProcessReceiptInvalidRetailer(t *testing.T) {
	receipt := `{"retailer":"","purchaseDate":"not A Date","purchaseTime":"words","items":[{"shortDescription":"","price":"FREE!"},{"shortDescription":"Gatorade","price":"2.25"},{"shortDescription":"Gatorade","price":"2.25"},{"shortDescription":"Gatorade","price":"2.25"}],"total":"9.00"  }`
	req, err := http.NewRequest("POST", "/receipts/process", strings.NewReader(receipt))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	mux := buildRouter()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400; got %d", rec.Code)
	}

	body := rec.Body.String()
	expected := "The receipt is invalid."
	if !strings.Contains(body, expected) {
		t.Errorf("expected body to contain %q, got %q", expected, body)
	}
}
