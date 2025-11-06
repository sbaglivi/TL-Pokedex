package utils

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var re = regexp.MustCompile(`\s+`)

func RemoveWhitespace(s string) string {
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

func GetPort() (int, error) {
	defaultPort := 3000
	portStr := os.Getenv("PORT")
	if portStr == "" {
		return defaultPort, nil
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("cannot parse [%s] as int: %w", portStr, err)
	}

	return port, nil
}
