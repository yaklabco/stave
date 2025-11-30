package parse

import (
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"testing"
)

func TestToOneLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "multiline with newlines",
			s:    "line1\nline2\nline3",
			want: "line1 line2 line3",
		},
		{
			name: "leading and trailing spaces",
			s:    "  text  ",
			want: "text",
		},
		{
			name: "empty string",
			s:    "",
			want: "",
		},
		{
			name: "single line",
			s:    "single line",
			want: "single line",
		},
		{
			name: "multiple newlines",
			s:    "line1\n\n\nline2",
			want: "line1   line2",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := toOneLine(tt.s)
			if got != tt.want {
				t.Errorf("toOneLine(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestHasContextParam(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		code    string
		wantCtx bool
		wantErr bool
	}{
		{
			name:    "no params",
			code:    "func Foo() {}",
			wantCtx: false,
			wantErr: false,
		},
		{
			name:    "context first",
			code:    "func Foo(ctx context.Context) {}",
			wantCtx: true,
			wantErr: false,
		},
		{
			name:    "context not first",
			code:    "func Foo(s string, ctx context.Context) {}",
			wantCtx: false,
			wantErr: false,
		},
		{
			name:    "multiple contexts error",
			code:    "func Foo(ctx1, ctx2 context.Context) {}",
			wantCtx: false,
			wantErr: true,
		},
		{
			name:    "non-context param",
			code:    "func Foo(s string) {}",
			wantCtx: false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", "package p\nimport \"context\"\n"+tt.code, 0)
			if err != nil {
				t.Fatal(err)
			}
			fn := f.Decls[1].(*ast.FuncDecl)
			gotCtx, err := hasContextParam(fn.Type)
			if (err != nil) != tt.wantErr {
				t.Errorf("hasContextParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCtx != tt.wantCtx {
				t.Errorf("hasContextParam() = %v, want %v", gotCtx, tt.wantCtx)
			}
		})
	}
}

func TestHasErrorReturn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		code      string
		wantError bool
		wantErr   bool
	}{
		{
			name:      "void return",
			code:      "func Foo() {}",
			wantError: false,
			wantErr:   false,
		},
		{
			name:      "single error return",
			code:      "func Foo() error { return nil }",
			wantError: true,
			wantErr:   false,
		},
		{
			name:      "multiple returns error",
			code:      "func Foo() (int, error) { return 0, nil }",
			wantError: false,
			wantErr:   true,
		},
		{
			name:      "non-error return error",
			code:      "func Foo() int { return 0 }",
			wantError: false,
			wantErr:   true,
		},
		{
			name:      "multiple error returns",
			code:      "func Foo() (err1, err2 error) { return nil, nil }",
			wantError: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", "package p\n"+tt.code, 0)
			if err != nil {
				t.Fatal(err)
			}
			fn := f.Decls[0].(*ast.FuncDecl)
			gotError, err := hasErrorReturn(fn.Type)
			if (err != nil) != tt.wantErr {
				t.Errorf("hasErrorReturn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotError != tt.wantError {
				t.Errorf("hasErrorReturn() = %v, want %v", gotError, tt.wantError)
			}
		})
	}
}

