// Package utils provides shared utility functions for statistics and vector math operations.
package utils

import (
	"cmp"
	"math"
	"slices"
)

// CosineSimilarity calculates the cosine similarity between two vectors.
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	if len(a) == 0 {
		return 0.0
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// DotProduct calculates the dot product of two vectors.
func DotProduct(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	result := 0.0
	for i := 0; i < len(a); i++ {
		result += a[i] * b[i]
	}
	return result
}

// Normalize normalizes a vector to unit length.
func Normalize(vector []float64) []float64 {
	norm := 0.0
	for _, v := range vector {
		norm += v * v
	}
	norm = math.Sqrt(norm)

	if norm == 0 {
		return vector
	}

	result := make([]float64, len(vector))
	for i, v := range vector {
		result[i] = v / norm
	}
	return result
}

// Mean calculates the arithmetic mean of values.
func Mean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// Variance calculates the variance of values.
func Variance(values []float64) float64 {
	if len(values) <= 1 {
		return 0.0
	}
	m := Mean(values)
	sumSquares := 0.0
	for _, v := range values {
		diff := v - m
		sumSquares += diff * diff
	}
	return sumSquares / float64(len(values)-1)
}

// StdDev calculates the standard deviation of values.
func StdDev(values []float64) float64 {
	return math.Sqrt(Variance(values))
}

// Percentile calculates the p-th percentile of values (0-100).
func Percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	// Sort using slices.SortFunc for float64
	sorted := make([]float64, len(values))
	copy(sorted, values)
	slices.SortFunc(sorted, cmp.Compare[float64])

	// Calculate percentile position
	pos := p / 100.0 * float64(len(sorted)-1)
	lower := int(math.Floor(pos))
	upper := int(math.Ceil(pos))

	if lower == upper {
		return sorted[lower]
	}

	// Interpolate
	frac := pos - math.Floor(pos)
	return sorted[lower]*(1-frac) + sorted[upper]*frac
}

// Min returns the minimum value.
func Min(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// Max returns the maximum value.
func Max(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// Clamp clamps a value between min and max.
func Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
