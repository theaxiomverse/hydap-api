package vectors

import (
	"fmt"
	"math"
	"sync"
)

type InfiniteVectorIndex struct {
	mu                  sync.RWMutex
	vectorSpace         map[string]InfiniteVector
	dimensionGenerators map[string]func(int) float64
	metadataStore       map[string]map[string]interface{}
}

type InfiniteVector struct {
	mu        sync.RWMutex
	elements  []float64
	Generator func(int) float64
}

type DatabaseRecord struct {
	ID       string
	Metadata map[string]interface{}
	Vector   InfiniteVector
}

func NewInfiniteVectorIndex() *InfiniteVectorIndex {
	return &InfiniteVectorIndex{
		vectorSpace:         make(map[string]InfiniteVector),
		dimensionGenerators: make(map[string]func(int) float64),
		metadataStore:       make(map[string]map[string]interface{}),
	}
}

func (db *InfiniteVectorIndex) DefineVectorSpace(spaceName string, generator func(int) float64) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.dimensionGenerators[spaceName] = generator
}

func (db *InfiniteVectorIndex) Insert(record DatabaseRecord) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if record.Vector.Generator == nil {
		record.Vector.Generator = func(dim int) float64 {
			return math.Sin(float64(dim)) * math.Exp(-float64(dim)/10.0)
		}
	}

	db.vectorSpace[record.ID] = record.Vector
	db.metadataStore[record.ID] = record.Metadata

	return nil
}

func (db *InfiniteVectorIndex) QueryByDimension(
	dimensionSelector func(vector InfiniteVector) bool,
	maxResults int,
) []DatabaseRecord {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []DatabaseRecord

	for id, vector := range db.vectorSpace {
		if dimensionSelector(vector) {
			results = append(results, DatabaseRecord{
				ID:       id,
				Metadata: db.metadataStore[id],
				Vector:   vector,
			})

			if len(results) >= maxResults {
				break
			}
		}
	}

	return results
}

func (db *InfiniteVectorIndex) AdvancedQuery(
	similarityThreshold float64,
	queryVector InfiniteVector,
	maxDimensions int,
) []DatabaseRecord {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []DatabaseRecord

	for id, vector := range db.vectorSpace {
		similarity := ComputeVectorSimilarity(queryVector, vector, maxDimensions)
		if similarity >= similarityThreshold {
			results = append(results, DatabaseRecord{
				ID:       id,
				Metadata: db.metadataStore[id],
				Vector:   vector,
			})
		}
	}

	return results
}

func ComputeVectorSimilarity(v1, v2 InfiniteVector, dimensions int) float64 {
	var sumXY, sumX, sumY, sumX2, sumY2 float64

	for i := 0; i < dimensions; i++ {
		x := v1.GetElement(i)
		y := v2.GetElement(i)

		sumXY += x * y
		sumX += x
		sumY += y
		sumX2 += x * x
		sumY2 += y * y
	}

	n := float64(dimensions)
	numerator := n*sumXY - sumX*sumY
	denominator := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

func (v *InfiniteVector) GetElement(dimension int) float64 {
	v.mu.Lock()
	defer v.mu.Unlock()

	if dimension >= len(v.elements) {
		for len(v.elements) <= dimension {
			v.elements = append(v.elements, v.Generator(len(v.elements)))
		}
	}

	return v.elements[dimension]
}

func ExampleUsage() {
	db := NewInfiniteVectorIndex()

	db.DefineVectorSpace("exponential", func(dim int) float64 {
		return math.Pow(0.5, float64(dim))
	})

	db.DefineVectorSpace("sinusoidal", func(dim int) float64 {
		return math.Sin(float64(dim)) * math.Pow(-1, float64(dim))
	})

	db.Insert(DatabaseRecord{
		ID: "record1",
		Metadata: map[string]interface{}{
			"category": "science",
			"tags":     []string{"physics", "quantum"},
		},
		Vector: InfiniteVector{
			Generator: func(dim int) float64 {
				return math.Exp(-float64(dim) / 5.0)
			},
		},
	})

	db.Insert(DatabaseRecord{
		ID: "record2",
		Metadata: map[string]interface{}{
			"category": "technology",
			"tags":     []string{"ai", "machine learning"},
		},
		Vector: InfiniteVector{
			Generator: func(dim int) float64 {
				return math.Sin(float64(dim)) * 0.5
			},
		},
	})

	queryVector := InfiniteVector{
		Generator: func(dim int) float64 {
			return math.Exp(-float64(dim) / 10.0)
		},
	}

	results := db.AdvancedQuery(0.7, queryVector, 50)
	fmt.Println("Query Results:", results)

	specificResults := db.QueryByDimension(
		func(vector InfiniteVector) bool {
			return vector.GetElement(5) > 0
		},
		5,
	)
	fmt.Println("\nSpecific Dimension Results:", specificResults)
}
