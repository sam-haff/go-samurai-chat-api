package testutils

import (
	"fmt"
	"testing"
)

const InvalidCommResponseFormatMessage = "Invalid comm response format"

func InvalidResponseHttpStatusCodeMessage(expected int, actual int) string {
	return fmt.Sprintf("Invalid http response status code, expected %d, got %d", expected, actual)
}
func InvalidResponseCommStatusCodeMessage(expected int, actual int) string {
	return fmt.Sprintf("Invalid response comm status code, expected %d, got %d", expected, actual)
}
func ValidateResponseHttpStatusCode(t *testing.T, expected int, actual int) {
	if expected != actual {
		t.Errorf("Invalid http response status code, expected %d, got %d", expected, actual)
	}
}
func ValidateResponseCommStatusCode(t *testing.T, expected int, actual int) {
	if expected != actual {
		t.Errorf("Invalid response comm status code, expected %d, got %d", expected, actual)
	}
}
