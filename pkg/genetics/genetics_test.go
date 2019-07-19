package genetics_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/inlined/goldmine/pkg/genetics"
)

func randomRoundtripSerialization(nBits, nGenes uint8) error {
	s, err := genetics.NewSpecies(nBits, nGenes, genetics.Geometric)
	if err != nil {
		return err
	}
	serialized := uint64(rand.Int63n(int64(nBits) * int64(nGenes)))
	chromosome, error := s.DeserializeChromosome(serialized)
	if error != nil {
		return error
	}
	test, error := s.SerializeChromosome(chromosome)
	if error != nil {
		return error
	}
	if test != serialized {
		return fmt.Errorf("Failed to roundtrip chromosome with nBits=%d nGenes=%d; got=%x want=%x", nBits, nGenes, test, serialized)
	}
	return nil
}

func TestSerialization(t *testing.T) {
	for nBits := uint8(1); nBits <= 8; nBits++ {
		// Can't test the top bit due to a lack of rand.uint63n
		for nGenenomes := uint8(1); nGenenomes*nBits < 64; nGenenomes++ {
			for run := 0; run < 100; run++ {
				err := randomRoundtripSerialization(nBits, nGenenomes)
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
