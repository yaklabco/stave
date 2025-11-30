package parse_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/yaklabco/stave/parse"
)

func TestFunctionID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   parse.Function
		want string
	}{
		{
			name: "with ImportPath",
			fn: parse.Function{
				ImportPath: "github.com/user/repo",
				Name:       "Build",
			},
			want: "github.com/user/repo.Build",
		},
		{
			name: "without ImportPath",
			fn: parse.Function{
				Name: "Build",
			},
			want: "<current>.Build",
		},
		{
			name: "with Receiver",
			fn: parse.Function{
				Name:     "Deploy",
				Receiver: "CI",
			},
			want: "<current>.CI.Deploy",
		},
		{
			name: "with ImportPath and Receiver",
			fn: parse.Function{
				ImportPath: "github.com/user/repo",
				Name:       "Test",
				Receiver:   "QA",
			},
			want: "github.com/user/repo.QA.Test",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.fn.ID()
			if got != tt.want {
				t.Errorf("Function.ID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFunctionTargetName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   parse.Function
		want string
	}{
		{
			name: "no alias or receiver",
			fn: parse.Function{
				Name: "Build",
			},
			want: "Build",
		},
		{
			name: "with receiver",
			fn: parse.Function{
				Name:     "Deploy",
				Receiver: "CI",
			},
			want: "CI:Deploy",
		},
		{
			name: "with alias",
			fn: parse.Function{
				Name:     "Build",
				PkgAlias: "pkg",
			},
			want: "pkg:Build",
		},
		{
			name: "with both alias and receiver",
			fn: parse.Function{
				Name:     "Test",
				PkgAlias: "pkg",
				Receiver: "QA",
			},
			want: "pkg:QA:Test",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.fn.TargetName()
			if got != tt.want {
				t.Errorf("Function.TargetName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFunctionExecCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		fn           parse.Function
		wantContains []string
	}{
		{
			name: "string arg",
			fn: parse.Function{
				Name: "Say",
				Args: []parse.Arg{{Name: "msg", Type: "string"}},
			},
			wantContains: []string{"arg0 := args.Args[x]"},
		},
		{
			name: "int arg",
			fn: parse.Function{
				Name: "Count",
				Args: []parse.Arg{{Name: "n", Type: "int"}},
			},
			wantContains: []string{"arg0, err := strconv.Atoi(args.Args[x])"},
		},
		{
			name: "float64 arg",
			fn: parse.Function{
				Name: "Calculate",
				Args: []parse.Arg{{Name: "val", Type: "float64"}},
			},
			wantContains: []string{"arg0, err := strconv.ParseFloat(args.Args[x], 64)"},
		},
		{
			name: "bool arg",
			fn: parse.Function{
				Name: "Toggle",
				Args: []parse.Arg{{Name: "flag", Type: "bool"}},
			},
			wantContains: []string{"arg0, err := strconv.ParseBool(args.Args[x])"},
		},
		{
			name: "time.Duration arg",
			fn: parse.Function{
				Name: "Wait",
				Args: []parse.Arg{{Name: "d", Type: "time.Duration"}},
			},
			wantContains: []string{"arg0, err := time.ParseDuration(args.Args[x])"},
		},
		{
			name: "with error return",
			fn: parse.Function{
				Name:    "Build",
				IsError: true,
			},
			wantContains: []string{"return Build("},
		},
		{
			name: "with context",
			fn: parse.Function{
				Name:      "Deploy",
				IsContext: true,
				IsError:   true,
			},
			wantContains: []string{"return Deploy(ctx)"},
		},
		{
			name: "with receiver",
			fn: parse.Function{
				Name:     "Test",
				Receiver: "CI",
			},
			wantContains: []string{"CI{}.Test("},
		},
		{
			name: "with package",
			fn: parse.Function{
				Name:    "Build",
				Package: "pkg",
			},
			wantContains: []string{"pkg.Build("},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.fn.ExecCode()
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("Function.ExecCode() does not contain %q\nGot:\n%s", want, got)
				}
			}
		})
	}
}

func TestFunctionsSort(t *testing.T) {
	t.Parallel()

	funcs := parse.Functions{
		&parse.Function{Name: "Zebra"},
		&parse.Function{Name: "Alpha"},
		&parse.Function{Name: "Beta"},
	}

	sort.Sort(funcs)

	expected := []string{"Alpha", "Beta", "Zebra"}
	for i, fn := range funcs {
		if fn.TargetName() != expected[i] {
			t.Errorf("Functions[%d] = %q, want %q", i, fn.TargetName(), expected[i])
		}
	}
}

func TestFunctionsSortInterface(t *testing.T) {
	t.Parallel()

	funcs := parse.Functions{
		&parse.Function{Name: "Foo"},
		&parse.Function{Name: "Bar"},
	}

	// Test Len
	if got := funcs.Len(); got != 2 {
		t.Errorf("Functions.Len() = %d, want 2", got)
	}

	// Test Less
	if !funcs.Less(1, 0) { // Bar < Foo
		t.Error("Functions.Less(1, 0) should be true (Bar < Foo)")
	}

	// Test Swap
	funcs.Swap(0, 1)
	if funcs[0].Name != "Bar" || funcs[1].Name != "Foo" {
		t.Errorf("Functions.Swap failed: got [%s, %s], want [Bar, Foo]", funcs[0].Name, funcs[1].Name)
	}
}

func TestImportsSort(t *testing.T) {
	t.Parallel()

	imports := parse.Imports{
		&parse.Import{UniqueName: "zebra_staveimport"},
		&parse.Import{UniqueName: "alpha_staveimport"},
		&parse.Import{UniqueName: "beta_staveimport"},
	}

	sort.Sort(imports)

	expected := []string{"alpha_staveimport", "beta_staveimport", "zebra_staveimport"}
	for i, imp := range imports {
		if imp.UniqueName != expected[i] {
			t.Errorf("Imports[%d].UniqueName = %q, want %q", i, imp.UniqueName, expected[i])
		}
	}
}

func TestImportsSortInterface(t *testing.T) {
	t.Parallel()

	imports := parse.Imports{
		&parse.Import{UniqueName: "foo_staveimport"},
		&parse.Import{UniqueName: "bar_staveimport"},
	}

	// Test Len
	if got := imports.Len(); got != 2 {
		t.Errorf("Imports.Len() = %d, want 2", got)
	}

	// Test Less
	if !imports.Less(1, 0) { // bar < foo
		t.Error("Imports.Less(1, 0) should be true (bar < foo)")
	}

	// Test Swap
	imports.Swap(0, 1)
	if imports[0].UniqueName != "bar_staveimport" || imports[1].UniqueName != "foo_staveimport" {
		t.Errorf("Imports.Swap failed: got [%s, %s], want [bar_staveimport, foo_staveimport]",
			imports[0].UniqueName, imports[1].UniqueName)
	}
}

