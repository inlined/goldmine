package maps

/*
type Path struct {
	Steps    []Vertex
	Map      *Map
	value    int64
	Pickaxes int64
	Noops    int64
	Retraces int64
}

const (
	Left  = 'l'
	Right = 'r'
	Down  = 'd'
	Up    = 'u'
)

// GoString is the format used when printing with the %v format command;
// this will provide a better debug print.
func (p Path) GoString() string {
	return fmt.Sprintf("Path: %s, Value: %d", p.Steps, p.Value())
}

func dir(from, to Vertex) byte {
	if to.Row == from.Row+1 {
		return Down
	} else if to.Row == from.Row-1 {
		return Up
	} else if to.Col == from.Col+1 {
		return Right
	} else if to.Col == from.Col-1 {
		return Left
	} else {
		panic(fmt.Sprintf("Invalid path transition %s -> %s", from, to))
	}
}

// When printing a Path directly this will print the path using udrl characters.
func (p Path) String() string {
	if len(p.Steps) <= 1 {
		return ""
	}

	buff := make([]byte, 0, len(p.Steps)-1)
	for i := 1; i < len(p.Steps); i++ {
		buff = append(buff, dir(p.Steps[i-1], p.Steps[i]))
	}
	return string(buff)
}

// Dest is the vertex which a path leads to. It is used to determine which next edges are
// valid and to determine which other contender Paths might need to be culled to keep the
// search space small.
func (p *Path) Dest() Vertex {
	return p.Steps[len(p.Steps)-1]
}

// Value is the number of points earned by a particular path.
func (p *Path) Value() int64 {
	return p.value
}

func (p *Path) CityDistance() int64 {
	first := p.Steps[0]
	last := p.Dest()
	height := first.Col - last.Col
	if height < 0 {
		height = -height
	}
	width := first.Row - last.Row
	if width < 0 {
		width = -width
	}
	return int64(height + width)
}

func (p *Path) HorizDistance() int64 {
	return int64(p.Steps[0].Col - p.Dest().Col)
}

func (p *Path) VertDistance() int64 {
	return int64(p.Steps[0].Row - p.Dest().Row)
}

// Push will add a vertex to the path
// If this spot has been visited before it won't add to the value.
func (p *Path) Push(v Vertex) Path {
	hasVisited := false
	for _, existing := range p.Steps {
		if existing == v {
			hasVisited = true
		}
	}

	sprite := p.Map.Data[v.Row][v.Col]

	res := *p
	res.Steps = make([]Vertex, len(p.Steps)+1)
	copy(res.Steps, p.Steps)
	res.Steps[len(p.Steps)] = v
	if !hasVisited {
		if sprite == Pickaxe {
			res.Pickaxes++
		} else if sprite >= '1' && sprite <= '9' {
			earned := int64(sprite-'0') << uint(res.Pickaxes)
			res.value += earned
		} else {
			res.Noops++
		}
	} else {
		res.Retraces++
	}
	return res
}

func bestPath(options []Path) Path {
	var best *Path
	for ndx := range options {
		if best == nil || best.Value() < options[ndx].Value() {
			best = &options[ndx]
		}
	}
	return *best
}

type ScoredPath struct {
	Score int64
	Path  Path
}

func collect(m map[Vertex]ScoredPath) []Path {
	var accum []Path
	for ndx := range m {
		accum = append(accum, m[ndx].Path)
	}
	return accum
}
*/
