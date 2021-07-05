# testlog

Print debugging, but a little bit nicer.

The use case this is primarily designed for is effectively debugging problematic, flaky tests.

# Usage

Use instead of `fmt.Println`, `fmt.Printf`, `log.Println`, `log.Printf` or `$YOUR_FAVORITE_PRINT_FUNCTION`.

On first write, `testlog` will create a log file if necessary.
If `TESTLOG_PATH` is specified, that path will be used for the log file, otherwise `testlog` will create a new file in your temporary file directory with `testlog-` prefix.

# Notes
If you invoke `testlog` from multiple packages, say `A` and `B`, and you test them simultaneously via e.g. `go test A B`, `testlog` may be executed by two (or more) test processes simultaneously (`go` testing framework may run tests for `A` and `B` as two different processes), in which case it would write to two different log files if `TESTLOG_PATH` is not set.
