package bruteforce

import (
	"fmt"

	"github.com/inlined/genetics"
	"github.com/inlined/goldmine/pkg/maps"
	"github.com/inlined/goldmine/pkg/solver"
)

const (
	// How many extra genes to allow for walking into disallowed spaces
	genomePaddingRatio = 1.2
)

func init() {
	solver.RegisterSolverFlag("bruteforce", func(i solver.Input) solver.Solver {
		return &Solver{Input: i}
	})
}

// Solver solves a Goldmine map with brute force
type Solver struct {
	solver.Input
	species    *genetics.Species
	population []genetics.Chromosome
	best       genetics.Chromosome
	score      int
}

func toDir(g genetics.Gene) maps.Direction {
	switch g {
	case 0:
		return maps.Up
	case 1:
		return maps.Down
	case 2:
		return maps.Left
	case 3:
		return maps.Right
	default:
		panic(fmt.Sprintf("Unexpected gene %d", g))
	}
}

// Path transaltes a Chromosome into a valid Path
func (s Solver) Path(c genetics.Chromosome) maps.Path {
	p := maps.Path(make([]maps.Direction, 0, s.Map.StepsAllowed))
	v := s.Map.PointsOfInterest[0]

	for i := 0; i < c.Species.NumGenes && p.Len() <= s.Map.StepsAllowed; i++ {
		d := toDir(c.Genes[i])
		v2 := v.Move(d)
		if !s.Map.CanBeAt(v2) {
			continue
		}
		p.Append(d)
		v = v2
	}

	// In case we run out of valid genes before StepsAllowed
	p.Pad(s.Map)
	return p
}

// Init creates all necessary private variables
func (s *Solver) Init(popSize int) error {
	numGenes := float32(s.Map.StepsAllowed) * genomePaddingRatio
	s.species = genetics.NewSpecies(int(numGenes), 3)
	s.population = make([]genetics.Chromosome, popSize)
	for i := 0; i < popSize; i++ {
		s.population[i], _ = s.species.NewRand(s.Rand)
	}

	return nil
}

// Step iterates through count generations of evolution,
// updating the population, score, and best path
func (s *Solver) Step(count int) {
	fitness := make([]genetics.Fitness, len(s.population))
	for i := 0; i < count; i++ {
		for n, c := range s.population {
			p := s.Path(c)
			score := p.Score(s.Map)
			fitness[n] = genetics.Fitness(score)
			if score > s.score {
				s.score = score
				s.best = c
			}
		}
		s.Evolver.Evolve(s.Rand, s.population, fitness)
	}
}

// Score accesses the current best score
func (s *Solver) Score() int {
	return s.score
}

// Best returns the winning chromosome.
func (s *Solver) Best() genetics.Chromosome {
	return s.best
}
