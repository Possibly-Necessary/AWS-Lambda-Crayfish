// Lambda function that intialized sub-populations and trigger Step Function state machine
// Suppopulation are added into an s3 bucket due to its large payload, and only the bucket's name and key are passed 

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"math/rand"
	benchmark "step-one/benchmark"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/service/s3"
)

// Structure for the input event of the state machine
type crayfishParameters struct {
	N int    `json:"n"`
	K int    `json:"k"`
	T int    `json:"t"`
	F string `json:"f"`
}

// Structure for the output event for the processing lambda function (within the distributed map)
type SubPopulation struct {
	SubPop [][]float64 `json:"subpop"` // Subpopulation i
	T      int         `json:"t"`
	F      string      `json:"f"`
}

type PopulationData struct {
	SubPopulations []SubPopulation `json:"subPopulations"`
}

// Funtion to initialize and divide the population
func initializePopulation(N, k, t int, f string) PopulationData { // Instead of returning ([]byte, error)
	// Set the timer
	//startTime := time.Now()

	// Get the benchmark function data
	funcData := benchmark.GetFunction(f) // will hold the string name of the function e.g. "F6"
	lb := funcData.LB                    // Get the upper/lower bound and dimension assocciated with the selected function
	ub := funcData.UB
	dim := funcData.Dim

	// Initialize the population N x Dim matrix, X
	X := make([][]float64, N)
	for i := 0; i < N; i++ {
		X[i] = make([]float64, dim)
	}

	for i := range X {
		for j := range X[i] {
			X[i][j] = rand.Float64()*(ub[0]-lb[0]) + lb[0]
		}
	}

	// Split the population based on k
	totalSize := len(X)
	baseSubPopSize := totalSize / k // N/k
	remainder := totalSize % k

	//Xsub := make([][][]float64, k)
	Xsub := make([]SubPopulation, k)
	startIndex := 0
	//subPopCount := 0

	for i := 0; i < k; i++ {
		subPopSize := baseSubPopSize
		if remainder > 0 { // In case the division is not even
			subPopSize++ // Add one of the remaining individuals to this sub-population
			remainder--
		}
		//Xsub[i] = X[startIndex : startIndex+subPopSize]
		//startIndex += subPopSize

		Xsub[i] = SubPopulation{
			SubPop: X[startIndex : startIndex+subPopSize],
			T:      t,
			F:      f,
		}
		startIndex += subPopSize
	}

	return PopulationData{SubPopulations: Xsub}

}

func HandleDivision(ctx context.Context, parameters crayfishParameters) (PopulationData, error) {

	// These parameters are defined as an input in the State Machine
	// Crayfish parameters: population, sub-populations, COA iteration
	//N, k, t := 500, 20, 500
	//F := Benchmark function

	result := initializePopulation(parameters.N, parameters.K, parameters.T, parameters.F)

	return result, nil

}

func main() {
	lambda.Start(HandleDivision)
}
