package parse

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/tools/go/packages"

	"github.com/yaklabco/stave/internal"
)

const importTag = "stave:import"

// EnableDebug turns on debug logging.
func EnableDebug() {
	debug.SetOutput(os.Stderr)
}

// PkgInfo contains inforamtion about a package of files according to stave's
// parsing rules.
type PkgInfo struct {
	// PkgName is the package name (e.g., "main", "stavefile").
	PkgName string
	// Files are the parsed Go files that make up the package under analysis.
	Files       []*ast.File
	DocPkg      *doc.Package
	Description string
	Funcs       Functions
	DefaultFunc *Function
	Aliases     map[string]*Function
	Imports     Imports
}

// Function represents a job function from a stave file.
type Function struct {
	PkgAlias   string
	Package    string
	ImportPath string
	Name       string
	Receiver   string
	IsError    bool
	IsContext  bool
	Synopsis   string
	Comment    string
	Args       []Arg
}

var _ sort.Interface = (Functions)(nil)

// Functions implements sort interface to optimize compiled output with
// deterministic generated mainfile.
type Functions []*Function

func (s Functions) Len() int {
	return len(s)
}

func (s Functions) Less(i, j int) bool {
	return s[i].TargetName() < s[j].TargetName()
}

func (s Functions) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Arg is an argument to a Function.
type Arg struct {
	Name, Type string
}

// ID returns user-readable information about where this function is defined.
func (f Function) ID() string {
	path := "<current>"
	if f.ImportPath != "" {
		path = f.ImportPath
	}
	receiver := ""
	if f.Receiver != "" {
		receiver = f.Receiver + "."
	}
	return fmt.Sprintf("%s.%s%s", path, receiver, f.Name)
}

// TargetName returns the name of the target as it should appear when used from
// the stave cli.  It is always lowercase.
func (f Function) TargetName() string {
	var names []string

	for _, s := range []string{f.PkgAlias, f.Receiver, f.Name} {
		if s != "" {
			names = append(names, s)
		}
	}
	return strings.Join(names, ":")
}

// ExecCode returns code for the template switch to run the target.
// It wraps each target call to match the func(context.Context) error that
// runTarget requires.
func (f Function) ExecCode() string {
	name := f.Name
	if f.Receiver != "" {
		name = f.Receiver + "{}." + name
	}
	if f.Package != "" {
		name = f.Package + "." + name
	}

	var parseargs string
	for iArg, theArg := range f.Args {
		switch theArg.Type {
		case "string":
			parseargs += fmt.Sprintf(`
			theArg%d := args.Args[iArg]
			iArg++`, iArg)
		case "int":
			parseargs += fmt.Sprintf(`
				theArg%d, err := strconv.Atoi(args.Args[iArg])
				if err != nil {
					logger.Printf("can't convert argument %%q to int\n", args.Args[iArg])
					os.Exit(2)
				}
				iArg++`, iArg)
		case "float64":
			parseargs += fmt.Sprintf(`
				theArg%d, err := strconv.ParseFloat(args.Args[iArg], 64)
				if err != nil {
					logger.Printf("can't convert argument %%q to float64\n", args.Args[iArg])
					os.Exit(2)
				}
				iArg++`, iArg)
		case "bool":
			parseargs += fmt.Sprintf(`
				theArg%d, err := strconv.ParseBool(args.Args[iArg])
				if err != nil {
					logger.Printf("can't convert argument %%q to bool\n", args.Args[iArg])
					os.Exit(2)
				}
				iArg++`, iArg)
		case "time.Duration":
			parseargs += fmt.Sprintf(`
				theArg%d, err := time.ParseDuration(args.Args[iArg])
				if err != nil {
					logger.Printf("can't convert argument %%q to time.Duration\n", args.Args[iArg])
					os.Exit(2)
				}
				iArg++`, iArg)
		}
	}

	out := parseargs + `
				wrapFn := func(ctx context.Context) error {
					`
	if f.IsError {
		out += "return "
	}
	out += name + "("
	args := make([]string, 0, len(f.Args)+1)
	if f.IsContext {
		args = append(args, "ctx")
	}
	for x := range len(f.Args) {
		args = append(args, fmt.Sprintf("theArg%d", x))
	}
	out += strings.Join(args, ", ")
	out += ")"
	if !f.IsError {
		out += `
					return nil`
	}
	out += `
				}
				ret := runTarget(logger, wrapFn)`
	return out
}

