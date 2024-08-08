package main

import (
	"context"
	"math"

	"github.com/aws/aws-lambda-go/lambda"
)

type Result struct {
	BestFit   float64   `json:"bestFit"`
	BestPos   []float64 `json:"bestPos"`
	GlobalCov []float64 `json:"globalCov"`
}

type AggregationResult struct {
	OverallBestFit   float64   `json:"overallBestFit"`
	OverallBestPos   []float64 `json:"overallBestPos"`
	OverallGlobalCov []float64 `json:"overallGlobalCov"`
}

type Aggregator struct {
	overallBestFit   float64
	overallBestPos   []float64
	overallGlobalCov []float64
	//startTime        time.Time
}

func NewAggregator() *Aggregator {
	return &Aggregator{
		overallBestFit: math.Inf(1),
		//startTime:      time.Now(),
	}
}

func (a *Aggregator) updateOverallResults(result Result) {
	if result.BestFit < a.overallBestFit {
		a.overallBestFit = result.BestFit
		a.overallBestPos = make([]float64, len(result.BestPos))
		copy(a.overallBestPos, result.BestPos)
	}
	if a.overallGlobalCov == nil {
		a.overallGlobalCov = make([]float64, len(result.GlobalCov))
	} else {
		for i, cov := range result.GlobalCov {
			a.overallGlobalCov[i] += cov
		}
	}
}

// Lambda function that aggregates results from the aggregate result state in the state machine
func HandleAggregation(ctx context.Context, results []Result) (AggregationResult, error) {
	aggregator := NewAggregator()
	for _, result := range results {
		aggregator.updateOverallResults(result)
	}
	return AggregationResult{
		OverallBestFit:   aggregator.overallBestFit,
		OverallBestPos:   aggregator.overallBestPos,
		OverallGlobalCov: aggregator.overallGlobalCov,
	}, nil
}

func main() {
	lambda.Start(HandleAggregation)
}
