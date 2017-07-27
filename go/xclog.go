// Go support for leveled logs//
//
//
package xclog

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type severity int32

const (
	NONE severity = iota
	FATAL
	CRIT
	ERROR
	WARN
	NOTICE
	INFO
	VERBOSE
	DEBUG
	numSeverity
)

const severityChar = " FCEWNIVD"

var severityName = []string{
	FATAL:   "FATAL",
	CRIT:    "CRITICAL",
	ERROR:   "ERROR",
	WARN:    "WARNING",
	NOTICE:  "NOTICE",
	INFO:    "INFO",
	VERBOSE: "VERBOSE",
	DEBUG:   "DEBUG",
}

func init() {
	logging.difflevel = WARN
	logging.errlevel = ERROR
	logging.outlevel = INFO
}

// Flush flushes all pending log I/O.
func Flush() {
	os.Stdout.Sync()
	os.Stderr.Sync()
	//logging.lockAndFlushAll()
}

// loggingT collects all the global state of the logging setup.
type loggingT struct {
	// mu protects the remaining elements of this structure and is
	// used to synchronize logging.
	mu sync.Mutex
	// file holds writer for each of the log types.

	difflevel severity
	errlevel  severity
	outlevel  severity
}

var logging loggingT

func (l *loggingT) setDiffLevel(s severity) {
	atomic.StoreInt32((*int32)(&l.difflevel), int32(s))
}
func (l *loggingT) setErrLevel(s severity) {
	atomic.StoreInt32((*int32)(&l.errlevel), int32(s))
}
func (l *loggingT) setOutLevel(s severity) {
	atomic.StoreInt32((*int32)(&l.outlevel), int32(s))
}

func (l *loggingT) getDiffLevel() severity {
	return severity(atomic.LoadInt32((*int32)(&l.difflevel)))
}
func (l *loggingT) getErrLevel() severity {
	return severity(atomic.LoadInt32((*int32)(&l.errlevel)))
}
func (l *loggingT) getOutLevel() severity {
	return severity(atomic.LoadInt32((*int32)(&l.outlevel)))
}

