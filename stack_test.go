package errorstack

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"testing"
)

var initpc = caller()

type X struct{}

// val returns a Frame pointing to itself.
func (x X) val() Frame {
	return caller()
}

// ptr returns a Frame pointing to itself.
func (x *X) ptr() Frame {
	return caller()
}

func TestFrameFormat(t *testing.T) {
	var tests = []struct {
		Frame
		format string
		want   string
	}{{
		initpc,
		"%s",
		"stack_test.go",
	}, {
		initpc,
		"%+s",
		"github.com/ahmadelsagheer/errorstack.init\n" +
			"\t.+/errorstack/stack_test.go",
	}, {
		0,
		"%s",
		"unknown",
	}, {
		0,
		"%+s",
		"unknown",
	}, {
		initpc,
		"%d",
		"12",
	}, {
		0,
		"%d",
		"0",
	}, {
		initpc,
		"%n",
		"init",
	}, {
		func() Frame {
			var x X
			return x.ptr()
		}(),
		"%n",
		`\(\*X\).ptr`,
	}, {
		func() Frame {
			var x X
			return x.val()
		}(),
		"%n",
		"X.val",
	}, {
		0,
		"%n",
		"",
	}, {
		initpc,
		"%v",
		"stack_test.go:12",
	}, {
		initpc,
		"%+v",
		"github.com/ahmadelsagheer/errorstack.init\n" +
			"\t.+/errorstack/stack_test.go:12",
	}, {
		0,
		"%v",
		"unknown:0",
	}}

	for i, tt := range tests {
		testFormatRegexp(t, i, tt.Frame, tt.format, tt.want)
	}
}

func TestFuncname(t *testing.T) {
	tests := []struct {
		name, want string
	}{
		{"", ""},
		{"runtime.main", "main"},
		{"github.com/ahmadelsagheer/errorstack.funcname", "funcname"},
		{"funcname", "funcname"},
		{"io.copyBuffer", "copyBuffer"},
		{"main.(*R).Write", "(*R).Write"},
	}

	for _, tt := range tests {
		got := funcname(tt.name)
		want := tt.want
		if got != want {
			t.Errorf("funcname(%q): want: %q, got %q", tt.name, want, got)
		}
	}
}

func TestStackTrace(t *testing.T) {
	testCases := []struct {
		err  error
		want []string
	}{
		{

			err: WithStack(errors.New("ooh")),
			want: []string{
				"github.com/ahmadelsagheer/errorstack.TestStackTrace\n" +
					"\t.+/errorstack/stack_test.go:126",
			},
		},
		{
			err: WithStack(fmt.Errorf("ahh: %w", errors.New("ooh"))),
			want: []string{
				"github.com/ahmadelsagheer/errorstack.TestStackTrace\n" +
					"\t.+/errorstack/stack_test.go:133", // this is the stack of fmt.Errorf
			},
		},
	}
	for i, tc := range testCases {
		st, ok := GetStack(tc.err)
		if !ok {
			t.Errorf("expected %#v to have a stack", tc.err)
			continue
		}
		for j, want := range tc.want {
			testFormatRegexp(t, i, st[j], "%+v", want)
		}
	}
}

func stackTrace() StackTrace {
	const depth = 8
	var pcs [depth]uintptr
	n := runtime.Callers(1, pcs[:])
	var st stack = pcs[0:n]
	return st.StackTrace()
}

func TestStackTraceFormat(t *testing.T) {
	tests := []struct {
		StackTrace
		format string
		want   string
	}{{
		nil,
		"%s",
		`\[\]`,
	}, {
		nil,
		"%v",
		`\[\]`,
	}, {
		nil,
		"%+v",
		"",
	}, {
		nil,
		"%#v",
		`\[\]errorstack.Frame\(nil\)`,
	}, {
		make(StackTrace, 0),
		"%s",
		`\[\]`,
	}, {
		make(StackTrace, 0),
		"%v",
		`\[\]`,
	}, {
		make(StackTrace, 0),
		"%+v",
		"",
	}, {
		make(StackTrace, 0),
		"%#v",
		`\[\]errorstack.Frame{}`,
	}, {
		stackTrace()[:2],
		"%s",
		`\[stack_test.go stack_test.go\]`,
	}, {
		stackTrace()[:2],
		"%v",
		`\[stack_test.go:155 stack_test.go:202\]`,
	}, {
		stackTrace()[:2],
		"%+v",
		"\n" +
			"github.com/ahmadelsagheer/errorstack.stackTrace\n" +
			"\t.+/errorstack/stack_test.go:155\n" +
			"github.com/ahmadelsagheer/errorstack.TestStackTraceFormat\n" +
			"\t.+/errorstack/stack_test.go:206",
	}, {
		stackTrace()[:2],
		"%#v",
		`\[\]errorstack.Frame{stack_test.go:155, stack_test.go:214}`,
	}}

	for i, tt := range tests {
		testFormatRegexp(t, i, tt.StackTrace, tt.format, tt.want)
	}
}

// a version of runtime.Caller that returns a Frame, not a uintptr.
func caller() Frame {
	var pcs [3]uintptr
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	frame, _ := frames.Next()
	return Frame(frame.PC)
}

func TestFrameMarshalText(t *testing.T) {
	var tests = []struct {
		Frame
		want string
	}{{
		initpc,
		`^github.com/ahmadelsagheer/errorstack\.init(\.ializers)? .+/errorstack/stack_test.go:\d+$`,
	}, {
		0,
		`^unknown$`,
	}}
	for i, tt := range tests {
		got, err := tt.Frame.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		if !regexp.MustCompile(tt.want).Match(got) {
			t.Errorf("test %d: MarshalJSON:\n got %q\n want %q", i+1, string(got), tt.want)
		}
	}
}

func TestFrameMarshalJSON(t *testing.T) {
	var tests = []struct {
		Frame
		want string
	}{{
		initpc,
		`^"github\.com/ahmadelsagheer/errorstack\.init(\.ializers)? .+/errorstack/stack_test.go:\d+"$`,
	}, {
		0,
		`^"unknown"$`,
	}}
	for i, tt := range tests {
		got, err := json.Marshal(tt.Frame)
		if err != nil {
			t.Fatal(err)
		}
		if !regexp.MustCompile(tt.want).Match(got) {
			t.Errorf("test %d: MarshalJSON:\n got %q\n want %q", i+1, string(got), tt.want)
		}
	}
}
