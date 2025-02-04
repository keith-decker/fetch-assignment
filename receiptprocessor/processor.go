package receiptprocessor

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/keith-decker/fetch-assignment/kvstore"
	"github.com/keith-decker/fetch-assignment/pb"
)

func ProcessReceipt(receipt *pb.Receipt) string {
	// Generate an ID for this receipt
	// @TODO: create a hash of the receipt to prevent duplicates. Date/Time + Store ID + Total?
	id := uuid.New().String()

	// store the receipt in the KV store, kick off the processing and return the ID
	kv := kvstore.New()
	// set the score to -1 to indicate that the receipt is being processed
	kv.Set(fmt.Sprintf("receipt-%s", id), "-1")

	go processReceipt(id, receipt)

	return id
}

func processReceipt(id string, receipt *pb.Receipt) {
	// Process the receipt
	kv := kvstore.New()
	kv.Set(fmt.Sprintf("receipt-%s", id), "99")
}
