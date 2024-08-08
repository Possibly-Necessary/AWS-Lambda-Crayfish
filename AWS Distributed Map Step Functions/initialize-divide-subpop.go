// Lambda function that intialized sub-populations and trigger Step Function state machine

package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sfn"
)

type crayfishParameters struct {
	N int    `json:"n"`
	K int    `json:"k"`
	T int    `json:"t"`
	F string `json:"f"`
}

type SubPopulation struct {
	subPop [][]float64 `json:"subpop"` // Subpopulation i
	T      int         `json:"t"`
	F      string      `json:"f"`
}

type PopulationData struct {
	SubPopulations []SubPopulation `json:"subPopulations"`
}

// Funtion to initialize and divide the population
func initializePopulation(N, k, t int, f string) (*PopulationData, error) { // Instead of returning ([]byte, error)
	// Set the timer
	//startTime := time.Now()

	// Get the benchmark function data
	funcData := benchmark.GetFunction(f) // will hold the string name of the function e.g. "F6"
	lb := funcData.LB
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
			subPop: X[startIndex : startIndex+subPopSize],
			T:      t,
			F:      f,
		}
		startIndex += subPopSize
	}

	return &PopulationData{SubPopulations: Xsub}, nil

}

func HandleDivision(ctx context.Context) (string, error) {

	// These parameters are defined as an input in the State Machine
	// Crayfish parameters: population, sub-populations, COA iteration
	//N, k, t := 500, 20, 500
	// Benchmark function
	//F := "F16"

	var parameters crayfishParameters
	N := parameters.N
	k := parameters.K
	t := parameters.T
	F := parameters.F

	// Initialize and divide population
	subP, err := initializePopulation(N, k, t, F)
	if err != nil {
		log.Fatalf("Error initializing and/or dividing populations: %v", err)
		return "", err
	}

	// Encode subpopulations to JSON
	input, err := json.Marshal(subP)
	if err != nil {
		return "", err
	}

	// Setting up a new AWS Step Function 'sfn' client
	s := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("aws-region-1"),
	}))
	client := sfn.New(s)

	// Start execution of the Step Functions state machine
	result, err := client.StartExecution(&sfn.StartExecutionInput{
		StateMachineArn: aws.String("arn:aws:states:<aws-region>:<account-id>:stateMachine:<state-machine-name>"),
		Input:           aws.String(string(input)),
	})

	if err != nil {
		log.Fatalf("Failed to initiate state machine: %v", err)
		return "", err
	}

	return result.String(), nil
}

func main() {
	lambda.Start(HandleDivision)
}
