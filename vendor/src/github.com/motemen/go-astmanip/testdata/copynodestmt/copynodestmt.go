package eg

import (
	"bytes"
	"io"
)

// DeclStmt
var v1 bytes.Buffer

var (
	v2, v3 int
	v4     bool
)

const version = "1"

const (
	kind0 = iota
	kind1
)

const (
	str0 = "0"
	str1 = "1"
)

func Fn1()
func Fn2() bool
func Fn3() (io.Reader, error)
func Fn4([]string) (io.Reader, error)
func Fn4(s []string) (r io.Reader, err error)

func Fn5(n int, ch chan bool) (io.Reader, error) {
	// DeferStmt
	defer func() {
		// ExprStmt
		fmt.Println("deferred")
	}()

	// BlockStmt
	{
		// EmptyStmt
	}

	a := []int{}
	// ForStmt
	for i := 0; i < n; i++ {
		// AssignStmt
		a = append(a, i)
	}

	// LabeledStmt
RangedFor:
	for i, x := range a {
		// GoStmt
		go func(i, x int) error {
			fmt.Println(x)
			// ReturnStmt
			return nil
		}(i, x)

		// IfStmt
		if i > 3 {
			continue
		}
	}

	// SelectStmt
	select {
	case x := <-ch:
		fmt.Println(x)
	default:
	}

	// SendStmt
	ch <- true

	// SwitchStmt
	switch n {
	case 1, 2, 3:
	case 4, 5:
	default:
	}

	var i interface{}

	// TypeSwitchStmt
	switch i.(type) {
	case map[string]string:
	case []bool:
	}
}
