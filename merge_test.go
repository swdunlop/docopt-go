// Licensed under terms of MIT license (see LICENSE-MIT)
// Copyright (c) 2013 Keith Batten, kbatten@gmail.com

package docopt

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

func TestMerge(t *testing.T) {
	var o testOpts
	m := map[string]interface{}{
		"INT": "-3",
		"OCT": "0755",
		"-o":  "0755",
		"-v":  true,
		"I":   "1",
		"F":   "2e2",
		"AI":  []string{"1", "2", "3"},
	}
	err := Merge(&o, m)
	if err != nil {
		t.Errorf("merge-error: %v", err)
	}
	fmt.Fprintf(os.Stderr, "merge produced: %v\n", &o)
	assert(t, o.INT == -3)
	assert(t, o.OCT == 0755)
	assert(t, o.O == 0755)
	assert(t, o.V)
	assert(t, o.I == 1)
	assert(t, o.F == 2e2)
	assert(t, len(o.AI) == 3)
}

func assert(t *testing.T, q bool) {
	if q {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		panic("assertion failed, runtime cannot determine caller")
	}
	file = filepath.Base(file)
	fmt.Fprintf(os.Stderr, "%v:%v assertion failed\n", file, line)
	t.Fail()
}

type testOpts struct {
	INT int
	OCT testOctal
	O   testOctal `docopt:"-o"`
	V   bool      `docopt:"-v"`
	Q   bool      `docopt:"-"`
	I   int
	F   float64
	AI  []int
}

type testOctal uint64

func (oct *testOctal) MergeDocopt(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("expected octal, got %v", v)
	}
	o, err := strconv.ParseUint(str, 8, 64)
	if err == nil {
		*oct = testOctal(o)
	}
	return err
}
