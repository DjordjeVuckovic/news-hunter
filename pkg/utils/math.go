package utils

// RoundDecimal rounds a float64 value to the specified number of decimal places.
// For example, RoundDecimal(3.14159, 2) returns 3.14.
func RoundDecimal(value float64, decimals int) float64 {
	pow := 1.0
	for i := 0; i < decimals; i++ {
		pow *= 10
	}

	return float64(int(value*pow+0.5)) / pow
}