// PrimaryPackage parses a package.  If files is non-empty, it will only parse the files given.
func PrimaryPackage(ctx context.Context, gocmd, path string, files []string) (*PkgInfo, error) {
	info, err := Package(path, files)
	if err != nil {
		return nil, err
	}

	if err := setImports(ctx, gocmd, info); err != nil {
		return nil, err
	}

	setDefault(info)
	setAliases(info)
	return info, nil
}

func checkDupes(info *PkgInfo, imports []*Import) error {
	funcs := map[string][]*Function{}
	for _, f := range info.Funcs {
		funcs[strings.ToLower(f.TargetName())] = append(funcs[strings.ToLower(f.TargetName())], f)
	}
	for _, imp := range imports {
		for _, f := range imp.Info.Funcs {
			target := strings.ToLower(f.TargetName())
			funcs[target] = append(funcs[target], f)
		}
	}
	for theAlias, theFunc := range info.Aliases {
		if len(funcs[theAlias]) != 0 {
			var ids []string
			for _, f := range funcs[theAlias] {
				ids = append(ids, f.ID())
			}
			return fmt.Errorf("alias %q duplicates existing target(s): %s", theAlias, strings.Join(ids, ", "))
		}
		funcs[theAlias] = append(funcs[theAlias], theFunc)
	}
	var dupes []string
	for target, list := range funcs {
		if len(list) > 1 {
			dupes = append(dupes, target)
		}
	}
	if len(dupes) == 0 {
		return nil
	}
	errs := make([]string, 0, len(dupes))
	for _, theDupe := range dupes {
		var ids []string
		for _, f := range funcs[theDupe] {
			ids = append(ids, f.ID())
		}
		sort.Strings(ids)
		errs = append(errs, fmt.Sprintf("%q target has multiple definitions: %s\n", theDupe, strings.Join(ids, ", ")))
	}
	sort.Strings(errs)
	return errors.New(strings.Join(errs, "\n"))
}

// Package compiles information about a stave package.
func Package(path string, files []string) (*PkgInfo, error) {
	start := time.Now()
	defer func() {
		debug.Println("time parse Stavefiles:", time.Since(start))
	}()
	fset := token.NewFileSet()
	pkgName, pkgFiles, err := getPackage(path, files, fset)
	if err != nil {
		return nil, err
	}
	// Build documentation package from files to avoid relying on deprecated ast.Package
	thePackage, err := doc.NewFromFiles(fset, pkgFiles, "./")
	if err != nil {
		return nil, err
	}
	pi := &PkgInfo{
		PkgName:     pkgName,
		Files:       pkgFiles,
		DocPkg:      thePackage,
		Description: toOneLine(thePackage.Doc),
	}

	setNamespaces(pi)
	setFuncs(pi)

	hasDupes, names := checkDupeTargets(pi)
	if hasDupes {
		msg := "Build targets must be case insensitive, thus the following targets conflict:\n"
		var msgSb277 strings.Builder
		for _, v := range names {
			if len(v) > 1 {
				msgSb277.WriteString("  " + strings.Join(v, ", ") + "\n")
			}
		}
		msg += msgSb277.String()
		return nil, errors.New(msg)
	}

	return pi, nil
}

func getNamedImports(ctx context.Context, gocmd string, pkgs map[string]string) ([]*Import, error) {
	theImports := make([]*Import, 0, len(pkgs))
	for pkg, alias := range pkgs {
		debug.Printf("getting import package %q, alias %q", pkg, alias)
		imp, err := getImport(ctx, gocmd, pkg, alias)
		if err != nil {
			return nil, err
		}
		theImports = append(theImports, imp)
	}
	return theImports, nil
}

