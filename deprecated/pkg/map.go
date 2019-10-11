package behavioral

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	// MapStart ...
	MapStart = '='

	// Start ...
	Start = 's'

	// Space ...
	Space = '.'

	// Wall ...
	Wall = 'w'

	// Pickaxe ...
	Pickaxe = 'd'
)

// Vertex ...
type Vertex struct {
	Row int
	Col int
}

func (v Vertex) String() string {
	return fmt.Sprintf("(%d,%d)", v.Row, v.Col)
}

// Map is a non-reduced representation of map data
type Map struct {
	MaxMoves int
	StartAt  Vertex
	Data     [][]byte
}

// String returns a map in the same format as its input with some debug info at the top.
func (m Map) String() string {
	accum := fmt.Sprintf("Map with %d moves starting at %s\n", m.MaxMoves, m.StartAt)
	for _, line := range m.Data {
		accum = accum + string(line) + "\n"
	}
	return accum
}

// ReadMaps reads many maps from an io.Reader until an EoF is returned.
func ReadMaps(reader io.Reader) ([]Map, error) {
	var val []Map
	buffered := bufio.NewReader(reader)
	for {
		sub, err := readOne(buffered)
		if err == io.EOF {
			return val, nil
		}
		if err != nil {
			return nil, err
		}
		val = append(val, *sub)
	}
}

func readOne(reader *bufio.Reader) (*Map, error) {
	var m Map
	var width, height int
	foundStart := false

	header, err := reader.ReadString('\n')
	if err != nil {
		return nil, err // err can be io.EOF and that's OK
	}

	count, err := fmt.Sscanf(header, "=%d,%d,%d\n", &height, &width, &m.MaxMoves)
	if err != nil {
		return nil, err
	}
	if count != 3 {
		return nil, errors.New("Expected the first line to be '=<height>,<width>,<moves>'")
	}

	for rowNo := 0; rowNo < height; rowNo++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("Failed to read %dth row", rowNo)
		}
		// ReadString includes the delim
		line = line[0 : len(line)-1]
		if len(line) != width {
			return nil, fmt.Errorf("Expected rows to be %d wide, row %d is %d wide (%s)", width, rowNo, len(line), line)
		}
		m.Data = append(m.Data, []byte(line))

		start := strings.IndexByte(line, Start)
		if start == -1 {
			continue
		}
		startAt := Vertex{Row: rowNo, Col: start}
		if foundStart {
			return nil, fmt.Errorf("Found multiple starting points: %s and %s", m.StartAt, startAt)
		}
		m.StartAt = startAt
	}

	// In production code I'd also sanitize to make sure all characters were valid and we had
	// a starting point.

	return &m, nil
}

func (m *Map) canReach(v Vertex) bool {
	return v.Row >= 0 && v.Row < len(m.Data) && v.Col >= 0 && v.Col < len(m.Data[0]) && m.Data[v.Row][v.Col] != Wall
}

/*
// Mine makes one attempt to solve a map using a given strategy to resolve conflicts
func (m *Map) Mine(strategy Strategy) Path {
	paths := []Path{
		{
			Map:   m,
			Steps: []Vertex{m.StartAt},
		},
	}

	for round := 0; round < m.MaxMoves; round++ {
		scoredPaths := make(map[Vertex]ScoredPath, len(paths)*4)

		// On the last round, we just want the single highest score
		if round == m.MaxMoves-1 {
			strategy = Greedy
		}

		for _, path := range paths {
			from := path.Dest()
			tests := []Vertex{
				Vertex{Col: from.Col - 1, Row: from.Row},
				Vertex{Col: from.Col + 1, Row: from.Row},
				Vertex{Col: from.Col, Row: from.Row - 1},
				Vertex{Col: from.Col, Row: from.Row + 1},
			}

			for _, test := range tests {
				if !m.canReach(test) {
					continue
				}
				newPath := path.Push(test)
				score := strategy.Score(newPath)
				if existing, ok := scoredPaths[test]; !ok || existing.Score < score {
					scoredPaths[test] = ScoredPath{
						Path:  newPath,
						Score: score,
					}
				}
			}
		}

		// Collapse the best of this round into paths
		// Edge case: A start surrounded by four walls will have no best.
		newPaths := collect(scoredPaths)
		if len(newPaths) == 0 {
			break
		}
		paths = newPaths
	}

	// Now we've gone through all possible rounds; find the winner:
	return bestPath(paths)
}
*/
