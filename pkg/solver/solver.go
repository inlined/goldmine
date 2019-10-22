package solver

import (
	"fmt"

	"github.com/inlined/genetics"
	"github.com/inlined/rand"

	"github.com/inlined/goldmine/pkg/maps"
)

// This is a pretty bad violation of my normal rules against
// globals. Is there a better way?
var (
	factories = make(map[string]func(i Input) Solver)
)

// Solver is the generic interface for all solvers
type Solver interface {
	Init(popSize int) error
	Step(count int)
	Path(genetics.Chromosome) maps.Path
	Score() int
	Best() genetics.Chromosome
}

// Input is used to create a solver
type Input struct {
	Map     maps.Map
	Evolver genetics.Evolver
	Rand    rand.Rand
}

// Flag allows developers to specify a Solver via
// flag and create instances with New()
type Flag string

func (f Flag) String() string {
	if f == "" {
		return "bruteforce"
	}
	return string(f)
}

// Set implements flag.Value
func (f *Flag) Set(s string) error {
	if _, ok := factories[s]; !ok {
		return fmt.Errorf("solver.Flag.Set(%s) unknown solver %s", s, s)
	}
	*f = Flag(s)

	return nil
}

// New creates a new Solver with solver.Input
func (f Flag) New(i Input) Solver {
	name := f.String()
	return factories[name](i)
}

// RegisterSolverFlag is to be called in a solver package's init()
// function so that solver.Flag can include that solver in the parser.
func RegisterSolverFlag(flag string, f func(Input) Solver) {
	if _, ok := factories[flag]; ok {
		panic(fmt.Sprintf("Double registering factory %s", flag))
	}
	factories[flag] = f
}