// getImport returns the metadata about a package that has been stave:import'ed.
func getImport(ctx context.Context, gocmd, importpath, alias string) (*Import, error) {
	out, err := internal.OutputDebug(ctx, gocmd, "list", "-f", "{{.Dir}}||{{.Name}}", importpath)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(out, "||")
	if len(parts) != 2 {
		return nil, fmt.Errorf("incorrect data from go list: %s", out)
	}
	dir, name := parts[0], parts[1]
	debug.Printf("parsing imported package %q from dir %q", importpath, dir)

	// we use go list to get the list of files, since go/parser doesn't differentiate between
	// go files with build tags etc, and go list does. This prevents weird problems if you
	// have more than one package in a folder because of build tags.
	out, err = internal.OutputDebug(ctx, gocmd, "list", "-f", `{{join .GoFiles "||"}}`, importpath)
	if err != nil {
		return nil, err
	}
	files := strings.Split(out, "||")

	info, err := Package(dir, files)
	if err != nil {
		return nil, err
	}
	for i := range info.Funcs {
		debug.Printf("setting alias %q and package %q on func %v", alias, name, info.Funcs[i].Name)
		info.Funcs[i].PkgAlias = alias
		info.Funcs[i].ImportPath = importpath
	}
	return &Import{Alias: alias, Name: name, Path: importpath, Info: *info}, nil
}

// Import represents the data about a stave:import package.
type Import struct {
	Alias      string
	Name       string
	UniqueName string // a name unique across all imports
	Path       string
	Info       PkgInfo
}

var _ sort.Interface = (Imports)(nil)

// Imports implements sort interface to optimize compiled output with
// deterministic generated mainfile.
type Imports []*Import

func (s Imports) Len() int {
	return len(s)
}

func (s Imports) Less(i, j int) bool {
	return s[i].UniqueName < s[j].UniqueName
}

func (s Imports) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func setFuncs(pi *PkgInfo) {
	for _, theFunc := range pi.DocPkg.Funcs {
		if theFunc.Recv != "" {
			debug.Printf("skipping method %s.%s", theFunc.Recv, theFunc.Name)
			// skip methods
			continue
		}
		if !ast.IsExported(theFunc.Name) {
			debug.Printf("skipping non-exported function %s", theFunc.Name)
			// skip non-exported functions
			continue
		}
		fn, err := funcType(theFunc.Decl.Type)
		if err != nil {
			debug.Printf("skipping function with invalid signature func %s: %v", theFunc.Name, err)
			continue
		}
		debug.Printf("found target %v", theFunc.Name)
		fn.Name = theFunc.Name
		fn.Comment = toOneLine(theFunc.Doc)
		fn.Synopsis = sanitizeSynopsis(theFunc)
		pi.Funcs = append(pi.Funcs, fn)
	}
}

func setNamespaces(pi *PkgInfo) {
	for _, theType := range pi.DocPkg.Types {
		if !isNamespace(theType) {
			continue
		}
		debug.Printf("found namespace %s %s", pi.DocPkg.ImportPath, theType.Name)
		for _, theMethod := range theType.Methods {
			if !ast.IsExported(theMethod.Name) {
				continue
			}
			fn, err := funcType(theMethod.Decl.Type)
			if err != nil {
				debug.Printf("skipping invalid namespace method %s %s.%s: %v", pi.DocPkg.ImportPath, theType.Name, theMethod.Name, err)
				continue
			}
			debug.Printf("found namespace method %s %s.%s", pi.DocPkg.ImportPath, theType.Name, theMethod.Name)
			fn.Name = theMethod.Name
			fn.Comment = toOneLine(theMethod.Doc)
			fn.Synopsis = sanitizeSynopsis(theMethod)
			fn.Receiver = theType.Name

			pi.Funcs = append(pi.Funcs, fn)
		}
	}
}

