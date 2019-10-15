package maps_test

import (
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/inlined/goldmine/pkg/maps"
)

// normalize turns a test string into the output format expected
// by a map seraizliation. This allows a sort of whitespace
// insensitive compare.
func normalize(m string) string {
	trimmed := strings.ReplaceAll(m, " ", "")
	trimmed = strings.ReplaceAll(m, "\t", "")
	if trimmed[len(trimmed)-1] != '\n' {
		trimmed = trimmed + "\n"
	}
	return trimmed
}
func coordinates(rowsAndCols ...int) []maps.Vertex {
	v := make([]maps.Vertex, len(rowsAndCols)/2)
	for i := 0; i < len(v); i++ {
		v[i] = maps.Vertex{rowsAndCols[i*2], rowsAndCols[i*2+1]}
	}

	return v
}

func TestMaps(t *testing.T) {
	for _, test := range []struct {
		tag   string
		m     string
		err   string
		poi   []maps.Vertex
		steps int
	}{
		{
			tag: "parse and serialize",
			m: `=2,2,1
				ws
				1.`,
			poi:   coordinates(0, 1, 1, 0),
			steps: 1,
		}, {
			// poi[0] must always be start
			tag: "poi out of order",
			m: `=2,2,2
				wd
				.s`,
			poi:   coordinates(1, 1, 0, 1),
			steps: 2,
		}, {
			tag: "col mismatch",
			m: `=2,2,2
				w
				.s`,
			err: "Reader.Next(): row 0 (w); expected 2 columns got 1",
		}, {
			tag: "missing header",
			m: `w.
				.s`,
			err: "Reader.Next(): expected header of '=<height>,<width>,<moves>'",
		}, {
			tag: "malformed header",
			m: `=2,2,2w.
				.s`,
			err: "Reader.Next(): expected header of '=<height>,<width>,<moves>'",
		}, {
			tag: "row mismatch",
			m: `=2,2,2
				w.`,
			err: "Reader.Next(): failed to read row 1",
		}, {
			tag: "no start",
			m: `=2,2,2
				w.
				..`,
			err: "Reader.Next(): did not find starting location",
		}, {
			tag: "more than one start",
			m: `=2,2,2
				w.
				ss`,
			err: "Reader.Next(): found second starting location at (1, 1)",
		}, {
			tag: "invalid rune",
			m: `=2,2,2
				fi
				re`,
			err: "Reader.Next(): unknown rune f at column 0 of row 0 (fi)",
		},
	} {
		t.Run(test.tag, func(t *testing.T) {
			r := maps.NewReader(strings.NewReader(test.m))
			parsed, err := r.Next()
			if test.err != "" {
				if err == nil {
					t.Fatalf("maps.Reader.Next(): expected err %s; got nothing", test.err)
				}
				if err.Error() != test.err {
					t.Fatalf("maps.Reader.Next(): expected err %s; got %s", test.err, err)
				}

				// End of tests if an error was expected
				return
			} else if err != nil {
				t.Fatalf("maps.Reader.Next(): failed with err %s", err)
			}

			if diff := cmp.Diff(parsed.PointsOfInterest, test.poi); test.poi != nil && diff != "" {
				t.Errorf("maps.Reader.Next() returned wrong points of interest. got=%v; Want=%v; diff=%s", parsed.PointsOfInterest, test.poi, diff)
			}
			if test.steps != 0 && test.steps != parsed.StepsAllowed {
				t.Errorf("maps.Reader.Next(): expected %d steps, got %d", test.steps, parsed.StepsAllowed)
			}

			s := parsed.String()
			n := normalize(test.m)
			if s != n {
				t.Errorf("Round trip serialization failed; got=%s; want=%s; diff=%s", s, n, cmp.Diff(s, n))
			}
		})
	}
}

