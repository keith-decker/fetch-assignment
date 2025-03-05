package receiptprocessor_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/keith-decker/fetch-assignment/kvstore"
	"github.com/keith-decker/fetch-assignment/pb"
	"github.com/keith-decker/fetch-assignment/receiptprocessor"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestReceiptProcessor(t *testing.T) {
	var receipt1 = &pb.Receipt{}
	var receipt2 = &pb.Receipt{}

	err := protojson.Unmarshal([]byte(`{"retailer":"Target","purchaseDate":"2022-01-01","purchaseTime":"13:01","items":[{"shortDescription":"Mountain Dew 12PK","price":"6.49"},{"shortDescription":"Emils Cheese Pizza","price":"12.25"},{"shortDescription":"Knorr Creamy Chicken","price":"1.26"},{"shortDescription":"Doritos Nacho Cheese","price":"3.35"},{"shortDescription":"   Klarbrunn 12-PK 12 FL OZ  ","price":"12.00"}],"total":"35.35"}`), receipt1)

	if err != nil {
		t.Fatalf("could not unmarshal receipt1: %v", err)
	}

	err = protojson.Unmarshal([]byte(`{"retailer":"M&M Corner Market","purchaseDate":"2022-03-20","purchaseTime":"14:33","items":[{"shortDescription":"Gatorade","price":"2.25"},{"shortDescription":"Gatorade","price":"2.25"},{"shortDescription":"Gatorade","price":"2.25"},{"shortDescription":"Gatorade","price":"2.25"}],"total":"9.00"  }`), receipt2)

	if err != nil {
		t.Fatalf("could not unmarshal receipt2: %v", err)
	}

	kv := kvstore.New()

	t.Run("ProcessReceipt1", func(t *testing.T) {
		// Total Points: 28
		// Breakdown:
		//      6 points - retailer name has 6 characters
		//     10 points - 5 items (2 pairs @ 5 points each)
		//      3 Points - "Emils Cheese Pizza" is 18 characters (a multiple of 3)
		//                 item price of 12.25 * 0.2 = 2.45, rounded up is 3 points
		//      3 Points - "Klarbrunn 12-PK 12 FL OZ" is 24 characters (a multiple of 3)
		//                 item price of 12.00 * 0.2 = 2.4, rounded up is 3 points
		//      6 points - purchase day is odd
		//   + ---------
		//   = 28 points
		expected := 28
		// get the id
		isValid := receiptprocessor.ValidateReceipt(receipt1)
		if !isValid {
			t.Errorf("expected valid receipt, got invalid")
		}
		id := receiptprocessor.ProcessReceipt(receipt1)
		// pause for a moment to allow the kv store to update
		time.Sleep(1 * time.Second)
		// get the points
		pointsString, err := kv.Get(fmt.Sprintf("receipt-%s", id))
		if err != nil {
			t.Fatalf("could not get points for receipt1: %v", err)
		}
		points, err := strconv.Atoi(pointsString)
		if err != nil {
			t.Fatalf("could not convert points for receipt1: %v", err)
		}
		if points != expected {
			t.Errorf("expected %d, got %d", expected, points)
		}
	})

	t.Run("ProcessReceipt2", func(t *testing.T) {
		// Total Points: 109
		// Breakdown:
		//     50 points - total is a round dollar amount
		//     25 points - total is a multiple of 0.25
		//     14 points - retailer name (M&M Corner Market) has 14 alphanumeric characters
		//                 note: '&' is not alphanumeric
		//     10 points - 2:33pm is between 2:00pm and 4:00pm
		//     10 points - 4 items (2 pairs @ 5 points each)
		//   + ---------
		//   = 109 points
		expected := 109
		// get the id
		isValid := receiptprocessor.ValidateReceipt(receipt2)
		if !isValid {
			t.Errorf("expected valid receipt, got invalid")
		}
		id := receiptprocessor.ProcessReceipt(receipt2)
		// pause for a moment to allow the kv store to update
		time.Sleep(1 * time.Second)
		// get the points
		pointsString, err := kv.Get(fmt.Sprintf("receipt-%s", id))
		if err != nil {
			t.Fatalf("could not get points for receipt1: %v", err)
		}
		points, err := strconv.Atoi(pointsString)
		if err != nil {
			t.Fatalf("could not convert points for receipt1: %v", err)
		}
		if points != expected {
			t.Errorf("expected %d, got %d", expected, points)
		}
	})

	t.Run("ValidateReceipt", func(t *testing.T) {
		// test an invalid receipt
		var invalidReceipt = &pb.Receipt{}
		err = protojson.Unmarshal([]byte(`{"retailer":"","purchaseDate":"not A Date","purchaseTime":"words","items":[{"shortDescription":"","price":"FREE!"},{"shortDescription":"Gatorade","price":"2.25"},{"shortDescription":"Gatorade","price":"2.25"},{"shortDescription":"Gatorade","price":"2.25"}],"total":"9.00"  }`), invalidReceipt)

		if err != nil {
			t.Fatalf("could not unmarshal invalidReceipt: %v", err)
		}
		isValid := receiptprocessor.ValidateReceipt(invalidReceipt)
		if isValid {
			t.Errorf("expected invalid receipt, got valid")
		}
	})

	t.Run("ValidateReceiptInvalidDate", func(t *testing.T) {
		// test an invalid receipt
		var invalidReceipt = &pb.Receipt{}
		err := protojson.Unmarshal([]byte(`{"retailer":"Target","purchaseDate":"2022-17-01","purchaseTime":"13:01","items":[{"shortDescription":"Mountain Dew 12PK","price":"6.49"},{"shortDescription":"Emils Cheese Pizza","price":"12.25"},{"shortDescription":"Knorr Creamy Chicken","price":"1.26"},{"shortDescription":"Doritos Nacho Cheese","price":"3.35"},{"shortDescription":"   Klarbrunn 12-PK 12 FL OZ  ","price":"12.00"}],"total":"35.35"}`), invalidReceipt)

		if err != nil {
			t.Fatalf("could not unmarshal invalidReceipt: %v", err)
		}
		isValid := receiptprocessor.ValidateReceipt(invalidReceipt)
		if isValid {
			t.Errorf("expected invalid receipt, got valid")
		}
	})
}
