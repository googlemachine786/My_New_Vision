package analytics

import (
	"math"
	"testing"
)

func TestCalculateRecallAtK(t *testing.T) {
	tests := []struct {
		name     string
		retrieved []string
		relevant []string
		k        int
		want     float64
	}{
		{
			name:     "perfect recall",
			retrieved: []string{"a", "b", "c"},
			relevant: []string{"a", "b", "c"},
			k:        3,
			want:     1.0,
		},
		{
			name:     "partial recall",
			retrieved: []string{"a", "b", "d"},
			relevant: []string{"a", "b", "c"},
			k:        3,
			want:     2.0 / 3.0,
		},
		{
			name:     "no recall",
			retrieved: []string{"d", "e", "f"},
			relevant: []string{"a", "b", "c"},
			k:        3,
			want:     0.0,
		},
		{
			name:     "recall at k=1",
			retrieved: []string{"a", "b", "c"},
			relevant: []string{"a", "b", "c"},
			k:        1,
			want:     1.0 / 3.0,
		},
		{
			name:     "k exceeds retrieved length",
			retrieved: []string{"a", "b"},
			relevant: []string{"a", "b", "c"},
			k:        5,
			want:     2.0 / 3.0,
		},
		{
			name:     "empty retrieved",
			retrieved: []string{},
			relevant: []string{"a", "b"},
			k:        3,
			want:     0.0,
		},
		{
			name:     "empty relevant",
			retrieved: []string{"a", "b"},
			relevant: []string{},
			k:        3,
			want:     0.0,
		},
		{
			name:     "both empty",
			retrieved: []string{},
			relevant: []string{},
			k:        3,
			want:     0.0,
		},
		{
			name:     "duplicates in retrieved",
			retrieved: []string{"a", "a", "a"},
			relevant: []string{"a"},
			k:        3,
			want:     3.0, // counts each duplicate hit
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateRecallAtK(tt.retrieved, tt.relevant, tt.k)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("CalculateRecallAtK(%v, %v, %d) = %v, want %v",
					tt.retrieved, tt.relevant, tt.k, got, tt.want)
			}
		})
	}
}

func TestCalculateMRR(t *testing.T) {
	tests := []struct {
		name     string
		retrieved []string
		relevant []string
		want     float64
	}{
		{
			name:     "first result relevant",
			retrieved: []string{"a", "b", "c"},
			relevant: []string{"a", "b"},
			want:     1.0,
		},
		{
			name:     "second result relevant",
			retrieved: []string{"x", "a", "b"},
			relevant: []string{"a", "b"},
			want:     0.5,
		},
		{
			name:     "third result relevant",
			retrieved: []string{"x", "y", "a"},
			relevant: []string{"a"},
			want:     1.0 / 3.0,
		},
		{
			name:     "no relevant results",
			retrieved: []string{"x", "y", "z"},
			relevant: []string{"a", "b"},
			want:     0.0,
		},
		{
			name:     "empty retrieved",
			retrieved: []string{},
			relevant: []string{"a"},
			want:     0.0,
		},
		{
			name:     "empty relevant",
			retrieved: []string{"a", "b"},
			relevant: []string{},
			want:     0.0,
		},
		{
			name:     "single relevant at end",
			retrieved: []string{"d", "c", "b", "a"},
			relevant: []string{"a"},
			want:     0.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateMRR(tt.retrieved, tt.relevant)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("CalculateMRR(%v, %v) = %v, want %v",
					tt.retrieved, tt.relevant, got, tt.want)
			}
		})
	}
}

