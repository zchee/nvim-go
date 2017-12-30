// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dwarf_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/debug/dwarf"
)

var (
	pcspTempDir    string
	pcsptestBinary string
)

func doPCToSPTest(self bool) bool {
	// For now, only works on amd64 platforms.
	if runtime.GOARCH != "amd64" {
		return false
	}
	// Self test reads test binary; only works on Linux or Mac.
	if self {
		if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
			return false
		}
	}
	// Command below expects "sh", so Unix.
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		return false
	}
	if pcsptestBinary != "" {
		return true
	}
	var err error
	pcspTempDir, err = ioutil.TempDir("", "pcsptest")
	if err != nil {
		panic(err)
	}
	if strings.Contains(pcspTempDir, " ") {
		panic("unexpected space in tempdir")
	}
	// This command builds pcsptest from testdata/pcsptest.go.
	pcsptestBinary = filepath.Join(pcspTempDir, "pcsptest")
	command := fmt.Sprintf("go tool compile -o %s.6 testdata/pcsptest.go && go tool link -H %s -o %s %s.6",
		pcsptestBinary, runtime.GOOS, pcsptestBinary, pcsptestBinary)
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	return true
}

func endPCToSPTest() {
	if pcspTempDir != "" {
		os.RemoveAll(pcspTempDir)
		pcspTempDir = ""
		pcsptestBinary = ""
	}
}

func TestPCToSPOffset(t *testing.T) {
	if !doPCToSPTest(false) {
		return
	}
	defer endPCToSPTest()

	data, err := getData(pcsptestBinary)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := data.LookupFunction("main.test")
	if err != nil {
		t.Fatal("lookup startPC:", err)
	}
	startPC, ok := entry.Val(dwarf.AttrLowpc).(uint64)
	if !ok {
		t.Fatal(`DWARF data for function "main.test" has no low PC`)
	}
	endPC, ok := entry.Val(dwarf.AttrHighpc).(uint64)
	if !ok {
		t.Fatal(`DWARF data for function "main.test" has no high PC`)
	}

	const addrSize = 8 // TODO: Assumes amd64.
	const argSize = 8  // Defined by int64 arguments in test binary.

	// On 64-bit machines, the first offset must be one address size,
	// for the return PC.
	offset, err := data.PCToSPOffset(startPC)
	if err != nil {
		t.Fatal("startPC:", err)
	}
	if offset != addrSize {
		t.Fatalf("expected %d at start of function; got %d", addrSize, offset)
	}
	// On 64-bit machines, expect some 8s and some 32s. (See the
	// comments in testdata/pcsptest.go.
	// TODO: The test could be stronger, but not much unless we
	// disassemble the binary.
	count := make(map[int64]int)
	for pc := startPC; pc < endPC; pc++ {
		offset, err := data.PCToSPOffset(pc)
		if err != nil {
			t.Fatal("scanning function:", err)
		}
		count[offset]++
	}
	if len(count) != 2 {
		t.Errorf("expected 2 offset values, got %d; counts are: %v", len(count), count)
	}
	if count[addrSize] == 0 {
		t.Errorf("expected some values at offset %d; got %v", addrSize, count)
	}
	if count[addrSize+3*argSize] == 0 {
		t.Errorf("expected some values at offset %d; got %v", addrSize+3*argSize, count)
	}
}
