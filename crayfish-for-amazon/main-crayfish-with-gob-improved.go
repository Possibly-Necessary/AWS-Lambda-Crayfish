package main

import (
	benchmark "amazon-m/benchmark"
	"bytes"
	"context"
	"encoding/gob"
	"log"
	"math"
	"math/rand"
	"sync" // for goroutines
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-xray-sdk-go/xray"
)

// ___Part that sends the sub-population to the SQS input queue________
// For the input SQS
type Message struct {
	SubPopulation [][]float64
	SubPopNum     int
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
	SubPopN   []int // To store the tracking number of the sub-population
}

type Aggregator struct {
	overallBestFit   float64
	overallBestPos   []float64
	overallGlobalCov []float64
	startTime        time.Time
}

// Funtion to initialize and divide the population
func initializePopulation(ctx context.Context, N, k, t int, fn string, svc *sqs.SQS, sqsUrl string) error { // Instead of returning ([]byte, error)
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

	var wg sync.WaitGroup // For goroutines

	for i := 0; i < k; i++ {
		subPopSize := baseSubPopSize
		if remainder > 0 { // In case the division is not even
			subPopSize++ // Add one of the remaining individuals to this sub-population
			remainder--
		}
		Xsub[i] = X[startIndex : startIndex+subPopSize]
		startIndex += subPopSize

		// Using Goroutines for I/O tasks
		wg.Add(1)
		go func(ctx context.Context, subPop [][]float64, subPopNum int) {
			defer wg.Done()
			var localBuffer bytes.Buffer
			encoder := gob.NewEncoder(&localBuffer)
			msg := Message{
				SubPopulation: subPop,
				SubPopNum:     subPopNum,
				T:             t,
				F:             fn,
				StartTime:     startTime,
			}

			if err := encoder.Encode(msg); err != nil {
				log.Fatalf("Failed to encode message: %v", err)
			}

			// Wrap the SQS SendMessageInput with X-Ray
			xray.Capture(ctx, "SendSQSMessage", func(ctx1 context.Context) error {
				payload := &sqs.SendMessageInput{
					MessageBody: aws.String(localBuffer.String()),
					QueueUrl:    &sqsUrl,
				}
				_, err := svc.SendMessageWithContext(ctx1, payload)
				if err != nil {
					log.Fatalf("Failed to send populations to SQS: %v", err)
				}
				return err
			})

		}(ctx, Xsub[i], i+1) // To number the subpopulations starting from 1
	}

	wg.Wait()
	return nil

}

// Function(s) that handle the aggregation part
func NewAggregator() *Aggregator {
	return &Aggregator{
		overallBestFit: math.Inf(1),
		startTime:      time.Now(),
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

	var waitStart, waitEnd time.Time
	var totalWaitTime time.Duration

	var subPopTrack [][]int // To track the numbers

	// Crayfish parameters: population, sub-populations, COA iteration
	N, k, t := 500, 20, 500
	// Benchmark function
	F := "F11"
	receivedSubPop := 0 // intially

	// Configure X-Ray Daemon address for context tracing
	xray.Configure(xray.Config{
		DaemonAddr: "127.0.0.1:2000", // Default X-Ray Daemon IP address and port
	}) // Must install (or pull it from Docker) and run the X-Ray Daemon locally first

	// Create a new SQS session/service to send the sub-populations to the SQS queue
	// AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("aws-region"), // Set the AWS region
	})
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}

	// Create an SQS service client
	svc := sqs.New(sess)
	// Wrap the SQS service client object with X-Ray for event-driven tracing
	xray.AWS(svc.Client)

	sqsUrl1 := "url-of-input-queue"  // Url of the input SQS queue
	sqsUrl2 := "url-of-output-queue" // Url of the output SQS queue

	aggregator := NewAggregator()
	// Call the function to nitialize and divide population, then publishing them to the input SQS queue

	ctx := context.Background()
	err = initializePopulation(ctx, N, k, t, F, svc, sqsUrl1)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Wait for each sub-populations to compute the fitness value and publishes the results in SQS output queue
	for {

		if receivedSubPop >= k {
			break
		}
		waitStart = time.Now() // Timing the waiting operation

		result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(sqsUrl2),
			MaxNumberOfMessages: aws.Int64(1),
			//WaitTimeSeconds:     aws.Int64(20), // Long polling
		})

		waitEnd = time.Now()
		totalWaitTime += waitEnd.Sub(waitStart)

		if err != nil {
			log.Fatalf("Failed to receive messages: %v", err)
		}

		if len(result.Messages) == 0 {
			continue
		}

		// Using goroutines for I/O tasks
		var wg sync.WaitGroup
		for _, message := range result.Messages {
			wg.Add(1)
			go func(msg *sqs.Message) {
				defer wg.Done()
				var res Result

				msgByte := bytes.NewBufferString(*msg.Body)
				decoder := gob.NewDecoder(msgByte)

				if err := decoder.Decode(&res); err != nil {
					log.Fatalf("Failed to decode message: %v", err)
				}

				aggregator.updateOverallResults(res)
				receivedSubPop++
				subPopTrack = append(subPopTrack, res.SubPopN)

				// Delete the message from the queue after processing
				_, err = svc.DeleteMessage(&sqs.DeleteMessageInput{
					QueueUrl:      aws.String(sqsUrl2),
					ReceiptHandle: msg.ReceiptHandle,
				})
				if err != nil {
					log.Fatalf("Failed to delete message: %v", err)
				}
			}(message)
		}
		wg.Wait()
	}

	// Log results
	endTime := time.Now()
	workflowExecTime := endTime.Sub(aggregator.startTime)

	log.Printf("Overall Best Fitness: %f", aggregator.overallBestFit)

	log.Printf("Executed in: %s", workflowExecTime.String())
	log.Printf("Sub-population numbers: %v", subPopTrack)
	log.Printf("Waiting For messages from SQS: %s", totalWaitTime.String())

}
