package utils

import "strings"

func StringIsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
