package results

import (
	"fmt"
	"strings"
)

const FINLength = 7

func NormalizeFIN(value string) (string, error) {
	fin := strings.ToUpper(strings.TrimSpace(value))
	if len(fin) != FINLength {
		return "", fmt.Errorf("invalid FIN length")
	}
	for _, character := range fin {
		if (character < 'A' || character > 'Z') && (character < '0' || character > '9') {
			return "", fmt.Errorf("invalid FIN characters")
		}
	}
	return fin, nil
}
