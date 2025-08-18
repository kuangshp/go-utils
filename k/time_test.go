package k

import (
	"fmt"
	"testing"
)

func TestCheckDateTime(t *testing.T) {
	fmt.Println(CheckDateTime("2024-12-20"))
	fmt.Println(CheckDateTime("2024-12-20 12:12:12"))
}

func TestDateStrToTime(t *testing.T) {
	fmt.Println(DateStrToTime("2024-12-20"))
}

func TestDateTimeStrToTime(t *testing.T) {
	fmt.Println(DateTimeStrToTime("2024-12-20"))
	fmt.Println(DateTimeStrToTime("2024-12-20 12:12:12"))
}
