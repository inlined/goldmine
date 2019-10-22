package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/inlined/genetics"
	"github.com/inlined/rand"

	"github.com/inlined/goldmine/pkg/debug"
	"github.com/inlined/goldmine/pkg/maps"
	"github.com/inlined/goldmine/pkg/solver"

	// Import for flag side-effects
	_ "github.com/inlined/goldmine/pkg/bruteforce"
	_ "github.com/inlined/goldmine/pkg/graph"
)

var (
	selectionFlag genetics.NaturalSelectionFlag
	crossoverFlag genetics.CrossoverFlag
	mutationFlag  genetics.MutationFlag
	solverFlag    solver.Flag

	populationSize   = flag.Int("generation_size", 50, "number of chromosomes in each generation")
	replacementCount = flag.Int("replacement_count", 20, "number of chromosomes to replace each generation")
	mutationRate     = flag.Float64("mutation_rate", 0.02, "the frequency that children will have a mutation")

	input  = flag.String("input", "", "input file or blank for stdin")
	output = flag.String("output", "", "output file or blank for stdout")
)

func init() {
	flag.Var(&selectionFlag, "selection", "algorithm for selecting parents")
	flag.Var(&crossoverFlag, "crossover", "genetic crossover strategy for creating children")
	flag.Var(&mutationFlag, "mutation", "mutations new children may exhibit")
	flag.Var(&solverFlag, "strategy", "Strategy used to solve goldmine maps")
	flag.Var(debug.Flag, "debug", "Debug location or file descriptor; empty to turn off debugging")
}

func main() {
	flag.Parse()

	var err error
	var in io.Reader = os.Stdin
	if *input != "" {
		if in, err = os.Open(*input); err != nil {
			panic(fmt.Sprintf("Unexpected error opening %s: %s", *input, err))
		}
	}
	var out io.WriteCloser = os.Stdout
	if *output != "" {
		if out, err = os.Create(*output); err != nil {
			panic(fmt.Sprintf("Unexpected error opening %s: %s", *output, err))
		}
		defer out.Close()
	}

	evolver := genetics.Evolver{
		ReplacementCount: *replacementCount,
		MutationRate:     float32(*mutationRate),
		Selector:         selectionFlag.Get(),
		Crossover:        crossoverFlag.Get(),
		Mutator:          mutationFlag.Get(),
	}

	var solvers []solver.Solver
	r := maps.NewReader(in)
	var m maps.Map
	for m, err = r.Next(); err == nil; m, err = r.Next() {
		input := solver.Input{
			Evolver: evolver,
			Map:     m,
			Rand:    rand.New(),
		}
		solvers = append(solvers, solverFlag.New(input))
	}

	if err != nil && err != io.EOF {
		panic(fmt.Sprintf("Unexpected error reading maps: %s", err))
	}

	// TODO: Parallel solve each map and add strategy for choosing
	// which map to further investigate.
	const stepCount = 100
	const numSteps = 1000
	const sampleRate = 10
	for i, s := range solvers {
		fmt.Fprintf(debug.Out, "Map %d\n", i)
		if err := s.Init(*populationSize); err != nil {
			panic(fmt.Sprintf("Could not initializes solver:%s", err))
		}
		for x := 0; x < numSteps; x++ {
			s.Step(stepCount)
			if (x+1)%sampleRate == 0 {
				fmt.Fprintf(debug.Out, "%d,", s.Score())
			}
		}
		fmt.Fprintf(debug.Out, "\n%s\n\n", s.Path(s.Best()))
		fmt.Fprintln(out, s.Path(s.Best()))
	}
}