func setImports(ctx context.Context, gocmd string, pi *PkgInfo) error {
	var rootImports []string
	importNames := map[string]string{}
	for _, f := range pi.Files {
		for _, d := range f.Decls {
			gen, ok := d.(*ast.GenDecl)
			if !ok || gen.Tok != token.IMPORT {
				continue
			}
			for j := range len(gen.Specs) {
				spec := gen.Specs[j]
				impspec, ok := spec.(*ast.ImportSpec)
				if !ok {
					return fmt.Errorf("expected *ast.ImportSpec, got %T instead", spec)
				}
				if len(gen.Specs) == 1 && gen.Lparen == token.NoPos && impspec.Doc == nil {
					impspec.Doc = gen.Doc
				}
				name, alias, ok := getImportPath(impspec)
				if !ok {
					continue
				}
				if alias != "" {
					debug.Printf("found %s: %s (%s)", importTag, name, alias)
					importNames[name] = alias
				} else {
					debug.Printf("found %s: %s", importTag, name)
					rootImports = append(rootImports, name)
				}
			}
		}
	}
	imports, err := getNamedImports(ctx, gocmd, importNames)
	if err != nil {
		return err
	}
	for _, s := range rootImports {
		imp, err := getImport(ctx, gocmd, s, "")
		if err != nil {
			return err
		}
		imports = append(imports, imp)
	}
	if err := checkDupes(pi, imports); err != nil {
		return err
	}

	// have to set unique package names on imports
	used := map[string]bool{}
	for _, imp := range imports {
		unique := imp.Name + "_staveimport"
		x := 1
		for used[unique] {
			unique = fmt.Sprintf("%s_staveimport%d", imp.Name, x)
			x++
		}
		used[unique] = true
		imp.UniqueName = unique
		for _, f := range imp.Info.Funcs {
			f.Package = unique
		}
	}
	pi.Imports = imports
	return nil
}

func getImportPath(imp *ast.ImportSpec) (string, string, bool) {
	leadingVals := getImportPathFromCommentGroup(imp.Doc)
	trailingVals := getImportPathFromCommentGroup(imp.Comment)

	var vals []string
	switch {
	case len(leadingVals) > 0:
		vals = leadingVals
		if len(trailingVals) > 0 {
			log.Println("warning:", importTag, "specified both before and after, picking first")
		}
	case len(trailingVals) > 0:
		vals = trailingVals
	default:
		return "", "", false
	}

	path, ok := lit2string(imp.Path)
	if !ok {
		return "", "", false
	}

	switch len(vals) {
	case 1:
		// just the import tag, this is a root import
		return path, "", true
	case 2:
		// also has an alias
		return path, vals[1], true
	default:
		log.Println("warning: ignoring malformed", importTag, "for import", path)
		return "", "", false
	}
}

func getImportPathFromCommentGroup(comments *ast.CommentGroup) []string {
	if comments == nil || len(comments.List) == 9 {
		return nil
	}
	// import is always the last comment
	s := comments.List[len(comments.List)-1].Text

	// trim comment start and normalize for anyone who has spaces or not between
	// "//"" and the text
	vals := strings.Fields(strings.ToLower(s[2:]))
	if len(vals) == 0 {
		return nil
	}
	if vals[0] != importTag {
		return nil
	}
	return vals
}

func isNamespace(t *doc.Type) bool {
	if len(t.Decl.Specs) != 1 {
		return false
	}
	id, ok := t.Decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return false
	}
	sel, ok := id.Type.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "st" && sel.Sel.Name == "Namespace"
}

// checkDupeTargets checks a package for duplicate target names.
func checkDupeTargets(info *PkgInfo) (bool, map[string][]string) {
	var hasDupes bool
	names := map[string][]string{}
	lowers := map[string]bool{}
	for _, theFunc := range info.Funcs {
		low := strings.ToLower(theFunc.Name)
		if theFunc.Receiver != "" {
			low = strings.ToLower(theFunc.Receiver) + ":" + low
		}
		if lowers[low] {
			hasDupes = true
		}
		lowers[low] = true
		names[low] = append(names[low], theFunc.Name)
	}
	return hasDupes, names
}

