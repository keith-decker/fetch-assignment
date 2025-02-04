package receiptprocessor

import (
	"testing"

	"github.com/keith-decker/fetch-assignment/pb"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestReceiptProcessorInternal(t *testing.T) {
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
		expected := 6
		if result := rule1.Process(receipt1); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		expected = 0
		if result := rule2.Process(receipt1); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		expected = 0
		if result := rule3.Process(receipt1); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		expected = 0
		if result := rule3.Process(receipt1); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		expected = 10
		if result := rule4.Process(receipt1); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		expected = 6
		if result := rule5.Process(receipt1); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		expected = 6
		if result := rule6.Process(receipt1); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		expected = 0
		if result := rule7.Process(receipt1); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
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
		expected := 14
		if result := rule1.Process(receipt2); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		expected = 50
		if result := rule2.Process(receipt2); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		expected = 25
		if result := rule3.Process(receipt2); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		expected = 10
		if result := rule4.Process(receipt2); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		expected = 0
		if result := rule5.Process(receipt2); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		expected = 0
		if result := rule6.Process(receipt2); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		expected = 10
		if result := rule7.Process(receipt2); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

	})

	t.Run("ProcessAdditionalSanityChecks", func(t *testing.T) {
		// Test Rule 4 with 2 items
		// Total Points: 5
		expected := 5
		if result := rule4.Process(&pb.Receipt{Items: []*pb.Item{{}, {}}}); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		// Test Rule 4 with 3 items
		// Total Points: 5
		expected = 5
		if result := rule4.Process(&pb.Receipt{Items: []*pb.Item{{}, {}, {}}}); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		// Test Rule 4 with 4 items
		// Total Points: 10
		expected = 10
		if result := rule4.Process(&pb.Receipt{Items: []*pb.Item{{}, {}, {}, {}}}); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		// Test Rule 5 Line Item
		// "Emils Cheese Pizza" is 18 characters (a multiple of 3)
		// item price of 12.25 * 0.2 = 2.45, rounded up is 3 points
		expected = 3
		lineItem := &pb.Item{ShortDescription: "Emils Cheese Pizza", Price: "12.25"}
		if result := processRule5LineItem(lineItem); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		// Test Rule 5 Line Item for Trimmed Description
		// "Klarbrunn 12-PK 12 FL OZ" is 24 characters (a multiple of 3)
		// item price of 12.00 * 0.2 = 2.4, rounded up is 3 points
		expected = 3
		lineItem = &pb.Item{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"}
		if result := processRule5LineItem(lineItem); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		// Test Rule 5 Line Item for Non-Multiple of 3
		// "Gatorade" is 7 characters (not a multiple of 3)
		expected = 0
		lineItem = &pb.Item{ShortDescription: "Gatorade", Price: "2.25"}
		if result := processRule5LineItem(lineItem); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		// Test Rule 6 for Odd Day
		// Total Points: 6
		expected = 6
		if result := rule6.Process(&pb.Receipt{PurchaseDate: "2025-01-21"}); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		// Test Rule 6 for Even Day
		// Total Points: 0
		expected = 0
		if result := rule6.Process(&pb.Receipt{PurchaseDate: "2025-01-22"}); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		// Test Rule 7 for After 2:00pm and Before 4:00pm
		// Total Points: 10
		expected = 10
		if result := rule7.Process(&pb.Receipt{PurchaseTime: "14:33"}); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}

		// Test Rule 7 for Before 2:00pm
		// Total Points: 0
		expected = 0
		if result := rule7.Process(&pb.Receipt{PurchaseTime: "13:59"}); result != expected {
			t.Errorf("expected %d, got %d", expected, result)
		}
	})
}
