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
	kv := kvstore.New()
	pointsString, err := kv.Get(fmt.Sprintf("receipt-%s", r.PathValue("id")))
	if err != nil {
		http.Error(w, "No receipt found for that ID.", http.StatusNotFound)
		return
	}

	points, err := strconv.Atoi(pointsString)
	if err != nil {
		log.Print(err)
		http.Error(w, "The points are invalid.", http.StatusInternalServerError)
		return
	}

	getResponse := &pb.GetPointsResponse{
		Points: int32(points),
	}

	response, err := protojson.Marshal(getResponse)
	w.Write(response)
}

func processReceipt(w http.ResponseWriter, r *http.Request) {
	// Process the receipt
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
	id := receiptprocessor.ProcessReceipt(receipt)

	processResponse := &pb.ProcessReceiptResponse{
		Id: id,
	}

	response, err := protojson.Marshal(processResponse)
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
