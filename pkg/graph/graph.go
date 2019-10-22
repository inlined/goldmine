package graph

import (
	"fmt"

	"github.com/inlined/genetics"
	"github.com/inlined/goldmine/pkg/maps"
	"github.com/inlined/goldmine/pkg/solver"
)

func init() {
	solver.RegisterSolverFlag("graph", func(i solver.Input) solver.Solver {
		return &Solver{Input: i}
	})
}

// Solver reduces a map into a connectivity graph
// via Init() and then steps through permutations.
type Solver struct {
	solver.Input
	paths      [][]maps.Path
	species    *genetics.Species
	population []genetics.Chromosome
	best       genetics.Chromosome
	score      int
}

// Init creates the genetic components needed to solve a map and
// reduces it to a graph.
func (s *Solver) Init(popSize int) error {
	// -1 because we will always start at PoI[0]
	nodes := len(s.Map.PointsOfInterest) - 1
	s.species = genetics.NewSpecies(nodes, nodes-1)
	s.population = make([]genetics.Chromosome, 0, popSize)
	for i := 0; i < popSize; i++ {
		c, err := s.species.NewPerm(s.Rand)
		if err != nil {
			return err
		}
		s.population = append(s.population, c)
	}

	s.paths = make([][]maps.Path, len(s.Map.PointsOfInterest))
	poiLookup := make(map[maps.Vertex]int)
	for x, poi := range s.Map.PointsOfInterest {
		poiLookup[poi] = x
	}

	for x, v := range s.Map.PointsOfInterest {
		s.paths[x] = connectivityGraph(s.Map, poiLookup, v)
	}

	sum := 0
	for _, x := range s.paths {
		for _, p := range x {
			if p != nil {
				sum++
			}
		}
	}

	fmt.Printf("Map has %d points of interest and %d meaningful paths\n", len(s.Map.PointsOfInterest), sum)

	return nil
}

// Path exposes how this Solver would create a Path from a given Chromosome.
func (s Solver) Path(c genetics.Chromosome) maps.Path {
	p := maps.Path(make([]maps.Direction, 0, s.Map.StepsAllowed))

	from := 0
	for _, g := range c.Genes {
		to := int(g + 1) // +1 because poi[0] isn't a valid gene
		candidate := s.paths[from][to]
		if candidate == nil || candidate.Len()+p.Len() > s.Map.StepsAllowed {
			continue
		}
		p.Concat(candidate)
		from = to
	}

	// In case we run out of valid genes before StepsAllowed
	p.Pad(s.Map)

	return p
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

		// Due to the massive search space, we need to inject new genes as a sort of
		// alleling. Evolver currently doesn't have the right abstractions necessary
		// to do age-out selection (we always kill off the weakest genes). We'll simulate
		// age out selection by reeinitialzing one random gene every cycle.
		victim := int(s.Rand.Int31n(int32(len(s.population))))
		s.population[victim], _ = s.species.NewPerm(s.Rand)
	}
}

// Score accesses the current best score
func (s *Solver) Score() int {
	return s.score
}

// Best reveals the winning Chromosome
func (s *Solver) Best() genetics.Chromosome {
	return s.best
}

// connectivityGraph finds all the valid paths to all points of interest starting
// at a particular vertex. If a path cannot be found within the map's maximum, the
// slot for that Path is nil.
func connectivityGraph(m maps.Map, poiLookup map[maps.Vertex]int, from maps.Vertex) []maps.Path {
	type tie struct {
		maps.Path
		maps.Vertex
	}
	res := make([]maps.Path, len(m.PointsOfInterest))

	// TODO: use queues to reuse memory

	// This is a modified version of A*, except we don't need
	// to repeatedly check for shortest paths. We're constantly
	// increasing city-walk distance, so the first path that
	// finds a vertex is a shortest path. We can also cut the search
	// short as soon as we've used up our maximum number of paths.
	paths := []tie{{Path: nil, Vertex: from}}
	seen := make([]bool, m.Rows()*m.Cols())

	seen[from.Row*m.Cols()+from.Col] = true
	for dist := 0; dist < m.StepsAllowed; dist++ {
		var next []tie
		for _, t := range paths {
			for _, d := range []maps.Direction{maps.Up, maps.Down, maps.Left, maps.Right} {
				v2 := t.Vertex.Move(d)
				if !m.CanBeAt(v2) {
					continue
				}

				if seen[v2.Row*m.Cols()+v2.Col] {
					continue
				}
				seen[v2.Row*m.Cols()+v2.Col] = true

				p2 := t.Path.Push(d)
				if m.At(v2) != maps.Space {
					poi := poiLookup[v2]
					res[poi] = p2
					continue
				}

				next = append(next, tie{p2, v2})
			} // for each direction
		} // for each existing path
		paths = next

		// TODO: is it worth an early exit if we find all valid PoI?
	} // for maximum steps

	return res
}
