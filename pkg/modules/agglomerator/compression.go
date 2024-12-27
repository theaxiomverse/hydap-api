// Package agglomerator provides efficient SVD-based data compression with hybrid Rank-1/Adaptive modes.
//
// The package implements a sophisticated compression system that automatically selects between
// Rank-1 and Adaptive compression modes based on data characteristics. It uses Singular Value
// Decomposition (SVD) for dimensionality reduction while maintaining data fidelity.
//
// Basic usage example:
//
//	config := CompressorConfig{
//	    Tolerance:       0.01,
//	    MaxRank:        10,
//	    EnergyThreshold: 0.95,
//	    MinSparsity:    0.5,
//	}
//	compressor := NewAdaptiveCompressor(config)
//
//	// Compress data
//	compressed, err := compressor.CompressBlock(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Decompress data
//	decompressed, err := compressed.Decompress()
//	if err != nil {
//	    log.Fatal(err)
//	}
package agglomerator

import (
	"fmt"
	"gonum.org/v1/gonum/mat"
	"math"
	"sync"
)

type CompressionMode int

const (
	Rank1Mode CompressionMode = iota
	AdaptiveMode
	HybridMode
)

// AdaptiveCompressor implements advanced SVD compression with automatic mode selection.
// It provides thread-safe compression operations and adaptive rank selection based on
// data characteristics.
type AdaptiveCompressor struct {
	mu              sync.RWMutex
	tolerance       float64
	maxRank         int
	energyThreshold float64
	minSparsity     float64
	forceMaxRank    bool
}

// CompressorConfig holds configuration parameters for the compressor.
type CompressorConfig struct {
	// Tolerance controls quantization precision (typical range: 0.001 to 0.1)
	Tolerance float64

	// MaxRank limits the maximum components used (typical range: 10 to 50)
	MaxRank int

	// EnergyThreshold sets minimum energy retention (typical range: 0.9 to 0.99)
	EnergyThreshold float64

	// MinSparsity sets threshold for sparse compression (typical range: 0.3 to 0.7)
	MinSparsity  float64
	ForceMaxRank bool
}

// CompressedBlock represents compressed data and metadata.
// It contains the SVD components and original dimensions needed for reconstruction.
type CompressedBlock struct {
	U            [][]float64     // Left singular vectors
	V            [][]float64     // Right singular vectors
	S            []float64       // Singular values
	OriginalRows int             // Original matrix rows
	OriginalCols int             // Original matrix columns
	OriginalSize int             // Original data size
	Mode         CompressionMode // Compression mode used
}

// NewAdaptiveCompressor creates a new compressor with the given configuration.
// Example usage:
//
//	compressor := NewAdaptiveCompressor(CompressorConfig{
//	    Tolerance:       0.01,
//	    MaxRank:        10,
//	    EnergyThreshold: 0.95,
//	    MinSparsity:    0.5,
//	})
func NewAdaptiveCompressor(config CompressorConfig) *AdaptiveCompressor {
	return &AdaptiveCompressor{
		tolerance:       config.Tolerance,
		maxRank:         config.MaxRank,
		energyThreshold: config.EnergyThreshold,
		minSparsity:     config.MinSparsity,
		forceMaxRank:    config.ForceMaxRank,
	}
}

