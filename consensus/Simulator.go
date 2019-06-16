package consensus

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"time"
	"unsafe"
)

func AttackerSuccessProbability(q float64, z uint64) float64 {
	p := 1.0 - q
	lambda := float64(z) * (q / p)
	sum := 1.0
	var i, k uint64
	for k = 0; k <= z; k++ {
		poisson := math.Exp(-lambda)
		for i = 1; i <= k; i++ {
			poisson *= lambda / float64(i)
		}
		sum -= poisson * (1 - math.Pow(q/p, float64(z-k)))
	}
	return sum
}

func AdversaryCatchup(p float64, q float64, h float64, n int, nonce string) float64 {
	honestPhiFunc := float64(1) - math.Pow(float64(1)-h, p)
	adversaryPhiFunc := float64(1) - math.Pow(float64(1)-h, q)
	multFactor := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(256)), nil)
	honestThreshold := new(big.Int)
	new(big.Float).Mul(big.NewFloat(float64(honestPhiFunc)), new(big.Float).SetInt(multFactor)).Int(honestThreshold)
	adversaryThreshold := new(big.Int)
	new(big.Float).Mul(big.NewFloat(float64(adversaryPhiFunc)), new(big.Float).SetInt(multFactor)).Int(adversaryThreshold)

	honestDraw := "Honest"
	adversaryDraw := "Evil"

	honestWon := 0
	adversaryWon := 0

	for i := 0; i < n; i++ {
		adversaryDrawVal := calculateDrawValue(uint64(i), adversaryDraw, nonce)
		honestDrawVal := calculateDrawValue(uint64(i), honestDraw, nonce)
		if adversaryDrawVal.Cmp(adversaryThreshold) == -1 {
			adversaryWon += 1
		}

		if honestDrawVal.Cmp(honestThreshold) == -1 {
			honestWon += 1
		}
	}
	if honestWon > adversaryWon {
		return 1
	} else if honestWon == adversaryWon {
		return 1.0 / 2.0
	} else {
		return 0
	}
}

func SimulateAdversaryCatchup(p float64, q float64, h float64, n int, simulations int) float64 {
	honestWins := 0.0

	for i := 0; i < simulations; i++ {
		honestWins += AdversaryCatchup(p, q, h, n, RandString(30))
	}
	return float64(honestWins) / float64(simulations)
}

func CalculateGapNeededGivenToleranceAndThreshold(q float64, threshold float64) int {
	n := 0
	currentProb := 1.0
	hardness := 0.1
	for currentProb > threshold {
		n += 10
		currentProb = 1 - SimulateAdversaryCatchup(1-q, q, hardness, n, 10000)
	}

	str := fmt.Sprintf("q = %f    n = %v", q, n)
	fmt.Println(str)
	return n
}

func CalculateGapNeededGivenToleranceAndThresholdAndStartingGap(q float64, threshold float64, n int) int {
	hardness := 0.1
	currentProb := 1 - SimulateAdversaryCatchup(1-q, q, hardness, n, 10000)
	for currentProb > threshold {
		n += 10
		currentProb = 1 - SimulateAdversaryCatchup(1-q, q, hardness, n, 10000)
	}

	str := fmt.Sprintf("q = %f    n = %v", q, n)
	fmt.Println(str)
	return n
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