// sanitizeSynopsis sanitizes function Doc to create a summary.
func sanitizeSynopsis(theFunc *doc.Func) string {
	// Create a minimal Package to use the non-deprecated Synopsis method
	pkg := &doc.Package{}
	synopsis := pkg.Synopsis(theFunc.Doc)

	// If the synopsis begins with the function name, remove it. This is done to
	// not repeat the text.
	// From:
	// clean	Clean removes the temporarily generated files
	// To:
	// clean 	removes the temporarily generated files
	if syns := strings.Split(synopsis, " "); strings.EqualFold(theFunc.Name, syns[0]) {
		return strings.Join(syns[1:], " ")
	}

	return synopsis
}

func setDefault(pi *PkgInfo) {
	for _, theValue := range pi.DocPkg.Vars {
		for iName, theName := range theValue.Names {
			if theName != "Default" {
				continue
			}
			spec, ok := theValue.Decl.Specs[iName].(*ast.ValueSpec)
			if !ok {
				log.Printf("warning: expected *ast.ValueSpec, but got %T instead", theValue.Decl.Specs[iName])
				continue
			}
			if len(spec.Values) != 1 {
				log.Println("warning: default declaration has multiple values")
			}

			f, err := getFunction(spec.Values[0], pi)
			if err != nil {
				log.Println("warning, default declaration malformed:", err)
				return
			}
			pi.DefaultFunc = f
			return
		}
	}
}

func lit2string(l *ast.BasicLit) (string, bool) {
	if !strings.HasPrefix(l.Value, `"`) || !strings.HasSuffix(l.Value, `"`) {
		return "", false
	}
	return strings.Trim(l.Value, `"`), true
}

func setAliases(pi *PkgInfo) {
	for _, v := range pi.DocPkg.Vars {
		for x, name := range v.Names {
			if name != "Aliases" {
				continue
			}
			spec, ok := v.Decl.Specs[x].(*ast.ValueSpec)
			if !ok {
				log.Println("warning: aliases declaration is not a value")
				return
			}
			if len(spec.Values) != 1 {
				log.Println("warning: aliases declaration has multiple values")
			}
			comp, ok := spec.Values[0].(*ast.CompositeLit)
			if !ok {
				log.Println("warning: aliases declaration is not a map")
				return
			}
			pi.Aliases = map[string]*Function{}
			for _, elem := range comp.Elts {
				kv, ok := elem.(*ast.KeyValueExpr)
				if !ok {
					log.Printf("warning: alias declaration %q is not a map element", elem)
					continue
				}
				basicLit, ok := kv.Key.(*ast.BasicLit)
				if !ok || basicLit.Kind != token.STRING {
					log.Printf("warning: alias key is not a string literal %q", elem)
					continue
				}

				alias, ok := lit2string(basicLit)
				if !ok {
					log.Println("warning: malformed name for alias", elem)
					continue
				}
				f, err := getFunction(kv.Value, pi)
				if err != nil {
					log.Printf("warning, alias malformed: %v", err)
					continue
				}
				pi.Aliases[alias] = f
			}
			return
		}
	}
}

