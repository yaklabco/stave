package parse

import (
	"fmt"
	"go/ast"
	"go/doc"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/stave/pkg/st"
)

func init() {
	if err := os.Setenv(st.DebugEnv, "1"); err != nil {
		panic(fmt.Errorf("failed to set debug env var %q: %w", st.DebugEnv, err))
	}
}

func TestParse(t *testing.T) {
	ctx := t.Context()

	info, err := PrimaryPackage(ctx, "go", "./testdata", []string{"func.go", "command.go", "alias.go", "repeating_synopsis.go", "subcommands.go", "watch.go"}, false)
	if err != nil {
		t.Fatal(err)
	}

	expected := []Function{
		{
			Name:    "WatchTarget",
			IsWatch: true,
		},
		{
			Name:    "NonWatchTarget",
			IsWatch: false,
		},
		{
			Name:    "WatchDepsTarget",
			IsWatch: true,
		},
		{
			Name:     "ReturnsNilError",
			IsError:  true,
			Comment:  "Synopsis for \"returns\" error. And some more text.",
			Synopsis: `Synopsis for "returns" error.`,
		},
		{
			Name: "ReturnsVoid",
		},
		{
			Name:      "TakesContextReturnsError",
			IsError:   true,
			IsContext: true,
		},
		{
			Name:      "TakesContextReturnsVoid",
			IsError:   false,
			IsContext: true,
		},
		{
			Name:     "RepeatingSynopsis",
			IsError:  true,
			Comment:  "RepeatingSynopsis chops off the repeating function name. Some more text.",
			Synopsis: "chops off the repeating function name.",
		},
		{
			Name:     "Foobar",
			Receiver: "Build",
			IsError:  true,
		},
		{
			Name:     "Baz",
			Receiver: "Build",
			IsError:  false,
		},
	}

	if info.DefaultFunc == nil {
		t.Fatal("expected default func to exist, but was nil")
	}

	// DefaultIsError
	if !info.DefaultFunc.IsError {
		t.Fatalf("expected DefaultIsError to be true")
	}

	// DefaultName
	if info.DefaultFunc.Name != "ReturnsNilError" {
		t.Fatalf("expected DefaultName to be ReturnsNilError")
	}

	if info.Aliases["void"].Name != "ReturnsVoid" {
		t.Fatalf("expected alias of void to be ReturnsVoid")
	}

	f, ok := info.Aliases["baz"]
	if !ok {
		t.Fatal("missing alias baz")
	}
	if f.Name != "Baz" || f.Receiver != "Build" {
		t.Fatalf("expected alias of void to be Build.Baz")
	}

	if len(info.Aliases) != 2 {
		t.Fatalf("expected to only have two aliases, but have %#v", info.Aliases)
	}

	for _, expectedFunc := range expected {
		found := false
		for _, infoFn := range info.Funcs {
			if reflect.DeepEqual(expectedFunc, *infoFn) {
				found = true
				break
			}
			t.Logf("%#v", infoFn)
		}
		if !found {
			t.Fatalf("expected:\n%#v\n\nto be in:\n%#v", expectedFunc, info.Funcs)
		}
	}
}

func TestGetImportPathFromCommentGroupNil(t *testing.T) {
	// nil comments should return nil
	result := getImportPathFromCommentGroup(nil)
	if result != nil {
		t.Fatalf("expected nil for nil comments, got %v", result)
	}
}

func TestGetImportPathFromCommentGroupEmpty(t *testing.T) {
	// empty comment list should return nil
	cg := &ast.CommentGroup{List: []*ast.Comment{}}
	result := getImportPathFromCommentGroup(cg)
	if result != nil {
		t.Fatalf("expected nil for empty comments, got %v", result)
	}
}

func TestGetImportPathFromCommentGroupNoImportTag(t *testing.T) {
	cg := &ast.CommentGroup{List: []*ast.Comment{
		{Text: "// just a regular comment"},
	}}
	result := getImportPathFromCommentGroup(cg)
	if result != nil {
		t.Fatalf("expected nil for non-import comment, got %v", result)
	}
}

