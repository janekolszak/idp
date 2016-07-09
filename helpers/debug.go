// +build debug

package helpers

import (
	"fmt"
	"time"
)

// Debug prints given paramters with the prefix "<date> <file>:<line>"
func Debug(args ...interface{}) {
	f := callers().StackTrace()[0]
	fmt.Printf("%s %s:%d ", time.Now().Format(time.RFC3339), f, f)
	fmt.Println(args...)
}

// Debugf formats and prints given paramters with the prefix "<date> <file>:<line>"
func Debugf(f string, args ...interface{}) {
	frm := callers().StackTrace()[0]
	fmt.Printf("%s %s:%d ", time.Now().Format(time.RFC3339), frm, frm)
	fmt.Printf(f, args...)
}

// Verbose prints given paramters and full stack trace
func Verbose(args ...interface{}) {
	fs := callers().StackTrace()
	fmt.Println(args...)
	fmt.Printf("%+v\n", fs)
}

// Verbose formats and prints given paramters and full stack trace
func Verbosef(f string, args ...interface{}) {
	fs := callers().StackTrace()
	fmt.Printf(f, args...)
	fmt.Printf("%+v\n", fs)
}
