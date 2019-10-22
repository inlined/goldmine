package debug

import (
	"errors"
	"io"
	"os"
)

var (
	// Out is where debug data should be sent
	Out io.WriteCloser = uninitialized{}

	// Flag can be added with flag.Var(debug.Flag) to enable debugging control
	Flag = &flag{
		out: &Out,
	}
)

type uninitialized struct{}

func (uninitialized) Write(b []byte) (int, error) {
	return 0, errors.New("package debug must be initialized with flag.Var(debug.Flag)")
}

func (uninitialized) Close() error {
	return errors.New("package debug must be initialized with flag.Var(debug.Flag)")
}

type nop struct{}

func (nop) Write(b []byte) (int, error) {
	return len(b), nil
}
func (nop) Close() error {
	return nil
}

// Flag controls the location of the debug.Out writer
type flag struct {
	out *io.WriteCloser
	val string
}

func (f flag) String() string {
	return f.val
}

// Set implements flag.Value
func (f *flag) Set(s string) error {
	f.val = s
	var err error
	switch s {
	case "":
		*f.out = nop{}
	case "1":
		*f.out = os.Stdout
	case "2":
		*f.out = os.Stderr
	default:
		*f.out, err = os.Create(s)
	}
	return err
}
