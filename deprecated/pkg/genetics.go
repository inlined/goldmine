package goldmine

import (
	"flag"
	"math/rand"
)

var (
	crossoverRate = flag.Float64("goldmine.genetic_crossover_rate", 0.7, "The rate at which two genes swap part of their DNA")
	mutationRate  = flag.Int("goldmine.genetic_mutation_rate", 1000, "1 in this number bits will be flipped in a gene during the evolve phase.")
)

type OldGene int64
type Genome int8
type Score int64

const (
	maxGenome = 4
)

// private types for interface implementations
type geometric struct{}
type exponential struct{}

var (
	// Geometric growth
	Geometric TraitScorer = geometric{}

	// Exponential growth
	Exponential TraitScorer = exponential{}
)

// TraitScorer ...
type TraitScorer interface {
	ScoreTrait(trait int64, gene Genome) int64
}

func (geometric) ScoreTrait(trait int64, gene Genome) Score {
	if gene > maxGenome {
		gene = maxGenome
	}
	return Score(trait * int64(gene))
}

func (exponential) ScoreTrait(trait int64, genome Genome) Score {
	traitNegative := false
	if trait < 0 {
		traitNegative = true
		trait = -trait
	}
	genomeNegative := false
	if genome < 0 {
		genomeNegative = true
		genome = -genome
	}
	if genome > maxGenome {
		genome = maxGenome
	}
	var total Score = 1
	for genome != 0 {
		total *= Score(trait)
		genome--
	}
	if traitNegative != genomeNegative {
		total = -total
	}
	return total
}

// Strategy determines how the solver behaves.
// If the solver finds many Paths of the same length at the same coordinate, it
// will drop all but Strategy.Keep best elements. When deciding winners, PreferPickaxe
// tunes whether it is better to have a raw score or better to have pickaxes (which means
// a score is likely to climb faster in the future)
type Strategy struct {
	Op TraitScorer

	// Each weight can be 4 bits, plus a sign bit for distance weight
	PickaxeWeight  Genome
	ValueWeight    Genome
	DistanceWeight Genome
	RetracePenalty Genome
	NoopPenalty    Genome
	HorizPref      Genome
	VertPref       Genome
}

var (
	Greedy = Strategy{
		Op:          Geometric,
		ValueWeight: 1,
	}
)

func (g *OldGene) cut() Genome {
	val := Genome(*g & 0xF)
	*g >>= 4
	return val
}

func (g *OldGene) cutSigned() Genome {
	val := Genome(*g&0xF) - 0x8
	*g >>= 4
	return val
}

// OldGene syntax:
// [1b: op][4b noopPenalty][4b retracePenalty][4b (signed)distanceWeight][4b valueWeight][4b pickaxeWeight][4b (signed)horizPref][4b (signed)vertPref]
// Rather than signed values having a sign bit, they are adjusted so that
// the midpoint possible is zero. This ensures gradual behavor changes with
// gradual genome changes.
func (g OldGene) Strategy() Strategy {
	pickaxeWeight := g.cut()
	valueWeight := g.cut()
	distanceWeight := g.cutSigned()
	retracePenalty := g.cut()
	noopPenalty := g.cut()
	horizPref := g.cutSigned()
	vertPref := g.cutSigned()

	if g != 0 && g != 1 {
		panic("OldGene.Strategy() broken")
	}
	op := Geometric
	if g == 1 {
		op = Exponential
	}

	return Strategy{
		Op:             op,
		PickaxeWeight:  pickaxeWeight,
		ValueWeight:    valueWeight,
		DistanceWeight: distanceWeight,
		RetracePenalty: retracePenalty,
		NoopPenalty:    noopPenalty,
		HorizPref:      horizPref,
		VertPref:       vertPref,
	}
}

func (s Strategy) OldGene() OldGene {
	var gene OldGene
	switch s.Op {
	case Geometric:
		gene = 0
	case Exponential:
		gene = 1
	}
	/*
		gene = (gene << 4) | ((s.VertPref + 0x8) & 0xF)
		gene = (gene << 4) | ((s.HorizPref + 0x8) & 0xF)
		gene = (gene << 4) | (s.NoopPenalty & 0xF)
		gene = (gene << 4) | (s.RetracePenalty & 0xF)
		gene = (gene << 4) | ((s.DistanceWeight + 0x8) & 0xF)
		gene = (gene << 4) | (s.ValueWeight & 0xF)
		gene = (gene << 4) | (s.PickaxeWeight & 0xF)
	*/
	return gene
}

func (s Strategy) Score(p Path) int64 {
	// TODO: Could these be combined in another variable way?
	return s.Op.ScoreTrait(p.Value(), s.ValueWeight) +
		s.Op.ScoreTrait(p.Pickaxes, s.PickaxeWeight) +
		s.Op.ScoreTrait(p.CityDistance(), s.DistanceWeight) +
		s.Op.ScoreTrait(p.VertDistance(), s.VertPref) +
		s.Op.ScoreTrait(p.HorizDistance(), s.HorizPref) -
		s.Op.ScoreTrait(p.Retraces, s.RetracePenalty) -
		s.Op.ScoreTrait(p.Noops, s.NoopPenalty)
}