/*
header formats a log header as defined by the C++ implementation.
It returns a buffer containing the formatted header.

Log lines have this form:
	Lmmdd hh:mm:ss.uuuuuu threadid file:line] msg...
where the fields are defined as follows:
	L                A single character, representing the log level (eg 'I' for INFO)
	mm               The month (zero padded; ie May is '05')
	dd               The day (zero padded)
	hh:mm:ss.uuuuuu  Time in hours, minutes and fractional seconds
	threadid         The space-padded thread ID as returned by GetTID()
	file             The file name
	line             The line number
	msg              The user-supplied message
*/
func (l *loggingT) header(s severity) string {
	// Lmmdd hh:mm:ss.uuuuuu threadid file:line]
	now := time.Now()
	_, file, line, ok := runtime.Caller(3) // It's always the same number of frames to the user's call.
	if !ok {
		file = "???"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	if line < 0 {
		line = 0 // not a real line number, but acceptable to someDigits
	}
	// if s > fatalLog { s = infoLog }

	// Avoid Fprintf, for speed. The format is so simple that we can do it quickly by hand.
	// It's worth about 3X. Fprintf is hard.
	_, month, day := now.Date()
	hour, minute, second := now.Clock()

	header := fmt.Sprintf("%c%02d%02d %02d:%02d:%02d %s:%d] ", severityChar[s], int(month), day, hour, minute, second, file, line)
	return header
}

func (l *loggingT) println(s severity, args ...interface{}) {
	header := l.header(s)
	body := fmt.Sprintln(args...)
	l.output(s, header, body, false)
}

func (l *loggingT) print(s severity, args ...interface{}) {
	header := l.header(s)
	body := fmt.Sprint(args...)
	l.output(s, header, body, true)
}

func (l *loggingT) printf(s severity, format string, args ...interface{}) {
	header := l.header(s)
	body := fmt.Sprintf(format, args...)
	l.output(s, header, body, true)
}

// output writes the data to the log files and releases the buffer.
func (l *loggingT) output(s severity, header, body string, enter bool) {
	dl, el, ol := l.getDiffLevel(), l.getErrLevel(), l.getOutLevel()
	var out *os.File

	if s <= dl && s <= el {
		out = os.Stderr
	}
	if s > dl && s <= ol {
		out = os.Stdout
	}

	if out == nil {
		return
	}

	l.mu.Lock()
	fmt.Fprint(out, header, body)
	if enter {
		out.Write([]byte("\n"))
	}

	if s == FATAL {
		// Make sure we see the trace for the current goroutine on standard error.
		os.Stderr.Write(stacks(false))
		// Write the stack trace for all goroutines to the files.
		//trace := stacks(true)
		l.mu.Unlock()
		timeoutFlush(10 * time.Second)
		os.Exit(255) // C++ uses -1, which is silly because it's anded with 255 anyway.
	}
	l.mu.Unlock()
}

// timeoutFlush calls Flush and returns when it completes or after timeout
// elapses, whichever happens first.  This is needed because the hooks invoked
// by Flush may deadlock when glog.Fatal is called from a hook that holds
// a lock.
func timeoutFlush(timeout time.Duration) {
	done := make(chan bool, 1)
	go func() {
		Flush() // calls logging.lockAndFlushAll()
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		fmt.Fprintln(os.Stderr, "glog: Flush took longer than", timeout)
	}
}

// stacks is a wrapper for runtime.Stack that attempts to recover the data for all goroutines.
func stacks(all bool) []byte {
	// We don't know how big the traces are, so grow a few times if they don't fit. Start large, though.
	n := 10000
	if all {
		n = 100000
	}
	var trace []byte
	for i := 0; i < 5; i++ {
		trace = make([]byte, n)
		nbytes := runtime.Stack(trace, all)
		if nbytes < len(trace) {
			return trace[:nbytes]
		}
		n *= 2
	}
	return trace
}

// outer Interface

func SetDiffLevel(s severity) {
	logging.setDiffLevel(s)
}
func SetErrLevel(s severity) {
	logging.setErrLevel(s)
}
func SetOutLevel(s severity) {
	logging.setOutLevel(s)
}

func Fatal(args ...interface{}) {
	logging.print(FATAL, args...)
}
func Fatalf(format string, args ...interface{}) {
	logging.printf(FATAL, format, args...)
}
func Fatalln(args ...interface{}) {
	logging.println(FATAL, args...)
}

func Crit(args ...interface{}) {
	logging.print(CRIT, args...)
}
func Critf(format string, args ...interface{}) {
	logging.printf(CRIT, format, args...)
}
func Critln(args ...interface{}) {
	logging.println(CRIT, args...)
}

func Error(args ...interface{}) {
	logging.print(ERROR, args...)
}
func Errorf(format string, args ...interface{}) {
	logging.printf(ERROR, format, args...)
}
func Errorln(args ...interface{}) {
	logging.println(ERROR, args...)
}

func Warn(args ...interface{}) {
	logging.print(WARN, args...)
}
func Warnf(format string, args ...interface{}) {
	logging.printf(WARN, format, args...)
}
func Warnln(args ...interface{}) {
	logging.println(WARN, args...)
}

func Notice(args ...interface{}) {
	logging.print(NOTICE, args...)
}
func Noticef(format string, args ...interface{}) {
	logging.printf(NOTICE, format, args...)
}
func Noticeln(args ...interface{}) {
	logging.println(NOTICE, args...)
}

func Info(args ...interface{}) {
	logging.print(INFO, args...)
}
func Infof(format string, args ...interface{}) {
	logging.printf(INFO, format, args...)
}
func Infoln(args ...interface{}) {
	logging.println(INFO, args...)
}

func Verbose(args ...interface{}) {
	logging.print(VERBOSE, args...)
}
func Verbosef(format string, args ...interface{}) {
	logging.printf(VERBOSE, format, args...)
}
func Verboseln(args ...interface{}) {
	logging.println(VERBOSE, args...)
}

func Debug(args ...interface{}) {
	logging.print(DEBUG, args...)
}
func Debugf(format string, args ...interface{}) {
	logging.printf(DEBUG, format, args...)
}
func Debugln(args ...interface{}) {
	logging.println(DEBUG, args...)
}
