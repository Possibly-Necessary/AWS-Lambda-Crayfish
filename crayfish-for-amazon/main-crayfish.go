package main

import (
	benchmark "amazon-main/benchmark"
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// ___Part that sends the sub-population to the SQS input queue________
// For the input SQS
type Message struct {
	SubPopulation [][]float64
	T             int    // Iteration
	F             string // Function name
	StartTime     time.Time
}

// For the output SQS and aggregation of results
type Result struct {
	BestFit   float64
	BestPos   []float64
	GlobalCov []float64
	StartTime time.Time
}

type Aggregator struct {
	overallBestFit   float64
	overallBestPos   []float64
	overallGlobalCov []float64
	startTime        time.Time
}

// Funtion to initialize and divide the population
func initializePopulation(N, k, t int, fn string, svc *sqs.SQS, sqsUrl string) error { // Instead of returning ([]byte, error)
	// Set the timer
	startTime := time.Now()

	// Get the benchmark function data
	funcData := benchmark.GetFunction(fn) // will hold the string name of the function e.g. "F6"
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

	Xsub := make([][][]float64, k)

	startIndex := 0
	//subPopCount := 0

	for i := 0; i < k; i++ {
		subPopSize := baseSubPopSize
		if remainder > 0 { // In case the division is not even
			subPopSize++ // Add one of the remaining individuals to this sub-population
			remainder--
		}
		Xsub[i] = X[startIndex : startIndex+subPopSize]
		startIndex += subPopSize

		msg := Message{
			SubPopulation: Xsub[i],
			T:             t,
			F:             fn,
			StartTime:     startTime,
		}

		jsonData, err := json.Marshal(msg)
		if err != nil {
			log.Fatalf("Failed to encode message: %v", err)
		}

		payload := &sqs.SendMessageInput{
			MessageBody: aws.String(string(jsonData)),
			QueueUrl:    &sqsUrl,
		}
		// Publish the Amazon's SQS input queue
		_, err = svc.SendMessage(payload)
		if err != nil {
			log.Fatalf("Failed to send populations to SQS: %v", err)
		}

		//subPopCount++
	}

	return nil

}

// Function(s) that handle the aggregatio part
func NewAggregator() *Aggregator {
	return &Aggregator{
		overallBestFit: math.Inf(1),
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

//______ Part that receives the results from the result SQS queue and aggregates them_________

func main() {

	var endTime time.Time
	var workflowExecTime time.Duration

	// Crayfish parameters: population, sub-populations, COA iteration
	N, k, t := 500, 20, 500
	// Benchmark function
	F := "F16"

	// Create a new SQS session/service to send the sub-populations to the SQS queue
	// AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("queue-region"), // Set the AWS region
	})
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}

	// Create an SQS service client
	svc := sqs.New(sess)
	sqsUrl1 := "input-sqs-queue-url"  // Url of the input SQS queue
	sqsUrl2 := "output-sqs-queue-url" // Url od the output SQS queue                                                            // Url of the output SQS queue

	// Call the function to nitialize and divide population, then publishing them to the input SQS queue
	err = initializePopulation(N, k, t, F, svc, sqsUrl1)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Instantiage the Aggregator sturcture with the default settings
	aggregator := NewAggregator()

	// Wait for each sub-populations to compute the fitness value and publishes the results in SQS output queue
	for {
		result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(sqsUrl2),
			MaxNumberOfMessages: aws.Int64(1),
			//WaitTimeSeconds:     aws.Int64(20), // Long polling
		})

		if err != nil {
			log.Fatalf("Failed to receive messages: %v", err)
		}

		if len(result.Messages) == 0 {
			break
		}

		for _, message := range result.Messages {
			var res Result
			if err := json.Unmarshal([]byte(*message.Body), &res); err != nil {
				log.Fatalf("Error parsing message from JSON: %v", err)
			}

			aggregator.updateOverallResults(res)

			// Delete the message from the queue after processing
			_, err = svc.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(sqsUrl2),
				ReceiptHandle: aws.String(*message.ReceiptHandle),
			})
			if err != nil {
				log.Fatalf("Failed to delete message: %v", err)
			}
		}
	}

	// Log results
	endTime = time.Now()
	workflowExecTime = endTime.Sub(aggregator.startTime)
	log.Printf("Overall Best Fitness: %f", aggregator.overallBestFit)
	//if len(aggregator.overallBestPos) > 0 {
	//log.Printf("Best Position: %+v", aggregator.overallBestPos)
	//}
	log.Printf("Executed in: %s", workflowExecTime.String())

}
