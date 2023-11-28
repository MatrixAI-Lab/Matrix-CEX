package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

func Min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func Average(a, b float64) float64 {
	return (a + b) / 2
}

func FormatFloat(f float64) (float64, error) {
	s := fmt.Sprintf("%f", f)
	if strings.Contains(s, ".") && len(s[strings.Index(s, ".")+1:]) > 4 {
		s = fmt.Sprintf("%.4f", f)
	}
	return strconv.ParseFloat(s, 64)
}

func GenerateOrderID() string {
    randBytes := make([]byte, 8)
    _, err := rand.Read(randBytes)
    if err != nil {
        panic(err)
    }
    return "order" + hex.EncodeToString(randBytes)
}