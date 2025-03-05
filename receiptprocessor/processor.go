package receiptprocessor

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/keith-decker/fetch-assignment/kvstore"
	"github.com/keith-decker/fetch-assignment/pb"
)

type pointRuleInterface interface {
	Process(*pb.Receipt) int
	isEnabled() bool
}

type pointRule struct {
	processFunc   func(*pb.Receipt) int
	isEnabledFunc func() bool
}

func (p *pointRule) Process(receipt *pb.Receipt) int {
	return p.processFunc(receipt)
}

func (p pointRule) isEnabled() bool {
	return p.isEnabledFunc()
}

func ProcessReceipt(receipt *pb.Receipt) string {
	// Generate an ID for this receipt
	// @TODO: create a hash of the receipt to prevent duplicates. Date/Time + Store ID + Total?
	id := uuid.New().String()

	// store the receipt in the KV store, kick off the processing and return the ID
	kv := kvstore.New()
	// set the score to -1 to indicate that the receipt is being processed
	kv.Set(fmt.Sprintf("receipt-%s", id), "-1")

	processReceipt(id, receipt)

	return id
}

// dirty. I would refactor this to be more testable and extensible. I would also move the validation to a separate functions.
func ValidateReceipt(receipt *pb.Receipt) bool {
	errors := []string{}
	// check retailer against regex "^[\\w\\s\\-&]+$"
	matched, err := regexp.MatchString("^[\\w\\s\\-&]+$", receipt.Retailer)
	if err != nil {
		fmt.Printf("Error validating retailer regex: %v\n", err)
		errors = append(errors, fmt.Sprintf("Error validating retailer: %v", receipt.Retailer))
	}
	if !matched {
		errors = append(errors, fmt.Sprintf("Retailer name is invalid: %v", receipt.Retailer))
	}

	// Convert the date into a date object (Ideally this is done in the proto)
	layout := "2006-01-02"
	_, err = time.Parse(layout, receipt.PurchaseDate)
	if err != nil {
		fmt.Printf("Error validating purchase date regex: %v\n", err)
		errors = append(errors, fmt.Sprintf("Error validating purchase date %v", receipt.PurchaseDate))
	}

	// check purchase time against regex "^[0-9]{2}:[0-9]{2}$"
	matched, err = regexp.MatchString("^[0-9]{2}:[0-9]{2}$", receipt.PurchaseTime)
	if err != nil {
		fmt.Printf("Error validating purchase time regex: %v\n", err)
		errors = append(errors, fmt.Sprintf("Error validating purchase time: %v", receipt.PurchaseTime))
	}
	if !matched {
		errors = append(errors, fmt.Sprintf("Purchase time is invalid: %v", receipt.PurchaseTime))
	}
	// check total against regex "^\\d+\\.\\d{2}$"
	matched, err = regexp.MatchString("^\\d+\\.\\d{2}$", receipt.Total)
	if err != nil {
		fmt.Printf("Error validating total regex: %v\n", err)
		errors = append(errors, fmt.Sprintf("Error validating total: %v", receipt.Total))
	}
	if !matched {
		errors = append(errors, fmt.Sprintf("Total is invalid: %v", receipt.Total))
	}
	// check items
	shortDescRegex, err := regexp.Compile(`^[\w\s\-]+$`)
	if err != nil {
		fmt.Printf("Error compiling short description regex: %v\n", err)
		errors = append(errors, "Error compiling short description regex")
	}
	priceRegex, err := regexp.Compile(`^\d+\.\d{2}$`)
	if err != nil {
		fmt.Printf("Error compiling price regex: %v\n", err)
		errors = append(errors, "Error compiling price regex")
	}

	if len(receipt.Items) == 0 {
		errors = append(errors, "No items on receipt")
	}

	for _, item := range receipt.Items {
		// check short description against regex "^[\\w\\s\\-]+$"
		matched = shortDescRegex.MatchString(item.ShortDescription)
		if !matched {
			errors = append(errors, fmt.Sprintf("Item short description is invalid: %v", item.ShortDescription))
		}
		// check price against regex "^[\\w\\s\\-]+$"
		matched = priceRegex.MatchString(item.Price)
		if !matched {
			errors = append(errors, fmt.Sprintf("Item price is invalid: %v", item.Price))
		}
	}
	if len(errors) > 0 {
		fmt.Printf("Receipt is invalid: %v\n", errors)
		return false
	}
	return true
}

