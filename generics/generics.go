package generics

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const GoDateFormat = "2006-01-02"
const DateDisplayFormat = "02 Jan 2006"
const GoTimeFormat = "15:04:05"
const ScrapeDateFormat = "02.01.2006"

func Btoi(b bool) int {
	if b {
		return 1
	}

	return 0
}

var DayNames = map[int]string{
	0: "Duminica",
	1: "Luni",
	2: "Marti",
	3: "Miercuri",
	4: "Joi",
	5: "Vineri",
	6: "Sambata",
}

var DrawDays = map[int]string{
	0: "Duminica",
	4: "Joi",
}

var RomanNumbers = map[int]string{
	1: "I",
	2: "II",
	3: "III",
	4: "IV",
	5: "V",
	6: "VI",
	7: "VII",
	8: "VIII",
}

func IndexOf[T any](slice []T, predicate func(T) bool) int {
	for i, item := range slice {
		if predicate(item) {
			return i
		}
	}

	return -1
}

func FindFirst[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, item := range slice {
		if predicate(item) {
			return item, true
		}
	}

	var noT T
	return noT, false
}

var supportedDateFormats = []string{
	"2006-01-02",
	"2006/01/02",
	"2006.01.02",

	"02-01-2006",
	"02/01/2006",
	"02.01.2006",

	"2006-Jan-02",
	"2006/Jan/02",
	"2006.Jan.02",

	"02-Jan-2006",
	"02/Jan/2006",
	"02.Jan.2006",
}

func TryParseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	for _, format := range supportedDateFormats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported date format: %s", dateStr)
}

func SerializeIgnoreError(data any) string {
	dataStr, _ := json.Marshal(data)
	return string(dataStr)
}
