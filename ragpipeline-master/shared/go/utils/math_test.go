package utils

import (
	"math"
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a, b []float64
		want float64
	}{
		{
			name: "identical vectors",
			a:    []float64{1, 2, 3},
			b:    []float64{1, 2, 3},
			want: 1.0,
		},
		{
			name: "orthogonal vectors",
			a:    []float64{1, 0},
			b:    []float64{0, 1},
			want: 0.0,
		},
		{
			name: "opposite vectors",
			a:    []float64{1, 2, 3},
			b:    []float64{-1, -2, -3},
			want: -1.0,
		},
		{
			name: "different lengths",
			a:    []float64{1, 0, 0},
			b:    []float64{0, 1, 0},
			want: 0.0,
		},
		{
			name: "partial similarity",
			a:    []float64{1, 1, 1},
			b:    []float64{1, 1, 0},
			want: 0.8164965809277261, // 2/sqrt(6)
		},
		{
			name: "zero vectors",
			a:    []float64{0, 0, 0},
			b:    []float64{0, 0, 0},
			want: 0.0,
		},
		{
			name: "empty vectors",
			a:    []float64{},
			b:    []float64{},
			want: 0.0,
		},
		{
			name: "different dimensions",
			a:    []float64{1, 2},
			b:    []float64{1, 2, 3},
			want: 0.0,
		},
		{
			name: "one zero one non-zero",
			a:    []float64{0, 0, 0},
			b:    []float64{1, 2, 3},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CosineSimilarity(tt.a, tt.b)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("CosineSimilarity(%v, %v) = %.15f, want %.15f", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestDotProduct(t *testing.T) {
	tests := []struct {
		name string
		a, b []float64
		want float64
	}{
		{
			name: "basic dot product",
			a:    []float64{1, 2, 3},
			b:    []float64{4, 5, 6},
			want: 32, // 1*4 + 2*5 + 3*6
		},
		{
			name: "orthogonal vectors",
			a:    []float64{1, 0},
			b:    []float64{0, 1},
			want: 0,
		},
		{
			name: "zero vectors",
			a:    []float64{0, 0},
			b:    []float64{0, 0},
			want: 0,
		},
		{
			name: "empty vectors",
			a:    []float64{},
			b:    []float64{},
			want: 0.0,
		},
		{
			name: "different dimensions",
			a:    []float64{1, 2},
			b:    []float64{1, 2, 3},
			want: 0.0,
		},
		{
			name: "negative values",
			a:    []float64{-1, -2},
			b:    []float64{3, 4},
			want: -11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DotProduct(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("DotProduct(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name   string
		vector []float64
		want   []float64
	}{
		{
			name:   "unit vector",
			vector: []float64{1, 0, 0},
			want:   []float64{1, 0, 0},
		},
		{
			name:   "simple normalization",
			vector: []float64{3, 4},
			want:   []float64{0.6, 0.8},
		},
		{
			name:   "zero vector",
			vector: []float64{0, 0, 0},
			want:   []float64{0, 0, 0},
		},
		{
			name:   "negative values",
			vector: []float64{-1, -1},
			want:   []float64{-0.7071067811865475, -0.7071067811865475},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.vector)
			for i := range got {
				if math.Abs(got[i]-tt.want[i]) > 1e-9 {
					t.Errorf("Normalize(%v)[%d] = %v, want %v", tt.vector, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestMean(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"positive values", []float64{1, 2, 3, 4, 5}, 3.0},
		{"negative values", []float64{-1, -2, -3}, -2.0},
		{"mixed values", []float64{-1, 0, 1}, 0.0},
		{"single value", []float64{42}, 42.0},
		{"empty slice", []float64{}, 0.0},
		{"identical values", []float64{5, 5, 5}, 5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Mean(tt.values)
			if got != tt.want {
				t.Errorf("Mean(%v) = %v, want %v", tt.values, got, tt.want)
			}
		})
	}
}

func TestVariance(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"sample variance", []float64{2, 4, 4, 4, 5, 5, 7, 9}, 4.571428571428571},
		{"identical values", []float64{5, 5, 5}, 0.0},
		{"two values", []float64{1, 3}, 2.0},
		{"single value", []float64{42}, 0.0},
		{"empty slice", []float64{}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Variance(tt.values)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("Variance(%v) = %v, want %v", tt.values, got, tt.want)
			}
		})
	}
}

func TestStdDev(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"standard deviation", []float64{2, 4, 4, 4, 5, 5, 7, 9}, 2.138089935299395},
		{"identical values", []float64{5, 5, 5}, 0.0},
		{"empty slice", []float64{}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StdDev(tt.values)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("StdDev(%v) = %v, want %v", tt.values, got, tt.want)
			}
		})
	}
}

func TestPercentile(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		p      float64
		want   float64
	}{
		{
			name:   "median",
			values: []float64{1, 2, 3, 4, 5},
			p:      50,
			want:   3.0,
		},
		{
			name:   "25th percentile",
			values: []float64{1, 2, 3, 4, 5},
			p:      25,
			want:   2.0,
		},
		{
			name:   "75th percentile",
			values: []float64{1, 2, 3, 4, 5},
			p:      75,
			want:   4.0,
		},
		{
			name:   "min (0th percentile)",
			values: []float64{1, 2, 3, 4, 5},
			p:      0,
			want:   1.0,
		},
		{
			name:   "max (100th percentile)",
			values: []float64{1, 2, 3, 4, 5},
			p:      100,
			want:   5.0,
		},
		{
			name:   "interpolation",
			values: []float64{1, 2, 3, 4},
			p:      25,
			want:   1.75,
		},
		{
			name:   "single value",
			values: []float64{42},
			p:      50,
			want:   42.0,
		},
		{
			name:   "empty slice",
			values: []float64{},
			p:      50,
			want:   0.0,
		},
		{
			name:   "unsorted input",
			values: []float64{5, 1, 4, 2, 3},
			p:      50,
			want:   3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Percentile(tt.values, tt.p)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("Percentile(%v, %.0f) = %v, want %v", tt.values, tt.p, got, tt.want)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"positive values", []float64{3, 1, 4, 1, 5}, 1},
		{"negative values", []float64{-1, -5, -2}, -5},
		{"single value", []float64{42}, 42},
		{"empty slice", []float64{}, 0.0},
		{"float values", []float64{1.5, 0.3, 2.7}, 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Min(tt.values)
			if got != tt.want {
				t.Errorf("Min(%v) = %v, want %v", tt.values, got, tt.want)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"positive values", []float64{3, 1, 4, 1, 5}, 5},
		{"negative values", []float64{-1, -5, -2}, -1},
		{"single value", []float64{42}, 42},
		{"empty slice", []float64{}, 0.0},
		{"float values", []float64{1.5, 0.3, 2.7}, 2.7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Max(tt.values)
			if got != tt.want {
				t.Errorf("Max(%v) = %v, want %v", tt.values, got, tt.want)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name        string
		value, min, max float64
		want        float64
	}{
		{"within range", 5, 0, 10, 5},
		{"below min", -5, 0, 10, 0},
		{"above max", 15, 0, 10, 10},
		{"at min", 0, 0, 10, 0},
		{"at max", 10, 0, 10, 10},
		{"negative range", -5, -10, 0, -5},
		{"float values", 0.5, 0, 1, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Clamp(tt.value, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("Clamp(%v, %v, %v) = %v, want %v", tt.value, tt.min, tt.max, got, tt.want)
			}
		})
	}
}
