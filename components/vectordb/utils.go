package vectordb

// Float32s converts a Vector ([]float64) to []float32.
// This is necessary because ChromeM uses float32 for vector operations,
// while our interface uses float64 for broader compatibility.
func Float32s(v []float64) []float32 {
	result := make([]float32, len(v))
	for i, val := range v {
		result[i] = float32(val)
	}
	return result
}

func Float64s(v []float32) []float64 {
	result := make([]float64, len(v))
	for i, val := range v {
		result[i] = float64(val)
	}
	return result
}
