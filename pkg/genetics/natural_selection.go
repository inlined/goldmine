package genetics

import (
	"math/rand"
	"time"
)

// NaturalSelection is an interface to pick the selection method.
// A NaturalSelection MAY NOT BE GOROUTINE SAFE. It may only be used in one Evolve function at a time.
// This helps avoid the lock incurred by the top-level rand functions.
// TODO: consider nested interfaces (NaturalSelection has a Seed() function to return a Selector
// that implements SelectParents). This would avoid re-generating the roulette wheel in
// StochasticUniversalSampling
type NaturalSelection interface {
	SelectParents(numParents int, fitness []Fitness) (indexes []int)
}

type sus struct {
	rng *rand.Rand
}

// StochasticUniversalSampling creates a "roulette" wheel where each parent
// gets a slice in proportion to their fitness. We then spin the wheel with
// two fixed points to select which parents win.
// If src is nill, a new source is created with the current time.
func StochasticUniversalSampling(src rand.Source) NaturalSelection {
	if src == nil {
		src = rand.NewSource(time.Now().UTC().UnixNano())
	}
	return &sus{rng: rand.New(src)}
}

func (s sus) SelectParents(numParents int, fitness []Fitness) (indexes []int) {
	totalFitness := Fitness(0)
	for _, f := range fitness {
		totalFitness += f
	}

	// Use a fixed distance (uniform distribution) across the wheel.
	// Note: we choose here to use integer arithmetic instead of a float distribution.
	// This uses faster ALUs but introduces the possibility of error when totalFitness !>> numParents
	distance := totalFitness / Fitness(numParents)
	// Spin the wheel up to distance (equivalent to spinning the wheel randomly and then taking the modulo
	// of the size)
	pos := Fitness(s.rng.Int63n(int64(distance)))

	// Iterate through the fitness scores as if it were a roulete wheel (e.g. incrementing f by
	// fitness[n] rather than one) and remember the indexes which contain any pointers P.
	// In edge cases, a position may hit the same parent multiple times; in this case, the parent
	// is selected repeatedly.
	// TODO: Should this be instead selected with a weight to avoid a parent mating with itself?
	indexes = make([]int, 0, numParents)
	accumFitness := Fitness(0)
	for n := 0; len(indexes) < numParents; n++ {
		accumFitness += fitness[n]
		for ; pos < accumFitness; pos += distance {
			indexes = append(indexes, n)
		}
	}

	return indexes
}
