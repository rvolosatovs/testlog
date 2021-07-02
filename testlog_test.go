package testlog_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/rvolosatovs/testlog"
	. "github.com/rvolosatovs/testlog"
)

var as = [...][]interface{}{
	{"something happened"},
	{"something more", "happened"},
}

func someFunc() {
	for _, a := range as {
		Print(a...)
	}
}

func TestTest(t *testing.T) {
	f, err := ioutil.TempFile("", fmt.Sprintf("testlog_test-%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatalf("Failed to create temporary file: %s", err)
	}
	path := f.Name()

	t.Cleanup(func() {
		if err := os.RemoveAll(path); err != nil {
			t.Logf("Failed to remove %q: %s", path, err)
		}
	})

	SetPath(path)
	t.Logf("Set path to %q", path)

	var i int
	SetFormatter(func(_ int, a ...interface{}) string {
		defer func() { i++ }()

		switch {
		case i > len(as):
			t.Errorf("Invocation %d must not have happened (only %d invocations expected)", i, len(as))
		case !reflect.DeepEqual(as[i], a):
			t.Errorf("Invocation %d must have contents %v, got %v", i, as[i], a)
		}
		return fmt.Sprintln(i)
	})
	someFunc()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatalf("Failed to read log file from %q: %s", path, err)
	}
	if s := string(b); s != func() (expected string) {
		for i := range as {
			expected = fmt.Sprintf("%s%d\n", expected, i)
		}
		return
	}() {
		t.Errorf("Expected to read a sequence of numbers from 0 to %d from %q, got: %s", len(as), path, s)
	}
}

func Example() {
	testlog.Print("Print")
	testlog.Printf("Printf %d", 42)

	doSomething := func(v int) (err error) {
		l := testlog.Start("doSomething", "<- ", v)
		defer func() { l("-> ", err) }() // NOTE: this must be done in a function to ensure returned err value is used.

		return errors.New("error happened")
	}
	doSomething(1)
	doSomething(2)

	testlog.Stop("example")
}
