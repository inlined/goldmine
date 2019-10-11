package main

// Optimization wish list:
// 3. Mutations to resurrect previous generations (esp. successful ones)
// 4. Spend more time on higher scoring boards (another pickaxe in those boards is worth more)
// 7. Should there be stupid path pruning? A->B->A is ok, but A->B->A->B isn't; will the memo catch this?
//    I don't think so because it has guards between generations.

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"inlined.me/goldmine"
)

var (
	crossoverRate = flag.Float64("goldmine.genetic_crossover_rate", 0.7, "The rate at which two genes swap part of their DNA")
	mutationRate  = flag.Int("goldmine.genetic_mutation_rate", 1000, "1 in this number bits will be flipped in a gene during the evolve phase.")

	infile     = flag.String("input", "", "File from which to read (stdin if omitted)")
	outfile    = flag.String("output", "", "File to which to save output; stdout if omitted")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

	debugPrint = flag.Bool("debug", false, "Debug print values to stdout")
	timeout    = flag.Duration("timeout", 10*time.Minute, "Maximum time to run")
	train      = flag.Int("train", 0, "If set, goldmine doesn't solve the maps but finds a gene that does well across all maps. Runs for --train generations.")
)

func getStreams() (in io.ReadCloser, out io.WriteCloser, debug io.Writer) {
	var err error
	in = ioutil.NopCloser(os.Stdin)
	if *infile != "" {
		in, err = os.Open(*infile)
		if err != nil {
			panic(fmt.Sprintf("Could not open input file %s: %s", *infile, err))
		}
	}

	out = os.Stdout
	if *outfile != "" {
		out, err = os.Create(*outfile)
		if err != nil {
			panic(fmt.Sprintf("Could not open output file %s: %s", *outfile, err))
		}
	}

	debug = ioutil.Discard
	if *debugPrint {
		debug = os.Stdout
	}
	return
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	in, out, dbg := getStreams()
	defer in.Close()
	defer out.Close()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	maps, err := goldmine.ReadMaps(in)
	if err != nil {
		panic(fmt.Sprintf("Error reading maps: %s", err))
	}

	if *train == 0 {
		solver := goldmine.NewGeneticSolver(maps)
		solver.RunFor(*timeout)
		for _, path := range solver.BestPaths {
			fmt.Fprintf(dbg, "Best path: %#v\n", path)
			fmt.Fprintln(out, path)
		}
	} else {
		gene, score := goldmine.FindStrongBaseGene(maps, *train)
		fmt.Fprintf(dbg, "Gene %x scored %d points\n", gene, score)
		fmt.Fprintf(out, "0x%x, // %d\n", gene, score)
	}
}