func processReceipt(id string, receipt *pb.Receipt) {
	// Process the receipt
	kv := kvstore.New()
	kv.Set(fmt.Sprintf("receipt-%s", id), "99")

	totalScore := tallyScore(receipt)

	kv.Set(fmt.Sprintf("receipt-%s", id), fmt.Sprintf("%d", totalScore))

}

// TallyScore takes a receipt and processes it against the rules to determine the total score.
func tallyScore(receipt *pb.Receipt) int {
	rules := defaultRules()
	totalScore := 0
	for _, rule := range rules {
		if !rule.isEnabled() {
			continue
		}
		totalScore += rule.Process(receipt)
	}
	return totalScore
}

// ------ These rules could be setup as individual modules/imports. ------

func newPointRule(processFunc func(*pb.Receipt) int) *pointRule {
	return &pointRule{
		processFunc: processFunc,
		isEnabledFunc: func() bool {
			return true
		},
	}
}

var rule1 = newPointRule(func(receipt *pb.Receipt) int {
	// One point for every alphanumeric character in the retailer name.
	points := 0
	toTest := strings.ToUpper(receipt.Retailer)
	for _, char := range toTest {
		if (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			points++
		}
	}
	return points
})

var rule2 = newPointRule(func(receipt *pb.Receipt) int {
	// 50 points if the total is a round dollar amount with no cents.
	// convert the string to a float
	receiptTotal, err := strconv.ParseFloat(receipt.Total, 64)
	if err != nil {
		fmt.Printf("Error converting receipt total to float (%s), returning 0\n", receipt.Total)
		return 0
	}
	// check if the float is a whole number
	if receiptTotal == float64(int(receiptTotal)) {
		return 50
	}
	return 0
})

var rule3 = newPointRule(func(receipt *pb.Receipt) int {
	// 25 points if the total is a multiple of 0.25.
	receiptTotal, err := strconv.ParseFloat(receipt.Total, 64)
	if err != nil {
		fmt.Printf("Error converting receipt total to float (%s), returning 0\n", receipt.Total)
		return 0
	}
	decimalAmount := receiptTotal - float64(int(receiptTotal))
	if int(decimalAmount*100)%25 == 0 {
		return 25
	}
	return 0
})

var rule4 = newPointRule(func(receipt *pb.Receipt) int {
	// 5 points for every two items on the receipt.
	return int(len(receipt.Items)/2) * 5
})

var rule5 = newPointRule(func(receipt *pb.Receipt) int {
	// If the trimmed length of the item description is a multiple of 3,
	// multiply the price by 0.2 and round up to the nearest integer.
	// The result is the number of points earned.
	points := 0
	for _, item := range receipt.Items {
		points += processRule5LineItem(item)
	}
	return points
})

func processRule5LineItem(item *pb.Item) int {
	// If the trimmed length of the item description is a multiple of 3,
	// multiply the price by 0.2 and round up to the nearest integer.
	// The result is the number of points earned.
	trimmedDescription := strings.TrimSpace(item.ShortDescription)
	if len(trimmedDescription)%3 == 0 {
		price, err := strconv.ParseFloat(item.Price, 64)
		if err != nil {
			fmt.Printf("Error converting item price to float (%s), returning 0\n", item.Price)
			return 0
		}
		return int(math.Ceil(price * 0.2))
	}
	return 0
}

var rule6 = newPointRule(func(receipt *pb.Receipt) int {
	// 6 points if the day in the purchase date is odd.

	// Convert the date into a date object (Ideally this is done in the proto)
	layout := "2006-01-02"
	date, err := time.Parse(layout, receipt.PurchaseDate)
	if err != nil {
		fmt.Printf("Error converting purchase date to time (%s), returning 0\n", receipt.PurchaseDate)
		return 0
	}

	if date.Day()%2 != 0 {
		return 6
	}
	return 0
})

var rule7 = newPointRule(func(receipt *pb.Receipt) int {
	// 10 points if the time of purchase is after 2:00pm and before 4:00pm.

	// Convert the date into a date object (Ideally this is done in the proto)
	layout := "15:04"
	date, err := time.Parse(layout, receipt.PurchaseTime)
	if err != nil {
		fmt.Printf("Error converting purchase date to time (%s), returning 0\n", receipt.PurchaseTime)
		return 0
	}
	if date.Hour() >= 14 && date.Hour() < 16 {
		return 10
	}
	return 0
})

func defaultRules() []pointRuleInterface {
	return []pointRuleInterface{rule1, rule2, rule3, rule4, rule5, rule6, rule7}
}
