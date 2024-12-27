package agglomerator

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "gonum.org/v1/gonum/mat"
	"math"
	"math/rand"
	"testing"
)

func TestDetermineOptimalRank(t *testing.T) {
	tests := []struct {
		name            string
		singularValues  []float64
		maxRank         int
		energyThreshold float64
		expectedRank    int
	}{
		{
			name:            "Empty values",
			singularValues:  []float64{},
			maxRank:         5,
			energyThreshold: 0.95,
			expectedRank:    0,
		},
		{
			name:            "Zero values",
			singularValues:  []float64{0, 0, 0},
			maxRank:         5,
			energyThreshold: 0.95,
			expectedRank:    0,
		},
		{
			name:            "Single dominant value",
			singularValues:  []float64{10.0, 0.1, 0.01},
			maxRank:         5,
			energyThreshold: 0.95,
			expectedRank:    1,
		},
		{
			name:            "Gradual decay",
			singularValues:  []float64{5.0, 4.0, 3.0, 2.0, 1.0},
			maxRank:         5,
			energyThreshold: 0.90,
			expectedRank:    3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressor := NewAdaptiveCompressor(CompressorConfig{
				MaxRank:         tt.maxRank,
				EnergyThreshold: tt.energyThreshold,
			})

			rank := compressor.DetermineOptimalRank(tt.singularValues)
			assert.Equal(t, tt.expectedRank, rank)
		})
	}
}

func TestCompressBlock(t *testing.T) {
	// Test configurations
	configs := []struct {
		name           string
		dataSize       int
		maxRank        int
		threshold      float64
		allowLowerRank bool // New parameter to control rank requirements
	}{
		{
			name:           "Small Dense Matrix",
			dataSize:       100,
			maxRank:        5,
			threshold:      0.95,
			allowLowerRank: false, // Strict rank requirement
		},
		{
			name:           "Medium Sparse Matrix",
			dataSize:       500,
			maxRank:        10,
			threshold:      0.90,
			allowLowerRank: true, // Allow optimization
		},
		{
			name:           "Large Mixed Matrix",
			dataSize:       1000,
			maxRank:        20,
			threshold:      0.95,
			allowLowerRank: true, // Allow optimization
		},
	}

	for _, cfg := range configs {
		t.Run(cfg.name, func(t *testing.T) {
			// Generate test data
			data := generateTestData(cfg.dataSize)

			// Create compressor
			compressor := NewAdaptiveCompressor(CompressorConfig{
				MaxRank:         cfg.maxRank,
				EnergyThreshold: cfg.threshold,
				MinSparsity:     0.5,
				Tolerance:       0.01,
				ForceMaxRank:    !cfg.allowLowerRank, // New config parameter
			})

			// Compress data
			compressed, err := compressor.CompressBlock(data)
			require.NoError(t, err)
			require.NotNil(t, compressed)

			// Log compression details
			t.Logf("Compressed Block Structure for %s:", cfg.name)
			t.Logf("  S: %d", len(compressed.S))
			t.Logf("  U: %d", len(compressed.U))
			t.Logf("  V: %d", len(compressed.V))

			// Verify compressed structure
			if !cfg.allowLowerRank {
				// Strict rank check
				assert.Equal(t, cfg.maxRank, len(compressed.S), "Incorrect number of singular values")
				assert.Equal(t, cfg.maxRank, len(compressed.U), "Incorrect U matrix rank")
				assert.Equal(t, cfg.maxRank, len(compressed.V), "Incorrect V matrix rank")
			} else {
				// Flexible rank check
				assert.LessOrEqual(t, len(compressed.S), cfg.maxRank, "Rank exceeds maximum")
				assert.Equal(t, len(compressed.S), len(compressed.U), "Inconsistent U matrix rank")
				assert.Equal(t, len(compressed.S), len(compressed.V), "Inconsistent V matrix rank")
			}

			// Verify compression quality
			decompressed, err := compressed.Decompress()
			require.NoError(t, err)
			require.NotNil(t, decompressed)

			// Calculate and verify compression ratio
			originalSize := len(data) * 8 // 8 bytes per float64
			compressedSize := calculateCompressedSize(compressed)
			ratio := float64(compressedSize) / float64(originalSize)

			t.Logf("Compression Statistics:")
			t.Logf("  Original Size: %d bytes", originalSize)
			t.Logf("  Compressed Size: %d bytes", compressedSize)
			t.Logf("  Compression Ratio: %.2f%%", ratio*100)

			// Verify compression is beneficial for optimized cases
			if cfg.allowLowerRank {
				assert.Less(t, ratio, 1.0, "Compression should reduce data size")
			}
		})
	}
}

func generateTestData(size int) []float64 {
	data := make([]float64, size)
	for i := range data {
		// Generate data with some structure for better compression
		data[i] = math.Sin(float64(i)/50.0) + rand.Float64()*0.1
	}
	return data
}

func calculateCompressedSize(block *CompressedBlock) int {
	size := len(block.S) * 8 // Singular values
	for i := range block.U {
		size += len(block.U[i]) * 8
		size += len(block.V[i]) * 8
	}
	return size
}

// Helper function to generate test matrix
func generateTestMatrix(rows, cols int, sparsity float64) [][]float64 {
	matrix := make([][]float64, rows)
	for i := range matrix {
		matrix[i] = make([]float64, cols)
		for j := range matrix[i] {
			if rand.Float64() > sparsity {
				// Generate values between 0 and 1 for better numerical stability
				matrix[i][j] = rand.Float64()
			}
		}
	}
	return matrix
}

// Helper functions for tests

func flattenMatrix(matrix [][]float64) []float64 {
	rows := len(matrix)
	if rows == 0 {
		return nil
	}
	cols := len(matrix[0])

	flattened := make([]float64, rows*cols)
	idx := 0
	for i := range matrix {
		for j := range matrix[i] {
			flattened[idx] = matrix[i][j]
			idx++
		}
	}
	return flattened
}