func TestHasVoidReturn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code string
		want bool
	}{
		{
			name: "void return true",
			code: "func Foo() {}",
			want: true,
		},
		{
			name: "single return false",
			code: "func Foo() error { return nil }",
			want: false,
		},
		{
			name: "multiple returns false",
			code: "func Foo() (int, error) { return 0, nil }",
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", "package p\n"+tt.code, 0)
			if err != nil {
				t.Fatal(err)
			}
			fn := f.Decls[0].(*ast.FuncDecl)
			got := hasVoidReturn(fn.Type)
			if got != tt.want {
				t.Errorf("hasVoidReturn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLit2String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		value  string
		want   string
		wantOk bool
	}{
		{
			name:   "quoted string",
			value:  `"foo"`,
			want:   "foo",
			wantOk: true,
		},
		{
			name:   "no quotes fail",
			value:  "foo",
			want:   "",
			wantOk: false,
		},
		{
			name:   "single quotes fail",
			value:  "'foo'",
			want:   "",
			wantOk: false,
		},
		{
			name:   "empty quoted string",
			value:  `""`,
			want:   "",
			wantOk: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lit := &ast.BasicLit{Value: tt.value}
			got, gotOk := lit2string(lit)
			if gotOk != tt.wantOk {
				t.Errorf("lit2string() ok = %v, wantOk %v", gotOk, tt.wantOk)
			}
			if gotOk && got != tt.want {
				t.Errorf("lit2string() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsNamespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code string
		want bool
	}{
		{
			name: "mg.Namespace true",
			code: "package p\nimport \"github.com/yaklabco/stave/mg\"\ntype Foo mg.Namespace",
			want: true,
		},
		{
			name: "other type false",
			code: "package p\ntype Foo struct{}",
			want: false,
		},
		{
			name: "multiple specs false",
			code: "package p\ntype (Foo int\nBar string)",
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", tt.code, parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}

			pkg := &ast.Package{
				Name:  "p",
				Files: map[string]*ast.File{"test.go": f},
			}
			docPkg := doc.New(pkg, "./", 0)

			if len(docPkg.Types) == 0 {
				if tt.want {
					t.Fatal("expected type, but none found")
				}
				return
			}

			got := isNamespace(docPkg.Types[0])
			if got != tt.want {
				t.Errorf("isNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckDupeTargets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		funcs     []*Function
		wantDupes bool
	}{
		{
			name: "no dupes",
			funcs: []*Function{
				{Name: "Foo"},
				{Name: "Bar"},
			},
			wantDupes: false,
		},
		{
			name: "case-insensitive dupe",
			funcs: []*Function{
				{Name: "Foo"},
				{Name: "foo"},
			},
			wantDupes: true,
		},
		{
			name: "namespace dupe",
			funcs: []*Function{
				{Name: "Foo", Receiver: "Build"},
				{Name: "foo", Receiver: "build"},
			},
			wantDupes: true,
		},
		{
			name: "different namespace no dupe",
			funcs: []*Function{
				{Name: "Foo", Receiver: "Build"},
				{Name: "Foo", Receiver: "Test"},
			},
			wantDupes: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			info := &PkgInfo{Funcs: tt.funcs}
			gotDupes, _ := checkDupeTargets(info)
			if gotDupes != tt.wantDupes {
				t.Errorf("checkDupeTargets() = %v, want %v", gotDupes, tt.wantDupes)
			}
		})
	}
}

func TestSanitizeSynopsis(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		doc  string
		fn   string
		want string
	}{
		{
			name: "name prefix removed",
			doc:  "Clean removes temporary files",
			fn:   "Clean",
			want: "removes temporary files",
		},
		{
			name: "no prefix kept",
			doc:  "This function cleans",
			fn:   "Clean",
			want: "This function cleans",
		},
		{
			name: "empty doc",
			doc:  "",
			fn:   "Foo",
			want: "",
		},
		{
			name: "case insensitive match",
			doc:  "clean does cleaning",
			fn:   "Clean",
			want: "does cleaning",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := &doc.Func{
				Name: tt.fn,
				Doc:  tt.doc,
			}
			got := sanitizeSynopsis(f)
			if got != tt.want {
				t.Errorf("sanitizeSynopsis() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetImportPathFromCommentGroup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want []string
	}{
		{
			name: "nil comments",
			text: "",
			want: nil,
		},
		{
			name: "valid stave:import",
			text: "//stave:import",
			want: []string{"stave:import"},
		},
		{
			name: "with alias",
			text: "//stave:import alias",
			want: []string{"stave:import", "alias"},
		},
		{
			name: "not import tag",
			text: "//other comment",
			want: nil,
		},
		{
			name: "spaces normalized",
			text: "// stave:import  alias",
			want: []string{"stave:import", "alias"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var cg *ast.CommentGroup
			if tt.text != "" {
				cg = &ast.CommentGroup{
					List: []*ast.Comment{{Text: tt.text}},
				}
			}

			got := getImportPathFromCommentGroup(cg)
			if len(got) != len(tt.want) {
				t.Errorf("getImportPathFromCommentGroup() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("getImportPathFromCommentGroup()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

