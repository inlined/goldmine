package goldmine

import "math/rand"

// A trainer helps come up with good seed genes. It is considered
// poor spirit to do this on the final map, so the recommended
// usage is:
// for i in `seq X`; do java -jar Goldmine generate 1000 tempfile; goldmine --train 1000 --input tempfile --output gene$i; done

func FindStrongBaseGene(maps []Map, generations int) (gene int64, score int64) {
	genePool := make(map[int64]int64)
	for add := 0; add < *generationSize; add++ {
		genePool[rand.Int63n(1<<GeneBits)] = 0
	}

	var bestGene int64
	var bestScore int64
	for generation := 0; generation < generations; generation++ {
		aggregate := make(map[int64]int64, len(maps))
		for _, aMap := range maps {
			ParallelEvaluateGeneration(&aMap, genePool)
			for gene, weight := range genePool {
				if weight > bestScore {
					bestScore = weight
					bestGene = gene
				}
				aggregate[gene] += weight
			}
		}
		genePool = Evolve(genePool)
	}
	return bestGene, bestScore
}
