package maps

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	// Points of interest:

	Start   = 's'
	Space   = '.'
	Wall    = 'w'
	Pickaxe = 'd'

	header = '='

	// Directions for paths:

	Up    = 'u'
	Down  = 'd'
	Left  = 'l'
	Right = 'r'
)

var (
	InvalidVertex = Vertex{-1, -1}
)

// Direction is an individual u, d, l, or r
type Direction = byte

// Vertex is a 2d point in a map
type Vertex struct {
	Row int
	Col int
}

func (v Vertex) String() string {
	return fmt.Sprintf("(%d, %d)", v.Row, v.Col)
}

// Valid indicates whether this is a valid position on the board
func (v Vertex) Valid() bool {
	return v.Row >= 0 && v.Col >= 0
}

func (v Vertex) Move(d Direction) Vertex {
	switch d {
	case Up:
		v.Row -= 1
	case Down:
		v.Row += 1
	case Left:
		v.Col -= 1
	case Right:
		v.Col += 1
	}

	return v
}

// Map is a parsed representation of a Goldmine map
type Map struct {
	// Cells holds the uncompressed meaning of the entire goldmine map
	Cells [][]byte

	// PointsOfInterest allows quick indexing of the map to see
	// non-space and non-wall points.
	// PointsOfInterest[0] will always be the starting position
	PointsOfInterest []Vertex

	// StepsAllowed is the number of steps the user is allowed to
	// take in the game
	StepsAllowed int
}

// String returns a map in the same format as its input with some debug info at the top.
func (m Map) String() string {
	height, width := len(m.Cells), len(m.Cells[0])
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("=%d,%d,%d\n", height, width, m.StepsAllowed))
	b.Grow(height * (width + 1))
	for row := 0; row < height; row++ {
		b.Write(m.Cells[row])
		b.WriteByte('\n')
	}
	return b.String()
}

// At gets the board piece at a given vertex
func (m Map) At(v Vertex) byte {
	return m.Cells[v.Row][v.Col]
}

// Rows ...
func (m Map) Rows() int {
	return len(m.Cells)
}

// Cols ...
func (m Map) Cols() int {
	return len(m.Cells[0])
}

// CanBeAt returns whether the vertex is a valid location
// and not a wall
func (m Map) CanBeAt(v Vertex) bool {
	inBounds := v.Row >= 0 && v.Col >= 0
	inBounds = inBounds && v.Row < m.Rows() && v.Col < m.Cols()
	return inBounds && m.At(v) != Wall
}

// Interesting declares whether v is a point of interest.
func (m Map) Interesting(v Vertex) bool {
	return m.CanBeAt(v) && m.At(v) != Space
}

// Reader parses a map from an io.Reader
type Reader struct {
	reader *bufio.Reader
}

// NewReader creates a new maps.Reader
func NewReader(reader io.Reader) Reader {
	return Reader{
		reader: bufio.NewReader(reader),
	}
}

// a wrapper function designed to trim the results of ReadBytes in place
func trimInPlace(b []byte, err error) ([]byte, error) {
	if err != nil && err != io.EOF {
		return nil, err
	}
	// Remove spacing on the left (common in tests) and the delim on the right
	// Can't use bytes.Trim because this edits the underlying buffer
	skip := bytes.LastIndexAny(b, " \t")
	b = b[skip+1:]
	if len(b) == 0 {
		return nil, errors.New("empty row")
	}
	if b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}

	return b, nil
}

func sanitizeCells(m *Map) error {
	m.PointsOfInterest = make([]Vertex, 1)
	foundStart := false

	for ir, row := range m.Cells {
		for ic, val := range row {
			v := Vertex{Row: ir, Col: ic}
			switch val {
			case Wall, Space:
				continue
			case Start:
				if foundStart {
					return fmt.Errorf("Reader.Next(): found second starting location at %s", v)
				}
				foundStart = true
				m.PointsOfInterest[0] = v
			default:
				if val < '0' || val > '9' {
					return fmt.Errorf("Reader.Next(): unknown rune %c at column %d of row %d (%s)", val, v.Col, v.Row, string(row))
				}
				fallthrough
			case Pickaxe:
				m.PointsOfInterest = append(m.PointsOfInterest, v)
			}
		}
	}

	if !foundStart {
		return errors.New("Reader.Next(): did not find starting location")
	}
	return nil
}

