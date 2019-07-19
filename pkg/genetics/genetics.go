package genetics

import (
	"errors"
	"flag"
	"fmt"
)

var (
	crossoverRate = flag.Float64("goldmine.genetic_crossover_rate", 0.7, "The rate at which two genes swap part of their DNA")
	mutationRate  = flag.Int("goldmine.genetic_mutation_rate", 1000, "1 in this number bits will be flipped in a gene during the evolve phase.")
)

// Gene is a single trait to control behavior.
type Gene int8

// Fitness is an arbitrary fitness number based on genomes and their matching traits.
type Fitness int64

// Random implements a subset of math/rand.Rand suitable to be swapped for unit testing.
type Random interface {
	Int31n(n int32) int32
	Int63n(n int64) int64
	Float64() float64
}

const (
	geneBits = 4
	maxGene  = 0xF
)

// private types for interface implementations
type geometric struct{}
type exponential struct{}

var (
	// Geometric growth
	Geometric GeneScorer = geometric{}

	// Exponential growth
//	Exponential GeneScorer = exponential{}
)

// GeneScorer ...
type GeneScorer interface {
	ScoreFitness(trait int64, gene Gene) Fitness
}

func (geometric) ScoreFitness(trait int64, gene Gene) Fitness {
	if gene > maxGene {
		gene = maxGene
	}
	return Fitness(trait * int64(gene))
}

/*
func (exponential) FitnessTrait(trait int64, genome Gene) Fitness {
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
	if genome > maxGene {
		genome = maxGene
	}
	var total Fitness = 1
	for genome != 0 {
		total *= Fitness(trait)
		genome--
	}
	if traitNegative != genomeNegative {
		total = -total
	}
	return total
}
*/

// Chromosome represents a single genetic strategy for a Species.
type Chromosome struct {
	Species *Species
	Genes   []Gene
}

// Species is a factory for all Genes in a repeated evolutionary experiment.
// Separating this from the actual Chromosome allows easier reuse of genetic algorithms
// in multiple circumstances as well as experimentation with the ordering of Chromosomes
// which influences the rate at which they may be separated by crossovers.
type Species struct {
	BitsPerGene uint8
	NumGenes    uint8
	SignedGenes []uint8
	Scorer      GeneScorer
}

// NewSpecies initializes a Species
func NewSpecies(bitsPerGene, numGenes uint8, scorer GeneScorer, signedGenes ...uint8) (*Species, error) {
	// The typedefs for base types must be adjusted if the following numbers are out of line
	if bitsPerGene > 8 {
		return nil, errors.New("Cannot support more than 8 bits per genome")
	}
	if bitsPerGene*numGenes > 64 {
		return nil, errors.New("Cannot support more than 64 bits per chromosome")
	}

	for _, n := range signedGenes {
		if n > numGenes {
			return nil, fmt.Errorf("Signed genome %d is out of range [0, %d)", n, numGenes)
		}
	}
	return &Species{
		BitsPerGene: bitsPerGene,
		NumGenes:    numGenes,
		SignedGenes: signedGenes,
		Scorer:      scorer,
	}, nil
}

// SerializeChromosome creates a uint64 representation of the genomes suitable for serialization across training sessions.
func (s *Species) SerializeChromosome(gene Chromosome) (uint64, error) {
	if len(gene.Genes) != int(s.NumGenes) {
		return 0, fmt.Errorf("Wrong gene count; expected=%d got=%d", s.NumGenes, gene.Genes)
	}

	var serialized uint64
	for _, allele := range gene.Genes {
		serialized = serialized<<(s.BitsPerGene) + uint64(allele)
	}

	return serialized, nil
}

// DeserializeChromosome creates an in-memory representation for Chromosomes encoded with SerializeChromosome
func (s *Species) DeserializeChromosome(serialized uint64) (Chromosome, error) {
	// Precalculate the dynamic mask to use based on BitsPerGene
	unusedBits := uint64(64 - s.BitsPerGene)
	mask := ^uint64(0) // binary !0
	mask = mask >> unusedBits

	// Reverse fill to make bitshifting easier
	alleles := make([]Gene, s.NumGenes)
	for n := int(s.NumGenes) - 1; n >= 0; n-- {
		alleles[n] = Gene(serialized & mask)
		serialized = serialized >> s.BitsPerGene
	}
	if serialized != 0 {
		return Chromosome{}, fmt.Errorf("serialized gene is too long; after deserialization have %x", serialized)
	}
	return Chromosome{
		Genes:   alleles,
		Species: s,
	}, nil
}

// Evolve creates the next generation of genes based on their scores.
// Chromosomes are expected to all be of the Species species.
// The scores and genes variables are expected to be parallel arrays.
// This data structure is used rather than a map[Chromosome]Fitness to help facilitate
// lockless parallel generation of Fitnesss in application code.
func (s *Species) Evolve(evolver Evolver, chromosomes []Chromosome, fitness []Fitness) ([]Chromosome, error) {
	if len(chromosomes) != len(fitness) {
		return nil, fmt.Errorf("chromosomes and fitness scores are different lengths (%d and %d)", len(chromosomes), len(fitness))
	}
	//serializedChromosomes := make([]uint64, 0, len(Chromosome))
	return nil, nil
}

