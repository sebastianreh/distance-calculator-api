package strings

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	Empty    = ""
	minParts = 2
)

func IsEmpty(value string) bool {
	return value == Empty
}

func ErrorConcat(err error, layer, origin string) (message, layerOrigin string) {
	return err.Error(), fmt.Sprintf("%s.%s", layer, origin)
}

func TimeToInt(timeStr string) (int, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) < minParts {
		return 0, fmt.Errorf("invalid time format")
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	return hours*100 + minutes, nil
}
