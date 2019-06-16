package consensus

import (
	"fmt"
	"testing"
)

func TestAttackerSuccessProbability1(t *testing.T) {
	PrintSolveForThreshold(0.001)
	PrintSolveForThreshold(0.0001)
	PrintSolveForThreshold(0.00001)
}

func TestAttackerSuccessProbability01(t *testing.T) {
	PrintSolveForThreshold(0.001)
}

func TestCrap(t *testing.T) {
	x := uint(50)
	y := uint(75)
	fmt.Println(y - x)
}

func PrintSolveForThreshold(P float64) {
	str := fmt.Sprintf("P < %f", P)
	fmt.Println(str)
	n1 := CalculateGapNeededGivenToleranceAndThreshold(0.1, P)
	n2 := CalculateGapNeededGivenToleranceAndThresholdAndStartingGap(0.15, P, n1)
	n3 := CalculateGapNeededGivenToleranceAndThresholdAndStartingGap(0.2, P, n2)
	n4 := CalculateGapNeededGivenToleranceAndThresholdAndStartingGap(0.25, P, n3)
	n5 := CalculateGapNeededGivenToleranceAndThresholdAndStartingGap(0.3, P, n4)
	n6 := CalculateGapNeededGivenToleranceAndThresholdAndStartingGap(0.35, P, n5)
	n7 := CalculateGapNeededGivenToleranceAndThresholdAndStartingGap(0.4, P, n6)
	CalculateGapNeededGivenToleranceAndThresholdAndStartingGap(0.45, P, n7)
}

func TestSimulateAdversaryCatchup(t *testing.T) {
	fmt.Println(SimulateAdversaryCatchup(0.9, 0.1, 0.1, 50, 10000))
	fmt.Println(SimulateAdversaryCatchup(0.85, 0.15, 0.1, 50, 10000))
	fmt.Println(SimulateAdversaryCatchup(0.8, 0.2, 0.1, 50, 10000))
	fmt.Println(SimulateAdversaryCatchup(0.75, 0.25, 0.1, 50, 10000))
	fmt.Println(SimulateAdversaryCatchup(0.7, 0.3, 0.1, 50, 10000))
	fmt.Println(SimulateAdversaryCatchup(0.65, 0.35, 0.1, 50, 10000))
	fmt.Println(SimulateAdversaryCatchup(0.6, 0.4, 0.1, 50, 10000))
}

func TestSimulateAdversaryCatchup2(t *testing.T) {
	fmt.Println(SimulateAdversaryCatchup(0.9, 0.1, 0.1, 50, 10000))
}

func BitCoinCalculateGapNeededGivenToleranceAndThreshold(q float64, threshold float64) int {
	z := 0
	p := 1.0
	for p > threshold {
		z++
		p = AttackerSuccessProbability(q, uint64(z))
	}
	str := fmt.Sprintf("q = %f    z = %v", q, z)
	fmt.Println(str)
	return z
}
