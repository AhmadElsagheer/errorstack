# errorstack

A lightweight package to get stacktrace of go errors. Designed to work with `errors` standard package as of go1.20+. It also works well with `zap.Error` field to log error stacktrace.

This is a fork of [pkg/errors](https://github.com/pkg/errors) which is archived now. The package has been simplified to provide only stacktrace functionality.

## Usage
```go
package main

import (
	"errors"
	"fmt"

	"github.com/ahmadelsagheer/errorstack"
	"go.uber.org/zap"
)

func f1() error {
	return f2()
}

func f2() error {
	return f3()
}

func f3() error {
	return errorstack.WithStack(errors.New("f3 error"))
}

func main() {

	err := f1()

	fmt.Printf("%v\n", err)
	// f3 error
	fmt.Printf("--\n")

	fmt.Printf("%+v\n", err)
	// main.f3
	//         /Users/ahmad/codespace/errorstackexample/main.go:20
	// main.f2
	//         /Users/ahmad/codespace/errorstackexample/main.go:16
	// main.f1
	//         /Users/ahmad/codespace/errorstackexample/main.go:12
	// main.main
	//         /Users/ahmad/codespace/errorstackexample/main.go:25
	// runtime.main
	//         /usr/local/Cellar/go/1.21.0/libexec/src/runtime/proc.go:267
	// runtime.goexit
	//         /usr/local/Cellar/go/1.21.0/libexec/src/runtime/asm_amd64.s:1650
	fmt.Printf("--\n")

	logger := zap.NewExample()
	logger.Error("error occurred", zap.Error(err))
	// {"level":"error","msg":"error occurred","error":"f3 error","errorVerbose":"f3 error\nmain.f3\n\t/Users/ahmad/codespace/errorstackexample/main.go:20\nmain.f2\n\t/Users/ahmad/codespace/errorstackexample/main.go:16\nmain.f1\n\t/Users/ahmad/codespace/errorstackexample/main.go:12\nmain.main\n\t/Users/ahmad/codespace/errorstackexample/main.go:25\nruntime.main\n\t/usr/local/Cellar/go/1.21.0/libexec/src/runtime/proc.go:267\nruntime.goexit\n\t/usr/local/Cellar/go/1.21.0/libexec/src/runtime/asm_amd64.s:1650"}
}
```
