// Lambda function of the second state in the COA state machine; its logic is developed for processing a single sub-population
// This lambda function represents a single child workflow execution of applying COA algorithm on sub-population i and returning the corresponding results to the next state

package main

import (
	"context"
	"errors"
	"math"
	"math/rand"
	benchmarks "step-two/benchmark"

	"github.com/aws/aws-lambda-go/lambda"
)

type SubPopulation struct {
	SubPop [][]float64 `json:"subpop"` // Subpopulation i
	T      int         `json:"t"`
	F      string      `json:"f"`
}

type SubResult struct {
	BestFit   float64   `json:"bestFit"`
	BestPos   []float64 `json:"bestPos"`
	GlobalCov []float64 `json:"globalCov"`
}

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

func HandleSubPop(ctx context.Context, subP SubPopulation) (SubResult, error) {
	// The the benchmark specifications
	specs := benchmarks.GetFunction(subP.F)
	lb := specs.LB
	ub := specs.UB
	F := selectedBenchmark(subP.F) // Get the actual function from the string

	// Call the Crayfish (main algorithm) function
	bestFit, bestPos, globalCov := crayfish(subP.T, lb, ub, subP.F, subP.SubPop, F)

	res := SubResult{
		BestFit:   bestFit,
		BestPos:   bestPos,
		GlobalCov: globalCov,
	}

	return res, nil
}

func main() {
	lambda.Start(HandleSubPop)
}
