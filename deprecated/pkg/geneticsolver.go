package goldmine

import (
	"flag"
	"fmt"
	"math/rand"
	"time"
)

var (
	strategy       = flag.Int("goldmine.map_pick_strategy", 1, "Which strategy to use to pick a map")
	generationSize = flag.Int("goldmine.generation_size", 20, "The number of strategies per generation")
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type environment struct {
	Generation int64
	GenePool   map[int64]int64
}

type geneticSolver struct {
	maps         []Map
	BestPaths    []Path
	environments []environment
}

var seedGenes = []int64{
	0x1802920d, //28
	0x09a3a00d, //27
	0x1121970a, //27
	0x1121910b, //27
	0x10e5d70a, //27
	0x00e5d70a, //27
	0x10e5d70a, //27.6
	0x11219103, //26.7
	0x28156609, //28
	0x0cc6e10f, //28
	0x1d63e10f, //27.8
	0x1761760f, //27.8
	0x16e05107, //28
	0x05026d0d, //28.5
	0x0583aa0f, //28.5
}

// MultiMine tries to solve many maps with as many strategies as can be allowed within
// a timeout.
func NewGeneticSolver(maps []Map) *geneticSolver {
	environments := make([]environment, len(maps))
	for i := range environments {
		environments[i].GenePool = make(map[int64]int64)
		for _, seed := range seedGenes {
			environments[i].GenePool[seed] = 0
		}
		for add := len(seedGenes); add < *generationSize; add++ {
			environments[i].GenePool[rand.Int63n(1<<GeneBits)] = 0
		}
	}

	return &geneticSolver{
		maps:         maps,
		BestPaths:    make([]Path, len(maps)),
		environments: environments,
	}
}

type update struct {
	id   int
	best Path
}

func (m *geneticSolver) RunFor(duration time.Duration) {
	updates := make(chan update, 1)

	go m.simulateEvolution(updates)

	// consumer
	deadline := time.After(duration * 98 / 100)
	var done = false
	for !done {
		select {
		case <-deadline:
			done = true
		case update, ok := <-updates:
			if !ok { // All solvers have quit
				done = true
				break
			}
			m.BestPaths[update.id] = update.best
		}
	}
}

func (m *geneticSolver) simulateEvolution(updates chan update) {
	mapWeights := make(map[int64]int64)
	mapScores := make(map[int]int64)
	for ndx := range m.maps {
		mapWeights[int64(ndx)] = 9999999
	}
	for {
		pick := Roulette(1, mapWeights)
		i := int(pick[0])
		priorValue := mapScores[i]
		path := m.simulateGeneration(i)
		newValue := path.Value()
		bestValue := priorValue
		if newValue > priorValue {
			fmt.Printf("Generation %d of map %d earned %d more points\n", m.environments[i].Generation, i, newValue-priorValue)
			updates <- update{id: i, best: path}
			mapScores[i] = newValue
			bestValue = newValue
		}

		var newWeight int64
		switch *strategy {
		case 1:
			newWeight = bestValue
		case 2:
			newWeight = bestValue / m.environments[i].Generation
		case 3:
			newWeight = bestValue * bestValue / m.environments[i].Generation
		}
		mapWeights[int64(i)] = newWeight

	}
}

func (m *geneticSolver) simulateGeneration(i int) Path {
	path := ParallelEvaluateGeneration(&m.maps[i], m.environments[i].GenePool)
	m.environments[i].Generation++
	m.environments[i].GenePool = Evolve(m.environments[i].GenePool)
	return path
}
