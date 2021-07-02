package testlog

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

var (
	pathMu sync.Mutex
	path   string
)

func SetPath(elems ...string) {
	pathMu.Lock()
	defer pathMu.Unlock()

	if path != "" {
		panic("Path is already set")
	}
	path = filepath.Join(elems...)
}

type Formatter func(skip int, a ...interface{}) string

var (
	formatterMu sync.Mutex
	formatter   Formatter
)

func SetFormatter(f Formatter) {
	formatterMu.Lock()
	defer formatterMu.Unlock()

	if formatter != nil {
		panic("Formatter is already set")
	}
	formatter = f
}

var (
	once sync.Once

	pid int

	wg = func() *sync.WaitGroup {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		return wg
	}()
)

func DefaultFormatter(skip int, a ...interface{}) string {
	_, from, line, ok := runtime.Caller(skip + 1)
	if !ok {
		panic("Failed to identify caller")
	}
	return fmt.Sprintf(`[%s][%d] %s:%d	%s
`,
		time.Now().Format("15:04:05.000000000"),
		pid,
		from,
		line,
		fmt.Sprint(a...),
	)
}

func writeLine(skip int, a ...interface{}) {
	f := func() *os.File {
		openFile := func(path string) *os.File {
			f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				panic(fmt.Sprintf("Failed to open %q: %s", path, err))
			}
			return f
		}

		var f *os.File
		once.Do(func() {
			defer wg.Done()

			pid = os.Getpid()

			formatterMu.Lock()
			defer formatterMu.Unlock()
			if formatter == nil {
				formatter = DefaultFormatter
			}

			pathMu.Lock()
			defer pathMu.Unlock()

			f = func() *os.File {
				if path != "" {
					return openFile(path)
				}

				if s, ok := os.LookupEnv("TESTLOG_PATH"); ok {
					path = s
					return openFile(s)
				}

				f, err := ioutil.TempFile("", fmt.Sprintf("testlog-%d-*", time.Now().UnixNano()))
				if err != nil {
					panic(fmt.Sprintf("Failed to create temporary file: %s", err))
				}
				path = f.Name()
				return f
			}()
		})
		wg.Wait()
		if f != nil {
			return f
		}
		return openFile(path)
	}()

	name := f.Name()
	defer func() {
		if err := f.Close(); err != nil {
			panic(fmt.Sprintf("Failed to close %q: %s", name, err))
		}
	}()
	if _, err := fmt.Fprint(f, formatter(skip+1, a...)); err != nil {
		panic(fmt.Sprintf("Failed to write %q: %s", name, err))
	}
}

func Print(a ...interface{}) {
	writeLine(1, a...)
}

func Printf(format string, a ...interface{}) {
	writeLine(1, fmt.Sprintf(format, a...))
}

func stop(skip int, name string, a ...interface{}) {
	writeLine(skip+1, append([]interface{}{name, " stopped "}, a...)...)
}

func start(skip int, name string, a ...interface{}) func(...interface{}) {
	writeLine(skip+1, append([]interface{}{name, " started "}, a...)...)
	return func(a ...interface{}) {
		stop(1, name, a...)
	}
}

func Start(name string, a ...interface{}) func(...interface{}) {
	return start(1, name, a...)
}

func Stop(name string, a ...interface{}) {
	stop(1, name, a...)
}

func Test(t *testing.T, a ...interface{}) func(...interface{}) {
	return start(1, t.Name(), a...)
}
