package utils

import (
	"fmt"
	"os"
	"time"
)

const timeLayout = "02-01-2006"

func Fdate(date string) string {
	var parsedDate, err = time.Parse(timeLayout, date)

	if err != nil {
		fmt.Println("TimeError:", err)
		os.Exit(1)
	}
	return parsedDate.Format("Monday, 02 Jan 2006")
}