// Evolver mutates a gene pool to create the next generation.
type Evolver interface {
	// Offspring Count determines the number of new Chromosomes that will
	// be created from old Chromosomes, replacing old Chromosomes from the
	// population randomly, though biased by Fitness
	OffpsringCount() int
}

type StandardEvolver struct {
	CrossoverRate float64
	MutationRate  int32
	RNG           Random
}

/*
// OldChromosome syntax:
// [1b: op][4b noopPenalty][4b retracePenalty][4b (signed)distanceWeight][4b valueWeight][4b pickaxeWeight][4b (signed)horizPref][4b (signed)vertPref]
// Rather than signed values having a sign bit, they are adjusted so that
// the midpoint possible is zero. This ensures gradual behavor changes with
// gradual genome changes.
func (g OldChromosome) Strategy() Strategy {
	pickaxeWeight := g.cut()
	valueWeight := g.cut()
	distanceWeight := g.cutSigned()
	retracePenalty := g.cut()
	noopPenalty := g.cut()
	horizPref := g.cutSigned()
	vertPref := g.cutSigned()

	if g != 0 && g != 1 {
		panic("OldChromosome.Strategy() broken")
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

func (s Strategy) OldChromosome() OldChromosome {
	var gene OldChromosome
	switch s.Op {
	case Geometric:
		gene = 0
	case Exponential:
		gene = 1
	}
	gene = (gene << 4) | ((s.VertPref + 0x8) & 0xF)
	gene = (gene << 4) | ((s.HorizPref + 0x8) & 0xF)
	gene = (gene << 4) | (s.NoopPenalty & 0xF)
	gene = (gene << 4) | (s.RetracePenalty & 0xF)
	gene = (gene << 4) | ((s.DistanceWeight + 0x8) & 0xF)
	gene = (gene << 4) | (s.ValueWeight & 0xF)
	gene = (gene << 4) | (s.PickaxeWeight & 0xF)
	return gene
}

func (s Strategy) Fitness(p Path) int64 {
	// TODO: Could these be combined in another variable way?
	return s.Op.FitnessTrait(p.Value(), s.ValueWeight) +
		s.Op.FitnessTrait(p.Pickaxes, s.PickaxeWeight) +
		s.Op.FitnessTrait(p.CityDistance(), s.DistanceWeight) +
		s.Op.FitnessTrait(p.VertDistance(), s.VertPref) +
		s.Op.FitnessTrait(p.HorizDistance(), s.HorizPref) -
		s.Op.FitnessTrait(p.Retraces, s.RetracePenalty) -
		s.Op.FitnessTrait(p.Noops, s.NoopPenalty)
}

// EvanulateOldChromosomeration should instead be part of Goldmine, and not part of genes.
func EvaluateOldChromosomeration(m *Map, genes map[int64]int64) Path {
	var best Path
	for gene := range genes {
		strategy := StrategyFromOldChromosome(gene)
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
		strategy := StrategyFromOldChromosome(problem.which)
		path := problem.m.Mine(strategy)
		problem.genes[problem.which] = path.Value()
		answer <- path
	}
}

// ParallelEvaluateOldChromosomeration will return the best path from genes
// and also set the weigths in the genes parameter.
// This function is highly optimized and is not safe to be called from
// multiple goroutines (though it will itself use many goroutines)
func ParallelEvaluateOldChromosomeration(m *Map, genes map[int64]int64) Path {
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
func Roulette(n int, weights map[OldChromosome]Fitness) []Chromosome {
	var total int64 = 1
	for _, weight := range weights {
		total += weight
	}

	// 1. Pick a number between [0, total)
	// 2. If all weighters were in a line, find the weighter which #1 falls on when adding totals
	// 3. Remove that weighter from the set (by copying over it) & subtract its weight from the wheel
	result := make([]int64, n)
	removed := make(map[OldChromosome]Fitness, n)
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
	crossoverPoint := rand.Int31n(OldChromosomeBits-1) + 1 // 0 isn't a valid crossover.
	highMask := (int64(-1) >> uint(crossoverPoint)) << uint(crossoverPoint)
	lowMask := ^highMask

	aPrime := *a&highMask | *b&lowMask
	bPrime := *b&highMask | *a&lowMask
	*a = aPrime
	*b = bPrime
}

func mutate(gene *int64) {
	for i := uint(0); i <= OldChromosomeBits; i++ {
		if rand.Int31n(int32(*mutationRate)) == 0 {
			continue
		}

		*gene = *gene ^ (int64(1) << i)
	}
}

func Evolve(scoredOldChromosomes map[int64]int64) map[int64]int64 {
	result := make(map[int64]int64, len(scoredOldChromosomes))

	hackGuard := 0
	for len(result) < *generationSize {
		pickTwo := Roulette(2, scoredOldChromosomes)
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
				result[rand.Int63n(1<<OldChromosomeBits)] = 0
				result[rand.Int63n(1<<OldChromosomeBits)] = 0
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
*/
