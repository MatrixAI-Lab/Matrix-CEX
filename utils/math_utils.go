package utils

func Min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func Average(a, b float64) float64 {
	return (a + b) / 2
}