// CompressBlock compresses the input data using either Rank-1 or Adaptive SVD.
// The mode is automatically selected based on data characteristics.
//
// Parameters:
//   - blockData: Input data as float64 slice
//
// Returns:
//   - *CompressedBlock: Compressed data and metadata
//   - error: Any error that occurred during compression
//
// Example:
//
//	data := []float64{1.0, 2.0, 3.0, 4.0}
//	compressed, err := compressor.CompressBlock(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (ac *AdaptiveCompressor) CompressBlock(blockData []float64) (*CompressedBlock, error) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if len(blockData) == 0 {
		return nil, fmt.Errorf("empty block data")
	}

	// Calculate dimensions
	size := len(blockData)
	rows := int(math.Sqrt(float64(size)))
	cols := size / rows
	if size%rows != 0 {
		cols++
	}

	// Create matrix
	data := make([]float64, rows*cols)
	copy(data, blockData)
	matrix := mat.NewDense(rows, cols, data)

	// Perform SVD
	var svd mat.SVD
	ok := svd.Factorize(matrix, mat.SVDThin)
	if !ok {
		return nil, fmt.Errorf("SVD factorization failed")
	}

	// Get singular values
	singularValues := svd.Values(nil)

	// Determine rank while respecting maxRank
	rank := ac.maxRank
	if ac.forceMaxRank {
		// Use exactly maxRank components (or less if not available)
		if rank > len(singularValues) {
			rank = len(singularValues)
		}
	} else {
		// Calculate optimal rank up to maxRank
		rank = ac.calculateOptimalRank(singularValues, rows, cols, int(float64(size)*0.7))
		if rank > ac.maxRank {
			rank = ac.maxRank
		}
	}

	// Extract matrices
	var u, v mat.Dense
	svd.UTo(&u)
	svd.VTo(&v)

	// Create compressed block
	compressed := &CompressedBlock{
		U:            make([][]float64, rank),
		V:            make([][]float64, rank),
		S:            make([]float64, rank),
		OriginalRows: rows,
		OriginalCols: cols,
		OriginalSize: size,
	}

	// Extract top 'rank' components with exact values
	for i := 0; i < rank; i++ {
		uCol := mat.Col(nil, i, &u)
		vCol := mat.Col(nil, i, &v)

		// Copy without quantization
		compressed.U[i] = make([]float64, len(uCol))
		compressed.V[i] = make([]float64, len(vCol))
		copy(compressed.U[i], uCol)
		copy(compressed.V[i], vCol)
		compressed.S[i] = singularValues[i]
	}

	return compressed, nil
}

func (ac *AdaptiveCompressor) compressRank1(matrix *mat.Dense, svd *mat.SVD) (*CompressedBlock, error) {
	var u, v mat.Dense
	svd.UTo(&u)
	svd.VTo(&v)

	// Extract first singular vectors
	uCol := mat.Col(nil, 0, &u)
	vCol := mat.Col(nil, 0, &v)
	s := svd.Values(nil)[0]

	// Create compressed block
	compressed := &CompressedBlock{
		U:            [][]float64{quantizeVector(uCol, ac.tolerance)},
		V:            [][]float64{quantizeVector(vCol, ac.tolerance)},
		S:            []float64{s},
		OriginalRows: matrix.RawMatrix().Rows,
		OriginalCols: matrix.RawMatrix().Cols,
		OriginalSize: matrix.RawMatrix().Rows * matrix.RawMatrix().Cols,
		Mode:         Rank1Mode,
	}

	return compressed, nil
}

func (ac *AdaptiveCompressor) compressAdaptive(matrix *mat.Dense, svd *mat.SVD, singularValues []float64) (*CompressedBlock, error) {
	targetSize := int(float64(matrix.RawMatrix().Rows*matrix.RawMatrix().Cols) * 0.7)
	rank := ac.calculateOptimalRank(singularValues, matrix.RawMatrix().Rows, matrix.RawMatrix().Cols, targetSize)

	var u, v mat.Dense
	svd.UTo(&u)
	svd.VTo(&v)

	compressed := &CompressedBlock{
		U:            make([][]float64, rank),
		V:            make([][]float64, rank),
		S:            make([]float64, rank),
		OriginalRows: matrix.RawMatrix().Rows,
		OriginalCols: matrix.RawMatrix().Cols,
		OriginalSize: matrix.RawMatrix().Rows * matrix.RawMatrix().Cols,
		Mode:         AdaptiveMode,
	}

	for i := 0; i < rank; i++ {
		uCol := mat.Col(nil, i, &u)
		vCol := mat.Col(nil, i, &v)
		compressed.U[i] = quantizeVector(uCol, ac.tolerance)
		compressed.V[i] = quantizeVector(vCol, ac.tolerance)
		compressed.S[i] = singularValues[i]
	}

	return compressed, nil
}

// DecompressStream handles streaming decompression of multiple blocks
func (ac *AdaptiveCompressor) DecompressStream(blocks []*CompressedBlock) ([]float64, error) {
	if len(blocks) == 0 {
		return nil, fmt.Errorf("no blocks to decompress")
	}

	var result []float64
	for _, block := range blocks {
		decompressed, err := block.Decompress()
		if err != nil {
			return nil, fmt.Errorf("failed to decompress block: %w", err)
		}
		result = append(result, decompressed...)
	}

	return result, nil
}

// Helper functions

