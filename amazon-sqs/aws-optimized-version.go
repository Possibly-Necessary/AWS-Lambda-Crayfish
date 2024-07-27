/* 
	Changes that might help with cold start and Lambda's performance:

		- Changed encoding/decoding message from JSON to 'gob' binary encoding/decoding (gob outperforms JSON -- it's fast and generates small data)
		- SQS AWS SDK client initialization moved outside the Lambda Handler to be reused across multipl requests.
		- Moved Benchmark mapping outside (global scope of) the Handler function to minimize initialization time.
  	Other changes:
   
		- Added context logging extracted from Lambda's context object. 
  		- Added structured logging for AWS CloudWatch.
    		- Added sub-population tracking.
*/

package main

import (
	benchmarks "amazon-c/benchmark"
	"context"
	"encoding/json"
	"bytes"
	"encoding/gob"
	"log"
	"math"
	"math/rand"
	"os/exec"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda" // For extracting logging info. from the context object
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
)

// For AWS CloudWatch/Lambda insight structured logging
type LogEntry struct {
	FunctionName          string `json:"functionName"`
	FunctionVersion       string `json:"functionVersion"`
	AWSRequestID          string `json:"awsRequestId"`
	MemoryLimitInMB       int    `json:"memoryLimitInMB"`
	LogGroupName          string `json:"logGroupName"`
	LogStreamName         string `json:"logStreamName"`
	CognitoIdentityID     string `json:"cognitoIdentityId,omitempty"`
	CognitoIdentityPoolID string `json:"congnitoIdentityPoolId,omitempty"`
	ExecutionTime         string `json:"executionTime"`
	BenchmarkFunction     string `json:"benchmarkFunction"`
}

// For the input SQS
type Message struct {
	SubPopulation [][]float64
	T             int    // Iteration
	F             string // Function name
	StartTime     time.Time
}

// For the output SQS
type Result struct {
	StartTime time.Time
	BestFit   float64
	BestPos   []float64
	GlobalCov []float64
}

// Initializing SQS queue globally so that we do not create a new client for every invocation (helps with cold starts)
var svc *sqs.SQS
var sqsOutUrl = "output-sqs-queue-url"


// For gob encoding
var buffer bytes.Buffer

// Function mappings moved to the global scope
var functionMap = map[string]benchmarks.FunctionType{
	"F1":  benchmarks.F1,
	"F2":  benchmarks.F2,
	"F3":  benchmarks.F3,
	"F4":  benchmarks.F4,
	"F5":  benchmarks.F5,
	"F6":  benchmarks.F6,
	"F7":  benchmarks.F7,
	"F8":  benchmarks.F8,
	"F9":  benchmarks.F9,
	"F10": benchmarks.F10,
	"F11": benchmarks.F11,
	"F16": benchmarks.F16,
	"F17": benchmarks.F17,
	"F18": benchmarks.F18,
}

func init() {
	s := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("aws-region"),
	}))
	svc = sqs.New(s)
}

// Function for dynamic benchmark selection
func selectedBenchmark(F string) benchmarks.FunctionType {
	// Dynamically select a benchmark funciton
	candidateFunc, ok := functionMap[F]
	if !ok {
		//context.Logger.Error("Function does not exist..\n")
		errors.New("Function does not exist..\n")
	}
	return candidateFunc
}

// _____________ Main Crayfish Algorithm________________
// Equation 4: Mathematical model of crayfish intake
func p_obj(x float64) float64 {
	return 0.2 * (1 / (math.Sqrt(2*math.Pi) * 3)) * math.Exp(-math.Pow(x-25, 2)/(2*math.Pow(3, 2)))
}