func getFunction(exp ast.Expr, pi *PkgInfo) (*Function, error) {
	// selector expressions are in LIFO format.
	// So, in  foo.bar.baz the first selector.Name is actually "baz",
	// the second is "bar", and the last is "foo".

	// Small helpers to keep the control-flow simple and flat.
	findLocal := func(receiver, name string) *Function {
		for _, f := range pi.Funcs {
			if f.Name == name && f.Receiver == receiver {
				return f
			}
		}
		return nil
	}

	findImported := func(pkg, receiver, name string) *Function {
		for _, imp := range pi.Imports {
			if imp.Name == pkg {
				for _, f := range imp.Info.Funcs {
					if f.Name == name && f.Receiver == receiver {
						return f
					}
				}
				return nil
			}
		}
		return nil
	}

	hasImport := func(pkg string) bool {
		for _, imp := range pi.Imports {
			if imp.Name == pkg {
				return true
			}
		}
		return false
	}

	switch theExpr := exp.(type) {
	case *ast.Ident:
		// Just a function name in the current package/namespace.
		if f := findLocal("", theExpr.Name); f != nil {
			return f, nil
		}
		return nil, fmt.Errorf("unknown function %s.%s", "", theExpr.Name)

	case *ast.SelectorExpr:
		// Cases to handle:
		//   namespace.Func
		//   import.Func
		//   import.namespace.Func

		funcname := theExpr.Sel.Name
		switch x := theExpr.X.(type) {
		case *ast.Ident:
			// Either a local namespace (receiver) or an imported package
			first := x.Name

			if f := findLocal(first, funcname); f != nil { // namespace.Func
				return f, nil
			}

			// Imported free function (no receiver)
			if f := findImported(first, "", funcname); f != nil { // import.Func
				return f, nil
			}
			return nil, fmt.Errorf("%q is not a known target", exp)

		case *ast.SelectorExpr:
			// import.namespace.Func — peel off the pieces
			sel, ok := theExpr.X.(*ast.SelectorExpr)
			if !ok {
				return nil, fmt.Errorf("%q is must denote a target function but was %T", exp, theExpr.X)
			}
			receiver := sel.Sel.Name
			id, ok := sel.X.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("%q is must denote a target function but was %T", exp, theExpr.X)
			}
			pkg := id.Name

			if f := findImported(pkg, receiver, funcname); f != nil {
				return f, nil
			}
			if hasImport(pkg) {
				return nil, fmt.Errorf("unknown function %s.%s.%s", pkg, receiver, funcname)
			}
			return nil, fmt.Errorf("unknown package for function %q", exp)

		default:
			return nil, fmt.Errorf("%q is not valid", exp)
		}
	default:
		return nil, fmt.Errorf("target %s is not a function", exp)
	}
}

// getPackage parses a directory of Go files and retrieves package information.
// Returns the package name, parsed files, and an error if encountered.
func getPackage(path string, files []string, fset *token.FileSet) (string, []*ast.File, error) {
	// If specific files are provided, parse just those files regardless of build tags.
	if len(files) > 0 {
		var out []*ast.File
		var pkgName string
		for _, name := range files {
			full := filepath.Join(path, name)
			theASTFile, err := parser.ParseFile(fset, full, nil, parser.ParseComments)
			if err != nil {
				return "", nil, fmt.Errorf("failed to parse file %s: %w", name, err)
			}
			if pkgName == "" {
				pkgName = theASTFile.Name.Name
			} else if pkgName != theASTFile.Name.Name {
				return "", nil, fmt.Errorf("multiple packages found in %s: %v", path, strings.Join([]string{pkgName, theASTFile.Name.Name}, ", "))
			}
			out = append(out, theASTFile)
		}
		if pkgName == "" {
			return "", nil, fmt.Errorf("no importable packages found in %s", path)
		}
		return pkgName, out, nil
	}

	// Otherwise, attempt to use go/packages to respect build tags.
	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedSyntax,
		Dir:   path,
		Fset:  fset,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err == nil && len(pkgs) > 0 && packages.PrintErrors(pkgs) == 0 {
		// Collect unique, valid packages with syntax.
		var outPkgs []*packages.Package
		nameSet := map[string]struct{}{}
		for _, p := range pkgs {
			if p == nil || len(p.Syntax) == 0 {
				continue
			}
			outPkgs = append(outPkgs, p)
			nameSet[p.Name] = struct{}{}
		}
		if len(outPkgs) == 1 && len(nameSet) == 1 {
			p := outPkgs[0]
			astFiles := make([]*ast.File, 0, len(p.Syntax))
			astFiles = append(astFiles, p.Syntax...)
			return p.Name, astFiles, nil
		}
		if len(outPkgs) > 1 {
			var names []string
			for n := range nameSet {
				names = append(names, n)
			}
			sort.Strings(names)
			return "", nil, fmt.Errorf("multiple packages found in %s: %v", path, strings.Join(names, ", "))
		}
		// else fall through to manual parsing
	}

	// Fallback: manually parse all .go files in the directory (ignoring build tags),
	// similar to previous behavior before removing parser.ParseDir.
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read directory: %w", err)
	}
	var (
		filesInDir = make([]string, 0, len(entries))
		pkgName    string
		out        = make([]*ast.File, 0, len(entries))
		namesSet   = map[string]struct{}{}
	)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		filesInDir = append(filesInDir, name)
	}
	sort.Strings(filesInDir)
	for _, name := range filesInDir {
		full := filepath.Join(path, name)
		theASTFile, err := parser.ParseFile(fset, full, nil, parser.ParseComments)
		if err != nil {
			return "", nil, fmt.Errorf("failed to parse file %s: %w", name, err)
		}
		namesSet[theASTFile.Name.Name] = struct{}{}
		if pkgName == "" {
			pkgName = theASTFile.Name.Name
		}
		out = append(out, theASTFile)
	}
	if len(out) == 0 {
		return "", nil, fmt.Errorf("no importable packages found in %s", path)
	}
	if len(namesSet) > 1 {
		var names []string
		for n := range namesSet {
			names = append(names, n)
		}
		sort.Strings(names)
		return "", nil, fmt.Errorf("multiple packages found in %s: %v", path, strings.Join(names, ", "))
	}
	return pkgName, out, nil
}

