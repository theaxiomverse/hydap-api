package agglomerator

import (
	"github.com/theaxiomverse/hydap-api/pkg/vectors"
	"math"
)

// RouteMetrics holds metrics for route evaluation
type RouteMetrics struct {
	Speed      float64 // Based on TPS and block time
	Finality   float64 // Time to finality
	Cost       float64 // Transaction cost
	Similarity float64 // Vector similarity score
}

// calculateRouteMetrics computes metrics for a potential route
func calculateRouteMetrics(chain *Chain, tx *Transaction) RouteMetrics {
	protocol := determineProtocol(chain.ID)
	config, exists := getProtocolConfig(protocol)
	if !exists {
		return RouteMetrics{}
	}

	// Calculate base metrics
	speed := math.Log(1+config.TPS) / config.BlockTime
	finality := 1 / config.Finality // Inverse so higher is better
	cost := 1 - config.CostWeight   // Inverse so higher is better

	// Calculate vector similarity
	similarity := vectors.ComputeVectorSimilarity(
		chain.StateVector,
		tx.StateVector,
		50, // Consider parameterizing this
	)

	return RouteMetrics{
		Speed:      speed,
		Finality:   finality,
		Cost:       cost,
		Similarity: similarity,
	}
}

// evaluateRoute scores a potential route based on metrics
func evaluateRoute(metrics RouteMetrics) float64 {
	// Weights for different factors
	const (
		speedWeight      = 0.3
		finalityWeight   = 0.25
		costWeight       = 0.2
		similarityWeight = 0.25
	)

	// Combine weighted factors
	score := (metrics.Speed * speedWeight) +
		(metrics.Finality * finalityWeight) +
		(metrics.Cost * costWeight) +
		(metrics.Similarity * similarityWeight)

	return score
}

// findOptimalRoute determines the best route for a transaction
func findOptimalRoute(chains []*Chain, tx *Transaction) []*Chain {
	var bestRoute []*Chain
	var bestScore float64

	for _, chain := range chains {
		metrics := calculateRouteMetrics(chain, tx)
		score := evaluateRoute(metrics)

		if score > bestScore {
			bestScore = score
			bestRoute = []*Chain{chain}
		}
	}

	return bestRoute
}