func TestGetImportPathFromCommentGroupWithImportTag(t *testing.T) {
	cg := &ast.CommentGroup{List: []*ast.Comment{
		{Text: "// stave:import"},
	}}
	result := getImportPathFromCommentGroup(cg)
	if result == nil {
		t.Fatal("expected non-nil result for stave:import comment")
	}
	if len(result) != 1 || result[0] != "stave:import" {
		t.Fatalf("expected [stave:import], got %v", result)
	}
}

func TestGetImportPathFromCommentGroupWithAlias(t *testing.T) {
	cg := &ast.CommentGroup{List: []*ast.Comment{
		{Text: "// stave:import foo"},
	}}
	result := getImportPathFromCommentGroup(cg)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result) != 2 || result[0] != "stave:import" || result[1] != "foo" {
		t.Fatalf("expected [stave:import foo], got %v", result)
	}
}

func TestGetImportPathFromCommentGroupMultipleComments(t *testing.T) {
	// import tag should be read from the last comment
	cg := &ast.CommentGroup{List: []*ast.Comment{
		{Text: "// a preceding comment"},
		{Text: "// another comment"},
		{Text: "// stave:import myalias"},
	}}
	result := getImportPathFromCommentGroup(cg)
	if result == nil {
		t.Fatal("expected non-nil result for stave:import in last comment")
	}
	if len(result) != 2 || result[0] != "stave:import" || result[1] != "myalias" {
		t.Fatalf("expected [stave:import myalias], got %v", result)
	}
}

func TestCheckDupeTargetsNoDupes(t *testing.T) {
	info := &PkgInfo{
		Funcs: Functions{
			{Name: "Build"},
			{Name: "Test"},
			{Name: "Clean"},
		},
	}
	hasDupes, _ := checkDupeTargets(info)
	if hasDupes {
		t.Fatal("expected no duplicates")
	}
}

func TestCheckDupeTargetsCaseInsensitive(t *testing.T) {
	info := &PkgInfo{
		Funcs: Functions{
			{Name: "Build"},
			{Name: "build"},
		},
	}
	hasDupes, names := checkDupeTargets(info)
	if !hasDupes {
		t.Fatal("expected duplicates for case-insensitive match")
	}
	if len(names["build"]) != 2 {
		t.Fatalf("expected 2 entries for 'build', got %d", len(names["build"]))
	}
}

func TestCheckDupeTargetsWithReceiver(t *testing.T) {
	info := &PkgInfo{
		Funcs: Functions{
			{Name: "Build", Receiver: "Deploy"},
			{Name: "Build", Receiver: "Test"},
		},
	}
	hasDupes, _ := checkDupeTargets(info)
	if hasDupes {
		t.Fatal("expected no duplicates - different receivers")
	}
}

func TestSanitizeSynopsis(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		doc      string
		want     string
	}{
		{
			name:     "removes function name prefix",
			funcName: "Clean",
			doc:      "Clean removes all generated files.",
			want:     "removes all generated files.",
		},
		{
			name:     "case insensitive prefix removal",
			funcName: "Build",
			doc:      "build compiles the project.",
			want:     "compiles the project.",
		},
		{
			name:     "no prefix to remove",
			funcName: "Build",
			doc:      "Compiles the project.",
			want:     "Compiles the project.",
		},
		{
			name:     "empty doc",
			funcName: "Build",
			doc:      "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &doc.Func{Name: tt.funcName, Doc: tt.doc}
			got := sanitizeSynopsis(f)
			if got != tt.want {
				t.Errorf("sanitizeSynopsis() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetImportSelf(t *testing.T) {
	ctx := t.Context()
	cwd, err := os.Getwd()
	require.NoError(t, err)
	imp, err := getImport(ctx, "go", cwd, "github.com/yaklabco/stave/internal/parse/testdata/importself", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if imp.Info.PkgName != "importself" {
		t.Fatalf("expected package importself, got %v", imp.Info.PkgName)
	}
}