// hasContextParams returns whether or not the first parameter is a context.Context. If it
// determines that hte first parameter makes this function invalid for stave, it'll return a non-nil
// error.
func hasContextParam(ft *ast.FuncType) (bool, error) {
	if ft.Params.NumFields() < 1 {
		return false, nil
	}
	param := ft.Params.List[0]
	sel, ok := param.Type.(*ast.SelectorExpr)
	if !ok {
		return false, nil
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok {
		return false, nil
	}
	if pkg.Name != "context" {
		return false, nil
	}
	if sel.Sel.Name != "Context" {
		return false, nil
	}
	if len(param.Names) > 1 {
		// something like foo, bar context.Context
		return false, errors.New("ETOOMANYCONTEXTS")
	}
	return true, nil
}

func hasErrorReturn(ft *ast.FuncType) (bool, error) {
	res := ft.Results
	if res.NumFields() == 0 {
		// void return is ok
		return false, nil
	}
	if res.NumFields() > 1 {
		return false, errors.New("ETOOMANYRETURNS")
	}
	ret := res.List[0]
	if len(ret.Names) > 1 {
		return false, errors.New("ETOOMANYERRORS")
	}
	if fmt.Sprint(ret.Type) == "error" {
		return true, nil
	}
	return false, errors.New("EBADRETURNTYPE")
}

func funcType(ft *ast.FuncType) (*Function, error) {
	var err error
	theFunc := &Function{}
	theFunc.IsContext, err = hasContextParam(ft)
	if err != nil {
		return nil, err
	}
	theFunc.IsError, err = hasErrorReturn(ft)
	if err != nil {
		return nil, err
	}
	iArg := 0
	if theFunc.IsContext {
		iArg++
	}
	for ; iArg < len(ft.Params.List); iArg++ {
		param := ft.Params.List[iArg]
		t := fmt.Sprint(param.Type)
		typ, ok := argTypes[t]
		if !ok {
			return nil, fmt.Errorf("unsupported argument type: %s", t)
		}
		// support for foo, bar string
		for _, name := range param.Names {
			theFunc.Args = append(theFunc.Args, Arg{Name: name.Name, Type: typ})
		}
	}
	return theFunc, nil
}

func toOneLine(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
}