// EvanulateOldGeneration should instead be part of Goldmine, and not part of genes.
func EvaluateOldGeneration(m *Map, genes map[int64]int64) Path {
	var best Path
	for gene := range genes {
		strategy := StrategyFromOldGene(gene)
		path := m.Mine(strategy)
		genes[gene] = path.Value()
		if best.Value() < path.Value() {
			best = path
		}
	}
	return best
}

// This is a bit of a hack, but Goroutines allocate a 2-8KB stack
// depending on which version of Go you're using. Bursting through
// short-lived goroutines isn't really optimial. Profiling showed
// that GC and system interrupts from goexit were a massive drain.
var parallelSolverCount int

type parallelSolverProblem struct {
	m     *Map
	genes map[int64]int64
	which int64
}

var (
	question = make(chan parallelSolverProblem)
	answer   = make(chan Path)
)

func solver() {
	for {
		problem := <-question
		strategy := StrategyFromOldGene(problem.which)
		path := problem.m.Mine(strategy)
		problem.genes[problem.which] = path.Value()
		answer <- path
	}
}

// ParallelEvaluateOldGeneration will return the best path from genes
// and also set the weigths in the genes parameter.
// This function is highly optimized and is not safe to be called from
// multiple goroutines (though it will itself use many goroutines)
func ParallelEvaluateOldGeneration(m *Map, genes map[int64]int64) Path {
	for parallelSolverCount < len(genes) {
		go solver()
		parallelSolverCount++
	}
	for gene := range genes {
		question <- parallelSolverProblem{
			m:     m,
			genes: genes,
			which: gene,
		}
	}

	var best Path
	for range genes {
		res := <-answer
		if best.Value() < res.Value() {
			best = res
		}
	}
	return best
}

// Roulette takes N WeightedIDs and returns n different WeightedIDs with probabilities in proportion to their weight.
func Roulette(n int, weights map[OldGene]Score) []Gene {
	var total int64 = 1
	for _, weight := range weights {
		total += weight
	}

	// 1. Pick a number between [0, total)
	// 2. If all weighters were in a line, find the weighter which #1 falls on when adding totals
	// 3. Remove that weighter from the set (by copying over it) & subtract its weight from the wheel
	result := make([]int64, n)
	removed := make(map[OldGene]Score, n)
	defer func() {
		for id, weight := range removed {
			weights[id] = weight
		}
	}()

	for i := 0; i < n; i++ {
		target := rand.Int63n(total)
		var accum int64
		for id, weight := range weights {
			accum += weight
			if accum >= target {
				result[i] = id
				total -= weight
				delete(weights, id)
				removed[id] = weight
				break
			}
		}
	}
	return result
}

func crossover(a *int64, b *int64) {
	crossoverPoint := rand.Int31n(OldGeneBits-1) + 1 // 0 isn't a valid crossover.
	highMask := (int64(-1) >> uint(crossoverPoint)) << uint(crossoverPoint)
	lowMask := ^highMask

	aPrime := *a&highMask | *b&lowMask
	bPrime := *b&highMask | *a&lowMask
	*a = aPrime
	*b = bPrime
}

func mutate(gene *int64) {
	for i := uint(0); i <= OldGeneBits; i++ {
		if rand.Int31n(int32(*mutationRate)) == 0 {
			continue
		}

		*gene = *gene ^ (int64(1) << i)
	}
}

func Evolve(scoredOldGenes map[int64]int64) map[int64]int64 {
	result := make(map[int64]int64, len(scoredOldGenes))

	hackGuard := 0
	for len(result) < *generationSize {
		pickTwo := Roulette(2, scoredOldGenes)
		gene0 := pickTwo[0]
		gene1 := pickTwo[1]
		if rand.Float64() < *crossoverRate {
			crossover(&gene0, &gene1)
		}
		// TODO: there's got to be a muuuuch faster way to do this.
		mutate(&gene0)
		mutate(&gene1)

		// We want to ignore candidacy after any crossovers and mutations are applied because
		// successful genes (and their mutations) are *supposed* to be overrepresented in
		// results, just not as a duplicate.
		// In the case that we keep getting the same result many times, it likely means we're
		// just so weighted towards successful genes. Introduce something random to kick start things.
		liveLocked := true
		if _, ok := result[gene0]; !ok {
			result[gene0] = 0
			liveLocked = false
		}
		if _, ok := result[gene1]; !ok {
			result[gene1] = 0
			liveLocked = false
		}

		if liveLocked {
			hackGuard++
			if hackGuard >= 50 {
				result[rand.Int63n(1<<OldGeneBits)] = 0
				result[rand.Int63n(1<<OldGeneBits)] = 0
				hackGuard = 0
				// Other ideas: force mutation?
				// Why is this happening enough to live-lock? is the crossover rate to low? crossoverPoint can't be 0?
				// Is one value super high and the rest all 1s?
			}
			continue
		}
	}

	return result
}