// calculateOptimalRank determines best rank for compression while maintaining accuracy
func (ac *AdaptiveCompressor) calculateOptimalRank(singularValues []float64, rows, cols, targetSize int) int {
	// Get energy-based rank first
	baseRank := ac.DetermineOptimalRank(singularValues)

	// Ensure we don't exceed maxRank
	if baseRank > ac.maxRank {
		baseRank = ac.maxRank
	}

	// If compression isn't required, use base rank
	size := calculateStorageSize(baseRank, rows, cols)
	if size <= targetSize {
		return baseRank
	}

	// Find maximum rank that gives good compression
	for rank := baseRank; rank > 0; rank-- {
		size := calculateStorageSize(rank, rows, cols)
		if size <= targetSize {
			return rank
		}
	}

	// If we can't achieve target size, return minimum rank
	return 1
}

func (cb *CompressedBlock) Decompress() ([]float64, error) {
	if err := validateCompressedBlock(cb); err != nil {
		return nil, err
	}

	result := make([]float64, cb.OriginalRows*cb.OriginalCols)

	// Reconstruct using available components
	for i := 0; i < cb.OriginalRows; i++ {
		for j := 0; j < cb.OriginalCols; j++ {
			var sum float64
			for k := 0; k < len(cb.S); k++ {
				sum += cb.S[k] * cb.U[k][i] * cb.V[k][j]
			}
			result[i*cb.OriginalCols+j] = sum
		}
	}

	// Return exact size
	if len(result) > cb.OriginalSize {
		result = result[:cb.OriginalSize]
	}

	return result, nil
}

func evaluateRank1Quality(singularValues []float64) float64 {
	if len(singularValues) == 0 {
		return 0
	}

	totalEnergy := 0.0
	for _, val := range singularValues {
		totalEnergy += val * val
	}

	if totalEnergy == 0 {
		return 0
	}

	rank1Energy := singularValues[0] * singularValues[0]
	return rank1Energy / totalEnergy
}

func validateCompressedBlock(cb *CompressedBlock) error {
	if cb == nil {
		return fmt.Errorf("compressed block is nil")
	}

	if len(cb.U) == 0 || len(cb.V) == 0 || len(cb.S) == 0 {
		return fmt.Errorf("compressed block has empty components")
	}

	if len(cb.U) != len(cb.V) || len(cb.U) != len(cb.S) {
		return fmt.Errorf("inconsistent dimensions: U:%d, V:%d, S:%d",
			len(cb.U), len(cb.V), len(cb.S))
	}

	return nil
}

// verifyExactReconstruction tests if decompression is lossless
func (cb *CompressedBlock) verifyExactReconstruction(original []float64) bool {
	decompressed, err := cb.Decompress()
	if err != nil {
		return false
	}

	if len(decompressed) != len(original) {
		return false
	}

	const epsilon = 1e-10 // Numerical precision threshold
	for i := range original {
		if math.Abs(original[i]-decompressed[i]) > epsilon {
			return false
		}
	}

	return true
}

// DetermineOptimalRank calculates the optimal rank based on energy retention threshold
func (ac *AdaptiveCompressor) DetermineOptimalRank(singularValues []float64) int {
	if len(singularValues) == 0 {
		return 0
	}

	// Calculate total energy (sum of squared singular values)
	totalEnergy := 0.0
	for _, val := range singularValues {
		totalEnergy += val * val
	}

	if totalEnergy == 0 {
		return 0
	}

	// Accumulate energy until we reach the threshold
	currentEnergy := 0.0
	rank := 0

	for i, val := range singularValues {
		currentEnergy += val * val
		rank = i + 1

		// Check if we've retained enough energy
		if currentEnergy/totalEnergy >= ac.energyThreshold {
			break
		}

		// Don't exceed maxRank
		if rank >= ac.maxRank {
			break
		}
	}

	// Ensure rank is within bounds
	if rank > ac.maxRank {
		rank = ac.maxRank
	}
	if rank > len(singularValues) {
		rank = len(singularValues)
	}

	return rank
}

// Helper function to calculate size of compressed data
func calculateStorageSize(rank, rows, cols int) int {
	// U matrix: rank * rows floats
	// V matrix: rank * cols floats
	// S vector: rank floats
	// Each float64 takes 8 bytes
	return (rank * (rows + cols + 1)) * 8
}

func quantizeVector(vec []float64, tolerance float64) []float64 {
	result := make([]float64, len(vec))
	scale := 1.0 / tolerance

	for i, v := range vec {
		result[i] = math.Round(v*scale) / scale
	}

	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