func crayfish(T int, lb, ub []float64, f string, X [][]float64, F benchmarks.FunctionType) (x float64, y, z []float64) { // return bestFit, bestPos

	N := len(X)      // size of the sub-population
	dim := len(X[0]) // dimension of the sub-populationl

	var (
		globalCov   []float64 = make([]float64, T) // zero row vector of size T
		BestFitness           = math.Inf(1)
		BestPos     []float64 = make([]float64, dim)
		fitnessF    []float64 = make([]float64, N)
		GlobalPos   []float64 = make([]float64, dim)
	)

	for i := 0; i < N; i++ {
		fitnessF[i] = F(X[i]) // Get the fitness value from the benchmark function
		if fitnessF[i] < BestFitness {
			BestFitness = fitnessF[i]
			copy(BestPos, X[i])
		}
	}

	// Update best position to Global position
	copy(GlobalPos, BestPos)
	GlobalFitness := BestFitness

	Xf := make([]float64, dim) // For Xshade -- array for the cave
	Xfood := make([]float64, dim)

	Xnew := make([][]float64, N) // Initializing a 2d array
	for i := 0; i < N; i++ {
		Xnew[i] = make([]float64, dim)
	}

	t := 0
	for t < T {
		//Decreasing curve --> Equation 7
		C := 2 - (float64(t) / float64(T))
		//Define the temprature from Equation 3
		tmp := rand.Float64()*15 + 20

		for i := 0; i < dim; i++ { // Calculating the Cave -> Xshade = XL + XG/2
			Xf[i] = (BestPos[i] + GlobalPos[i]) / 2
		}
		copy(Xfood, BestPos) // copy the best position to the Xfood vector

		for i := 0; i < N; i++ {
			//Xnew[i] = make([]float64, dim) //--> took this part out
			if tmp > 30 { // Summer resort stage
				if rand.Float64() < 0.5 {
					for j := 0; j < dim; j++ { // Equation 6
						Xnew[i][j] = X[i][j] + C*rand.Float64()*(Xf[j]-X[i][j])
					}
				} else { // Competition Stage
					for j := 0; j < dim; j++ {
						z := rand.Intn(N) // Random crayfish
						//z := math.Round(rand.Float64()*(N-1)) + 1 //--> or try this
						Xnew[i][j] = X[i][j] - X[z][j] + Xf[j] // Equation 8
					}
				}
			} else { // Foraging stage
				P := 3 * rand.Float64() * fitnessF[i] / F(Xfood)
				if P > 2 {
					//Food is broken down becuase it's too big
					for j := 0; j < dim; j++ {
						Xfood[j] *= math.Exp(-1 / P)
						Xnew[i][j] = X[i][j] + math.Cos(2*math.Pi*rand.Float64())*Xfood[j]*p_obj(tmp) - math.Sin(2*math.Pi*rand.Float64())*Xfood[j]*p_obj(tmp)
					} // ^^ Equation 13: crayfish foraging
				} else {
					for j := 0; j < dim; j++ { // The case where the food is a moderate size
						Xnew[i][j] = (X[i][j]-Xfood[j])*p_obj(tmp) + p_obj(tmp)*rand.Float64()*X[i][j]
					}
				}
			}
		}

		// Boundary conditions checks
		for i := 0; i < N; i++ {
			for j := 0; j < dim; j++ {
				if len(ub) == 1 {
					Xnew[i][j] = math.Min(ub[0], Xnew[i][j])
					Xnew[i][j] = math.Max(lb[0], Xnew[i][j])
				} else {
					Xnew[i][j] = math.Min(ub[j], Xnew[i][j])
					Xnew[i][j] = math.Max(lb[j], Xnew[i][j])
				}
			}
		}

		//Global update stuff
		copy(GlobalPos, Xnew[0])
		GlobalFitness = F(GlobalPos)

		for i := 0; i < N; i++ {
			NewFitness := F(Xnew[i])
			if NewFitness < GlobalFitness {
				GlobalFitness = NewFitness
				copy(GlobalPos, Xnew[i])
			}

			// Update population to a new location
			if NewFitness < fitnessF[i] {
				fitnessF[i] = NewFitness
				copy(X[i], Xnew[i])
				if fitnessF[i] < BestFitness {
					BestFitness = fitnessF[i]
					copy(BestPos, X[i])
				}
			}
		}

		globalCov[t] = GlobalFitness

		t++
	}

	return BestFitness, BestPos, globalCov
}

// ________________________________________________________________
// Lambda function that receives an SQS event and gets triggered by it
func Handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	startTime := time.Now() // Record starting time of the Lambda function

	// Extract information from the Lambda context object 'ctx'
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		log.Println("Failed to extract Lambda context.")
		return errors.New("failed to extract Lambda conetxt.")
     	}

	// Reading Events (sub-populations)
	for _, message := range sqsEvent.Records {
		var sqsData Message
		dataBytes:= []byte(message.Body)
		dataBuffer := bytes.NewBuffer(dataBytes)
		decoder := gob.NewDecoder(dataBuffer)

		// Decode message
		err := decoder.Decode(&sqsData)
		//err := json.Unmarshal([]byte(message.Body), &sqsData)
		if err != nil {
			log.Printf("Failed to decode message: %v", err)
			return err
		}

		// Get the benchmark function specifications from rabbitData.F
		specs := benchmarks.GetFunction(sqsData.F)
		lb := specs.LB
		ub := specs.UB
		//dim := specs.Dim
		F := selectedBenchmark(sqsData.F) // Get the actual function from the string

		function := sqsData.F // Store the name of the benchmark function in the global variable for logging

		// Start crayfish algorithm and return results
		bestFit, bestPos, globalCov := crayfish(sqsData.T, lb, ub, sqsData.F, sqsData.SubPopulation, F)

		res := Result{
			StartTime: sqsData.StartTime,
			BestFit:   bestFit,
			BestPos:   bestPos,
			GlobalCov: globalCov,
		}

		// Encode results using gob
		encoder := gob.NewEncoder(&buffer)
		err = encoder.Encode(res)

		//jsonResult, err := json.Marshal(res)
		if err != nil {
			log.Fatalf("Failed to encode results: %v", err)
		}

		_, err = svc.SendMessage(&sqs.SendMessageInput{
			QueueUrl:    &sqsOutUrl,
			MessageBody: aws.String(buffer.String()), // Changed 'aws.String(string(jsonResult)) 
		})

		if err != nil {
			log.Printf("Failed to send message: %v", err)
		}
	}

	endTime := time.Since(startTime) // endTime - startTime
	
	// Log/print the request ID, which is unique across the concurrent functions - this is to identify the unique function instances spawned from a particular invocation
	// and calculate their execution time
	log.Printf("Execution time: %v, AWS Request ID: %s", endTime, lc.AwsRequestID)
	logEntry := LogEntry{
		FunctionName:      lc.FunctionName,
		FunctionVersion:   lc.FunctionVersion,
		AWSRequestID:      lc.AwsRequestID,
		MemoryLimitInMB:   lc.MemoryLimitInMB,
		LogGroupName:      lc.LogGroupName,
		LogStreamName:     lc.LogStreamName,
		ExecutionTime:     endTime.String(),
		BenchmarkFunction: sqsData.F,
	}

	if lc.Identity != nil {
		logEntry.CognitoIdentityID = lc.Identity.CognitoIdentityID
		logEntry.CognitoIdentityPoolID = lc.Identity.CognitoIdentityPoolID
	}

	logData, _ := json.Marshal(logEntry)
	log.Println(string(logData))

	return nil
}

func main() {
	lambda.Start(Handler) // Start the Lambda function execution
}
