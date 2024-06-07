package benchmarks // Change to module name 

import (
	"math"
	"math/rand"
)

// FunctionType for functions F1, F2, ..., F18
type FunctionType func([]float64) float64

// FunctionData stores the function along with its bounds and dimension
type FunctionData struct {
	function FunctionType
	LB       []float64
	UB       []float64
	Dim      int
}

// Function to returnt the selcted benchmark function (made all the dimension 500 to match the original paper)
func GetFunction(F string) FunctionData {
	switch F {
	case "F1":
		return FunctionData{F1, []float64{-100.0}, []float64{100.0}, 500}
	case "F2":
		return FunctionData{F2, []float64{-10.0}, []float64{10.0}, 500}
	case "F3":
		return FunctionData{F3, []float64{-100.0}, []float64{100.0}, 500}
	case "F4":
		return FunctionData{F4, []float64{-100.0}, []float64{100.0}, 500}
	case "F5":
		return FunctionData{F5, []float64{-30.0}, []float64{30.0}, 500}
	case "F6":
		return FunctionData{F6, []float64{-100.0}, []float64{100.0}, 500}
	case "F7":
		return FunctionData{F7, []float64{-1.28}, []float64{1.28}, 500}
	case "F8":
		return FunctionData{F8, []float64{-500.0}, []float64{500.0}, 500}
	case "F9":
		return FunctionData{F9, []float64{-32.0}, []float64{32.0}, 500}
	case "F10":
		return FunctionData{F10, []float64{-32.0}, []float64{32.0}, 500}
	case "F11":
		return FunctionData{F11, []float64{-600.0}, []float64{600.0}, 500}
	case "F16":
		return FunctionData{F16, []float64{-5.0}, []float64{5.0}, 500}
	case "F17":
		return FunctionData{F17, []float64{-5.0}, []float64{5.0}, 500}
	case "F18":
		return FunctionData{F18, []float64{-2.0}, []float64{2.0}, 500}
	}
	panic("Function not defined")
}

// Cigar benchmark
func BentCigarFunction(x []float64) float64 {
	if len(x) == 0 {
		panic("Input slice x must contain at least one element.")
	}
	// Initialize sum by squaring the first element x1^(2).
	sum := x[0] * x[0]

	// Calculate the summation for the squares of the rest of the elements, each multiplied by 10^6
	for i := 1; i < len(x); i++ {
		sum += 1e6 * x[i] * x[i] // This is part of the summation from i=2 to the dimension of x.
	}
	return sum
}

// Benchmark function F1 - Boundary range [-100,100]
func F1(x []float64) float64 {
	sum := 0.0
	for _, value := range x {
		sum += value * value
	}
	return sum
}

// Benchmark function F2 - Boundary range [-10, 10]
func F2(x []float64) float64 {
	sum := 0.0
	product := 1.0

	for _, value := range x {
		absValue := math.Abs(value)
		sum += absValue
		product *= absValue
	}
	return sum + product
}

// Benchmark function F3 - Boundary range [-100,100]
func F3(x []float64) float64 {
	dim := len(x)
	o := 0.0

	for i := 1; i <= dim; i++ {
		sum := 0.0
		for j := 0; j < i; j++ {
			sum += x[j]
		}
		o += sum * sum
	}
	return o
}

// Benchmark function F4 - Boundary range [-100, 100]
func F4(x []float64) float64 {
	maxVal := math.Abs(x[0])
	for _, value := range x {
		absVal := math.Abs(value)
		if absVal > maxVal {
			maxVal = absVal
		}
	}
	return maxVal
}

// Benchmark function F5 - Boundary range [-30,30]
func F5(x []float64) float64 {
	dim := len(x)
	o := 0.0

	for i := 0; i < dim-1; i++ {
		o += 100*math.Pow(x[i+1]-math.Pow(x[i], 2), 2) + math.Pow(x[i]-1, 2)
	}
	return o
}

// Benchmark function F6 - Boundary range [-100,100]
func F6(x []float64) float64 {
	var o float64
	for _, value := range x {
		o += math.Pow(math.Abs(value+0.5), 2)
	}
	return o
}

// Benchmark function F7 - Boundary range [-1.28, 1.28]
func F7(x []float64) float64 {
	//dim := len(x)
	var o float64

	for i, value := range x {
		o += float64(i+1) * math.Pow(value, 4)
	}

	o += rand.Float64() // Adding a random number
	return o
}

// Benchmark function F8 - Boundary range [-500, 500]
func F8(vec []float64) float64 {
	sum := 0.0
	for _, xi := range vec {
		//sum += (-xi * math.Sin(math.Sqrt(math.Abs(xi))))
		sum += (-xi * math.Sin(math.Sqrt(math.Abs(xi))))
	}
	//return 418.9829*float64(len(vec)) - sum
	return sum
}

// Benchmark function F9 - Boundary range [-5.12, 5.12]
func F9(x []float64) float64 {
	dim := len(x)
	o := 0.0

	for _, element := range x {
		o += math.Pow(element, 2) - 10*math.Cos(2*math.Pi*element)
	}
	o += 10 * float64(dim)

	return o
}

// Benchmark function F10 - Boundary range [-32, 32]
func F10(x []float64) float64 {
	dim := len(x)
	sumOfSquares := 0.0
	sumOfCos := 0.0

	for _, value := range x {
		sumOfSquares += value * value
		sumOfCos += math.Cos(2 * math.Pi * value)
	}

	eq1 := -20 * math.Exp(-0.2*math.Sqrt(sumOfSquares/float64(dim)))
	eq2 := -math.Exp(sumOfCos / float64(dim))
	o := eq1 + eq2 + 20 + math.Exp(1)

	return o
}

// Benchmark function F10 - Boundar range [-600,600]
func F11(x []float64) float64 {
	//dim := len(x)
	sumOfSquares := 0.0
	productOfCos := 1.0

	for i, value := range x {
		sumOfSquares += value * value
		productOfCos *= math.Cos(value / math.Sqrt(float64(i+1)))
	}

	o := sumOfSquares/4000 - productOfCos + 1

	return o
}

// Benchmark function F16 - Bound range [-5, 5]
func F16(x []float64) float64 {
	if len(x) < 2 { // Error check
		return 0.0
	}
	x1 := x[0] // x(1) in Matlab
	x2 := x[1] // x(2) in Matlab

	return 4*math.Pow(x1, 2) - 2.1*math.Pow(x1, 4) + math.Pow(x1, 6)/3 + x1*x2 - 4*math.Pow(x2, 2) + 4*math.Pow(x2, 4)
}

// Benchmark function F16 - Boundary range [-5,5]
func F17(x []float64) float64 {
	if len(x) < 2 {
		return 0.0 // Error check for these*
	}
	x1 := x[0] //*
	x2 := x[1]

	pi := math.Pi
	eq1 := x2 - (x1*x1)*5.1/(4*pi*pi) + (5/pi)*x1 - 6
	eq2 := 10*(1-1/(8*pi))*math.Cos(x1) + 10

	return eq1*eq1 + eq2
}

// Benchmark function F18 - Boundary range [-2, 2]
func F18(x []float64) float64 {
	if len(x) < 2 {
		return 0.0
	}

	x1 := x[0]
	x2 := x[1]

	eq1 := 1 + (x1+x2+1)*(x1+x2+1)*(19-14*x1+3*x1*x1-14*x2+6*x1*x2+3*x2*x2)
	eq2 := 30 + (2*x1-3*x2)*(2*x1-3*x2)*(18-32*x1+12*x1*x1+48*x2-36*x1*x2+27*x2*x2)

	return eq1 * eq2
}
