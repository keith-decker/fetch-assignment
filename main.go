package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/keith-decker/fetch-assignment/kvstore"
	"github.com/keith-decker/fetch-assignment/pb"
	"github.com/keith-decker/fetch-assignment/receiptprocessor"
	"google.golang.org/protobuf/encoding/protojson"
)

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func getPoints(w http.ResponseWriter, r *http.Request) {
	receiptId := r.PathValue("id")
	points, err := getPointsFromStore(receiptId)

	if err != nil {
		http.Error(w, "No receipt found for that ID.", http.StatusNotFound)
		return
	}

	getResponse := &pb.GetPointsResponse{
		Points: int32(points),
	}

	response, err := protojson.Marshal(getResponse)
	if err != nil {
		http.Error(w, "An error occurred while processing the request.", http.StatusInternalServerError)
		return
	}
	w.Write(response)
}

func processReceipt(w http.ResponseWriter, r *http.Request) {
	// Process the receipt
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	if r.Body == nil {
		http.Error(w, "The receipt is invalid.", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	// get the bytes from the request
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print(err)
		http.Error(w, "The receipt is invalid.", http.StatusBadRequest)
		return
	}

	// unmarshal the bytes into a receipt
	receipt := &pb.Receipt{}
	err = protojson.Unmarshal(data, receipt)

	if err != nil {
		http.Error(w, "The receipt is invalid.", http.StatusBadRequest)
		log.Print(err)
		return
	}

	// pass the receipt to the processor, return the ID
	if !receiptprocessor.ValidateReceipt(receipt) {
		http.Error(w, "The receipt is invalid.", http.StatusBadRequest)
		return
	}

	id := receiptprocessor.ProcessReceipt(receipt)

	processResponse := &pb.ProcessReceiptResponse{
		Id: id,
	}

	response, err := protojson.Marshal(processResponse)
	if err != nil {
		http.Error(w, "An error occurred while processing the request.", http.StatusInternalServerError)
		return
	}
	w.Write(response)
}

func main() {
	// Replace this with GorillaMux, or Chi, or another router
	mux := buildRouter()

	log.Print("starting server on :8080")
	err := http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}

func buildRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", home)
	mux.HandleFunc("/receipts/{id}/points", getPoints)
	mux.HandleFunc("/receipts/process", processReceipt)
	return mux
}

func getPointsFromStore(id string) (int, error) {
	kv := kvstore.New()
	pointsString, err := kv.Get(fmt.Sprintf("receipt-%s", id))
	if err != nil {
		return -1, err
	}

	points, err := strconv.Atoi(pointsString)
	if err != nil {
		return -1, err
	}

	return points, nil
}