// Next reads the next map from the io.Reader. Returns io.EOF if
// the io.Reader has terminated.
func (r *Reader) Next() (Map, error) {
	var m Map
	var header string
	var err error

	// Ignore leading newlines (allows repeated map entries to be newline delimited)
	// Should also ensure trailing newlines still cause io.EOF at the end of a buffer.
	for {
		header, err = r.reader.ReadString('\n')
		if err != nil {
			return m, err // err can be io.EOF and that's OK
		}
		header = strings.TrimLeft(header, " \t")
		if header != "\n" {
			break
		}
	}

	var width, height int
	count, err := fmt.Sscanf(header, "=%d,%d,%d\n", &height, &width, &m.StepsAllowed)
	if err != nil || count != 3 {
		return m, errors.New("Reader.Next(): expected header of '=<height>,<width>,<moves>'")
	}
	// Sanitize that all header values must be >= 1?

	// Start with size 1 because we'll insert the start location but otherwise append
	m.Cells = make([][]byte, height)
	for row := 0; row < height; row++ {
		m.Cells[row], err = trimInPlace(r.reader.ReadBytes('\n'))
		if err != nil && err != io.EOF {
			return m, fmt.Errorf("Reader.Next(): failed to read row %d", row)
		}
		if len(m.Cells[row]) != width {
			return m, fmt.Errorf("Reader.Next(): row %d (%s); expected %d columns got %d", row, string(m.Cells[row]), width, len(m.Cells[row]))
		}
	}

	if err = sanitizeCells(&m); err != nil {
		return m, err
	}
	return m, nil
}

// Path is a series of Directions
type Path []Direction

// ParsePath *directly casts* a string into a path; no error checking
// is performed
func ParsePath(s string) Path {
	return []Direction(s)
}

func (p Path) String() string {
	return string([]Direction(p))
}

// Len is the lenght of the path
func (p Path) Len() int {
	return len(p)
}

// Append adds a Direction to the path in place
func (p *Path) Append(d Direction) {
	*p = append(*p, d)
}

// Push is like append but creates a new path
func (p Path) Push(d Direction) Path {
	var p2 Path
	p2 = append(p2, p...)
	p2 = append(p2, d)
	return p2
}

// Concat adds another pat onto the end of this one
func (p *Path) Concat(p2 Path) {
	*p = append(*p, p2...)
}

// Copy creates another path to work with
func (p Path) Copy() Path {
	var p2 Path
	p2 = append([]Direction(nil), p...)
	return p2
}

// EndingVertex returns the final location after following a path on a map
// If the path is invalid, returns InvalidVertex
func (p Path) EndingVertex(m Map) Vertex {
	v := m.PointsOfInterest[0]
	for _, d := range p {
		v = v.Move(d)
		if !m.CanBeAt(v) {
			return InvalidVertex
		}
	}
	return v
}

// Pad adds random moves to p that are valid according to m until
// it has the required number of steps in m
func (p *Path) Pad(m Map) {
	v := p.EndingVertex(m)
	var d1, d2 Direction
	if m.CanBeAt(v.Move(Up)) {
		d1 = Up
		d2 = Down
	} else if m.CanBeAt(v.Move(Down)) {
		d1 = Down
		d2 = Up
	} else if m.CanBeAt(v.Move(Left)) {
		d1 = Left
		d2 = Right
	} else {
		d1 = Right
		d2 = Left
	}
	// Must be able to go one of the four directions

	for p.Len() < m.StepsAllowed {
		p.Append(d1)
		d1, d2 = d2, d1
	}
}

// Score generates the Goldmine game score that a path
// would have given a map
func (p *Path) Score(m Map) int {
	seen := make([]bool, m.Rows()*m.Cols())
	score := 0
	pickaxes := uint(0)
	v := m.PointsOfInterest[0]
	for _, d := range *p {
		v = v.Move(d)
		if !m.CanBeAt(v) {
			return 0
		}

		i := v.Row*m.Cols() + v.Col
		if seen[i] {
			continue
		}
		seen[i] = true

		switch x := m.At(v); x {
		case Wall:
			return 0 // cannot happen; CanBeAt(v)
		case Space, Start:
			continue
		case Pickaxe:
			pickaxes++
		default:
			earned := int(x - '0')
			earned = earned << pickaxes
			score += earned
		}
	}
	return score
}