func TestCalculateNDCG(t *testing.T) {
	tests := []struct {
		name     string
		retrieved []string
		relevant []string
		want     float64
	}{
		{
			name:     "perfect ranking",
			retrieved: []string{"a", "b", "c"},
			relevant: []string{"a", "b", "c"},
			want:     1.0,
		},
		{
			name:     "partial ranking",
			retrieved: []string{"x", "a", "b"},
			relevant: []string{"a", "b"},
			want:     0.6934264036172708,
		},
		{
			name:     "no relevant results",
			retrieved: []string{"x", "y", "z"},
			relevant: []string{"a", "b"},
			want:     0.0,
		},
		{
			name:     "empty retrieved",
			retrieved: []string{},
			relevant: []string{"a"},
			want:     0.0,
		},
		{
			name:     "empty relevant",
			retrieved: []string{"a", "b"},
			relevant: []string{},
			want:     0.0,
		},
		{
			name:     "single relevant first",
			retrieved: []string{"a", "x", "y"},
			relevant: []string{"a"},
			want:     1.0,
		},
		{
			name:     "retrieved longer than relevant",
			retrieved: []string{"a", "b", "x"},
			relevant: []string{"a", "b"},
			want:     1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateNDCG(tt.retrieved, tt.relevant)
			if tt.want == 0.0 || tt.want == 1.0 {
				if got != tt.want {
					t.Errorf("CalculateNDCG(%v, %v) = %v, want %v",
						tt.retrieved, tt.relevant, got, tt.want)
				}
			} else {
				if math.Abs(got-tt.want) > 1e-6 {
					t.Errorf("CalculateNDCG(%v, %v) = %v, want %v",
						tt.retrieved, tt.relevant, got, tt.want)
				}
			}
		})
	}
}

func TestEvaluateRetrieval(t *testing.T) {
	// Mock retrieval function
	mockRetrieval := func(query string, grade int, subject string) []string {
		switch query {
		case "photosynthesis":
			return []string{"chunk1", "chunk2", "chunk3"}
		case "cell membrane":
			return []string{"chunk4", "chunk5"}
		default:
			return []string{}
		}
	}

	tests := []struct {
		name    string
		qaPairs []QAPair
		topK    []int
		want    func(*testing.T, *RetrievalMetrics)
	}{
		{
			name: "basic evaluation",
			qaPairs: []QAPair{
				{
					Question:     "photosynthesis",
					GoldChunkIDs: []string{"chunk1", "chunk2"},
				},
				{
					Question:     "cell membrane",
					GoldChunkIDs: []string{"chunk4"},
				},
			},
			topK: []int{1, 3},
			want: func(t *testing.T, m *RetrievalMetrics) {
				t.Helper()
				if m.NumQueries != 2 {
					t.Errorf("NumQueries = %d, want 2", m.NumQueries)
				}
				if m.MRR <= 0 || m.MRR > 1 {
					t.Errorf("MRR = %v, expected between 0 and 1", m.MRR)
				}
				if m.NDCG < 0 || m.NDCG > 1 {
					t.Errorf("NDCG = %v, expected between 0 and 1", m.NDCG)
				}
				if _, ok := m.RecallAtK[1]; !ok {
					t.Error("Missing RecallAtK[1]")
				}
				if _, ok := m.RecallAtK[3]; !ok {
					t.Error("Missing RecallAtK[3]")
				}
			},
		},
		{
			name:    "empty QA pairs",
			qaPairs: []QAPair{},
			topK:    []int{5},
			want: func(t *testing.T, m *RetrievalMetrics) {
				t.Helper()
				if m.NumQueries != 0 {
					t.Errorf("NumQueries = %d, want 0", m.NumQueries)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvaluateRetrieval(tt.qaPairs, mockRetrieval, tt.topK, false)
			if got == nil {
				t.Fatal("EvaluateRetrieval returned nil")
			}
			tt.want(t, got)
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
		{"two values", []float64{1, 3}, 1.4142135623730951},
		{"single value", []float64{42}, 0.0},
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

func BenchmarkCalculateRecallAtK(b *testing.B) {
	retrieved := make([]string, 100)
	relevant := make([]string, 20)
	for i := range retrieved {
		retrieved[i] = "chunk" + itoa(i)
	}
	for i := range relevant {
		relevant[i] = "chunk" + itoa(i*5)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateRecallAtK(retrieved, relevant, 50)
	}
}

func BenchmarkCalculateMRR(b *testing.B) {
	retrieved := make([]string, 100)
	relevant := make([]string, 20)
	for i := range retrieved {
		retrieved[i] = "chunk" + itoa(i)
	}
	for i := range relevant {
		relevant[i] = "chunk" + itoa(i*5)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateMRR(retrieved, relevant)
	}
}

func BenchmarkCalculateNDCG(b *testing.B) {
	retrieved := make([]string, 100)
	relevant := make([]string, 20)
	for i := range retrieved {
		retrieved[i] = "chunk" + itoa(i)
	}
	for i := range relevant {
		relevant[i] = "chunk" + itoa(i*5)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateNDCG(retrieved, relevant)
	}
}

// itoa converts int to string
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
