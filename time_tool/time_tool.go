package time_tool

import (
	"fmt"
	"time"
)

func PrintDefault() {
	now := time.Now()
	fmt.Println("datetime:\t", now)
	fmt.Println("weekday:\t", now.Local().Weekday())
	fmt.Println("timestamp:\t", now.Unix())
}

func PrintTimestampToString(timestamp int64) {
	timeStr := time.Unix(timestamp, 0)
	fmt.Println(timeStr)
}

func PrintStringToTimestamp(timeStr string) {
	supportedFormat := []string{time.Layout, time.ANSIC, time.UnixDate, time.RubyDate,
		time.RFC822, time.RFC822Z, time.RFC850, time.RFC1123, time.RFC1123Z,
		time.RFC3339, time.RFC3339Nano, time.Kitchen, time.DateTime, time.DateOnly,
		time.TimeOnly, "2021-1-2", "2021-01-2", "2021-1-02"}

	for _, format := range supportedFormat {
		timeVar, err := time.Parse(format, timeStr)
		if err == nil {
			fmt.Println(timeVar.Unix())
			return
		}
	}
	fmt.Println("Error time format")
}

func PrintCurrentTimestamp() {
	fmt.Println("printCurrentTimestamp")
}
