package testutils

import (
	"fmt"
	"runtime"
	"testing"
)

func getCallingFuncName() string {
	pc, _, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		return details.Name()
	}
	return ""
}

func PrintTestName(t *testing.T, testName string) {
	t.Log("=== RUN " + t.Name() + "/" + testName)
}

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