func TestReadRepeatedly(t *testing.T) {
	s := `=2,2,2
		  ..
		  s.
		  
		 =1,1,1
		 s
		 
		 `
	r := maps.NewReader(strings.NewReader(s))

	expected1 := maps.Map{
		PointsOfInterest: coordinates(1, 0),
		StepsAllowed:     2,
		Cells: [][]byte{
			{'.', '.'},
			{'s', '.'},
		},
	}

	expected2 := maps.Map{
		PointsOfInterest: coordinates(0, 0),
		StepsAllowed:     1,
		Cells:            [][]byte{{'s'}},
	}

	m, err := r.Next()
	if err != nil {
		t.Errorf("maps.Reader.Next(); failed with err=%s", err)
	}
	if diff := cmp.Diff(m, expected1); diff != "" {
		t.Errorf("Did not get expected map; diff=%s", diff)
	}

	m, err = r.Next()
	if err != nil {
		t.Errorf("maps.Reader.Next(); failed with err=%s", err)
	}
	if diff := cmp.Diff(m, expected2); diff != "" {
		t.Errorf("Did not get expected map; diff=%s", diff)
	}

	_, err = r.Next()
	if err != io.EOF {
		t.Errorf("Expected EOF error; got %s", err)
	}
}

func TestPathManipulation(t *testing.T) {
	p := maps.ParsePath("uuddlrlr")
	if p.String() != "uuddlrlr" {
		t.Errorf("Round trip parse and encode got %s; expected uuddlrlr", p.String())
	}

	p2 := p.Concat('u')
	p.Append('u')
	if p.String() != p2.String() {
		t.Errorf("Concat vs Append 'd'; got %s vs %s", p2, p)
	}
}

func TestPathTraversal(t *testing.T) {
	s := `=3,5,5
 	      w...1
	      ..s..
	      2d1..`
	r := maps.NewReader(strings.NewReader(s))
	m, err := r.Next()
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		tag      string
		path     string
		expected maps.Vertex
		score    int
	}{
		{
			tag:      "nothing walk",
			path:     "rrdlr",
			expected: maps.Vertex{2, 4},
			score:    0,
		}, {
			tag:      "one point walk",
			path:     "rrull",
			expected: maps.Vertex{0, 2},
			score:    1,
		}, {
			tag:      "pickaxe at end",
			path:     "dlulr",
			expected: maps.Vertex{1, 1},
			score:    1,
		}, {
			tag:      "pickaxe before points",
			path:     "ldlurr",
			expected: maps.Vertex{1, 2},
			score:    4,
		}, {
			tag:      "pickaxe mid points",
			path:     "dllur",
			expected: maps.Vertex{1, 1},
			score:    5,
		}, {
			tag:      "repeats aren't counted",
			path:     "dllrr",
			expected: maps.Vertex{2, 2},
			score:    5,
		}, {
			tag:      "walk on wall",
			path:     "duull",
			expected: maps.InvalidVertex,
			score:    0,
		}, {
			tag:      "walk off world",
			path:     "ddddd",
			expected: maps.InvalidVertex,
			score:    0,
		},
	} {
		t.Run(test.tag, func(t *testing.T) {
			path := maps.ParsePath(test.path)
			v := path.EndingVertex(m)
			if v != test.expected {
				t.Errorf("path %s should have ended on %s; ended on %s", test.path, test.expected, v)
			}
			s := path.Score(m)
			if s != test.score {
				t.Errorf("path %s should have scored %d points; scored %d points", path, test.score, s)
			}
		})
	}
}

func TestPadding(t *testing.T) {
	s := `=3,5,5
		  .w.w.
		  w.s.w
		  .w.w.`
	r := maps.NewReader(strings.NewReader(s))
	m, err := r.Next()
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		tag  string
		path string
		add  string
	}{
		{
			tag:  "down",
			path: "u",
			add:  "dudu",
		}, {
			tag:  "up",
			path: "d",
			add:  "udud",
		}, {
			tag:  "left",
			path: "r",
			add:  "lrlr",
		}, {
			tag:  "right",
			path: "l",
			add:  "rlrl",
		},
	} {
		t.Run(test.tag, func(t *testing.T) {
			p := maps.ParsePath(test.path)
			p.Pad(m)
			if p.String() != test.path+test.add {
				t.Errorf("Expected to pad %s with %s; got %s", test.path, test.add, p)
			}
		})
	}
}
