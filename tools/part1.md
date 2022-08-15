### 1. golang UT coverage
当想执行某一个package下面所有UT，并且跑完之后显示UT覆盖率以及报告，可以使用下面的命令。

**Installation**

Get the necessary packages and its dependencies:

`$ go get github.com/axw/gocov/gocov`

`$ go get -u gopkg.in/matm/v1/gocov-html`


**Usage**

`$ gocov test ./... | gocov-html > result.html`

**Reference**

https://github.com/axw/gocov

https://github.com/matm/gocov-html

### 2. golang GC and GC heap size set.
打开GC开关：

设置环境变量：GODEBUG=gctrace=1
    
调整GC Heap大小：

设置环境变量：GOGC=400

GOGC默认大小100, golang官方解释：The GOGC variable sets the initial garbage collection target percentage. A collection is triggered when the ratio of freshly allocated data to live data remaining after the previous collection reaches this percentage. The default is GOGC=100. Setting GOGC=off disables the garbage collector entirely. 
```
说明：
gc 1 @0.005s 0%: 0+11+0.99 ms clock, 0+0/15/0+7.9 ms cpu, 25->25->24 MB, 26 MB goal, 8 P
gc 1：说明gc过程，不做解释，1 表示第几次执行gc操作
@0.005s：表示程序执行的总时间
0%: 表示gc时时间和程序运行总时间的百分比
0+11+0.99 ms clock,: wall-clock (还不清楚什么意思)
0+0/15/0+7.9 ms cpu,: CPU times for the phases of the GC (还不清楚什么意思)
25->25->24 MB,: gc开始时的堆大小；GC结束时的堆大小；存活的堆大小
26 MB goal,: 整体堆的大小
8 P: 使用的处理器的数量

说明:
scvg4: inuse: 143, idle: 70, sys: 213, released: 0, consumed: 213 (MB)
inuse: 143,：使用多少M内存
idle: 70,：0 剩下要清除的内存
sys: 213,： 系统映射的内存
released: 0,： 释放的系统内存
consumed: 213： 申请的系统内存

```

### 3. golang 生成pprof的cpu，mem图和火焰图
+ 使用go自带的pprof工具对dump出的文件进行分析
```bash
go tool pprof <code_binary> /tmp/cpu.pprof
go tool pprof <code_binary> /tmp/mem.pprof
pprof>svg
pprof>top
```
+ 使用go-torch生成火焰图
```bash
go get github.com/uber/go-torch
git clone https://github.com/brendangregg/FlameGraph
export PATH=$PATH:/path/FlameGraph
go-torch --binaryname=./<code_binary> --binaryinput=./cpu.pprof
```
### 4.golang使用analysis对代码静态检查
```go
package main

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(Analyzer)
}

var Analyzer = &analysis.Analyzer{
	Name: "firstparamcontext",
	Doc:  "Checks that functions first param type is Context",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	fileConstMap := make(map[string]string)
	inspect := func(node ast.Node) bool {
		switch delc := node.(type) {
		case *ast.GenDecl:
			if delc.Tok == token.CONST && len(delc.Specs) > 0 {
				for i := range delc.Specs {
					if valSpec, ok := delc.Specs[i].(*ast.ValueSpec); ok {
						for j := range valSpec.Names {
							if len(valSpec.Values) == len(valSpec.Names) {
								if valueDefine, ok := valSpec.Values[j].(*ast.BasicLit); ok {
									fileConstMap[valSpec.Names[j].Name] = strings.ReplaceAll(valueDefine.Value, "\"", "")
								}
							}

						}
					}
				}
			}
		case *ast.FuncDecl:
			var receiverName string
			if delc.Recv != nil && len(delc.Recv.List) > 0 {
				if expr, ok := delc.Recv.List[0].Type.(*ast.StarExpr); ok {
					receiverName = getIdentName(expr.X)
				}
			}
			analysisFunErr(pass, delc, receiverName, fileConstMap)
		default:
			return true
		}
		return true
	}

	for _, f := range pass.Files {
		ast.Inspect(f, inspect)
	}
	return nil, nil
}

func analysisFunErr(pass *analysis.Pass, funcDecl *ast.FuncDecl, receiverName string, fileConstMap map[string]string) {
	results := funcDecl.Type.Results
	if results == nil || len(results.List) == 0 {
		return
	}
	returnErrFlag := false
	for i := range results.List {
		if err, ok := results.List[i].Type.(*ast.Ident); ok && err.Name == "error" {
			returnErrFlag = true
			break
		}
	}
	if !returnErrFlag {
		return
	}
	mName := funcDecl.Name.Name
	constMap := make(map[string]string)
	for i := range funcDecl.Body.List {
		switch stmt := funcDecl.Body.List[i].(type) {
		case *ast.DeclStmt:
			if decl, ok := stmt.Decl.(*ast.GenDecl); ok && len(decl.Specs) > 0 {
				for i := range decl.Specs {
					if valSpec, ok := decl.Specs[i].(*ast.ValueSpec); ok && len(valSpec.Names) > 0 {
						for j := range valSpec.Names {
							if len(valSpec.Values) == len(valSpec.Names) {
								if valueDefine, ok := valSpec.Values[j].(*ast.BasicLit); ok {
									constMap[valSpec.Names[j].Name] = strings.ReplaceAll(valueDefine.Value, "\"", "")
								}
							}
						}
					}
				}
			}
		case *ast.IfStmt:
			processIfStmt(pass, stmt, mName, receiverName, constMap, fileConstMap)
		case *ast.ReturnStmt:
			processReturnStmt(pass, stmt, mName, receiverName, constMap, fileConstMap)
		}
	}
}

func processIfStmt(pass *analysis.Pass, stmt *ast.IfStmt, mName, receiverName string, constMap, fileConstMap map[string]string) {
	for i := range stmt.Body.List {
		switch newStmt := stmt.Body.List[i].(type) {
		case *ast.IfStmt:
			processIfStmt(pass, newStmt, mName, receiverName, constMap, fileConstMap)
		case *ast.ReturnStmt:
			processReturnStmt(pass, newStmt, mName, receiverName, constMap, fileConstMap)
		}
	}
}

func processReturnStmt(pass *analysis.Pass, stmt *ast.ReturnStmt, mName, receiverName string, constMap, fileConstMap map[string]string) {
	for i := range stmt.Results {
		if errCall, ok := stmt.Results[i].(*ast.CallExpr); ok {
			if errNew, ok := errCall.Fun.(*ast.SelectorExpr); ok {
				if pkgName, ok := errNew.X.(*ast.Ident); ok && pkgName.Name == "error" {
					if errNew.Sel.Name == "New" {
						if name, ok := fileConstMap[getIdentName(errCall.Args[0])]; ok && name != receiverName {
							pass.Reportf(stmt.Pos(), "error.New use fault receiver name\n")
						}
						if name, ok := constMap[getIdentName(errCall.Args[1])]; ok && name != mName {
							pass.Reportf(stmt.Pos(), "error.New use fault method name\n")
						}
					} else if errNew.Sel.Name == "Wrap" {
						if name, ok := fileConstMap[getIdentName(errCall.Args[1])]; ok && name != receiverName {
							pass.Reportf(stmt.Pos(), "error.Wrap use fault receiver name\n")
						}
						if name, ok := constMap[getIdentName(errCall.Args[2])]; ok && name != mName {
							pass.Reportf(stmt.Pos(), "error.Wrap use fault method name\n")
						}
					}

				}
			}
		}
	}
}

func getIdentName(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

```

### 5.golang Stringer生成方法
参考https://github.com/golang/tools/blob/master/cmd/stringer/stringer.go完成
```go
package main // import "golang.org/x/tools/cmd/stringer"

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/constant"
	"go/format"
	"go/token"
	"go/types"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	typeNames   = flag.String("type", "", "comma-separated list of type names; must be set")
	output      = flag.String("output", "", "output file name; default srcdir/<type>_string.go")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of stringer:\n")
	fmt.Fprintf(os.Stderr, "\tstringer [flags] -type T [directory]\n")
	fmt.Fprintf(os.Stderr, "\tstringer [flags] -type T files... # Must be a single package\n")
	fmt.Fprintf(os.Stderr, "For more information, see:\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("stringer: ")
	flag.Usage = Usage
	flag.Parse()
	if len(*typeNames) == 0 {
		flag.Usage()
		os.Exit(2)
	}
	types := strings.Split(*typeNames, ",")

	// We accept either one directory or a list of files. Which do we have?
	args := flag.Args()
	if len(args) == 0 {
		// Default: process whole package in current directory.
		args = []string{"."}
	}

	// Parse the package once.
	var dir string
	g := Generator{}
	if len(args) == 1 && isDirectory(args[0]) {
		dir = args[0]
	} else {
		dir = filepath.Dir(args[0])
	}

	g.parsePackage(args)

	// Print the header and package clause.
	g.Printf("// Code generated by stringer; DO NOT EDIT!!!\n")
	g.Printf("\n")
	g.Printf("package %s", g.pkg.name)
	g.Printf("\n")
	g.Printf("import \"strconv\"\n") // Used by all methods.

	// Run generate for each type.
	for _, typeName := range types {
		g.generate(typeName)
	}

	// Format the output.
	src := g.format()

	// Write to file.
	outputName := *output
	if outputName == "" {
		baseName := fmt.Sprintf("%s_string.go", types[0])
		outputName = filepath.Join(dir, strings.ToLower(baseName))
	}
	err := ioutil.WriteFile(outputName, src, 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
}

// isDirectory reports whether the named file is a directory.
func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}

// Generator holds the state of the analysis. Primarily used to buffer
// the output for format.Source.
type Generator struct {
	buf bytes.Buffer // Accumulated output.
	pkg *Package     // Package we are scanning.
}

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

// File holds a single parsed file and associated data.
type File struct {
	pkg  *Package  // Package to which this file belongs.
	file *ast.File // Parsed AST.
	// These fields are reset for each type being generated.
	typeName string  // Name of the constant type.
	values   []Value // Accumulator for constant values of that type.
}

type Package struct {
	name  string
	defs  map[*ast.Ident]types.Object
	files []*File
}

// parsePackage analyzes the single package constructed from the patterns and tags.
// parsePackage exits if there is an error.
func (g *Generator) parsePackage(patterns []string) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		// TODO: Need to think about constants in test files. Maybe write type_string_test.go
		// in a separate pass? For later.
		Tests:      false,
		BuildFlags: []string{},
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}
	g.addPackage(pkgs[0])
}

// addPackage adds a type checked Package and its syntax files to the generator.
func (g *Generator) addPackage(pkg *packages.Package) {
	g.pkg = &Package{
		name:  pkg.Name,
		defs:  pkg.TypesInfo.Defs,
		files: make([]*File, len(pkg.Syntax)),
	}

	for i, file := range pkg.Syntax {
		g.pkg.files[i] = &File{
			file:        file,
			pkg:         g.pkg,
		}
	}
}

// generate produces the String method for the named type.
func (g *Generator) generate(typeName string) {
	values := make([]Value, 0, 100)
	for _, file := range g.pkg.files {
		// Set the state for this run of the walker.
		file.typeName = typeName
		file.values = nil
		if file.file != nil {
			ast.Inspect(file.file, file.genDecl)
			values = append(values, file.values...)
		}
	}

	if len(values) == 0 {
		log.Fatalf("no values defined for type %s", typeName)
	}
	// Generate code that will fail if the constants change value.

	g.Printf("\tvar %sMapString = map[%s]string{\n", typeName, typeName)
	for _, v := range values {
		g.Printf("\t%s:\t\"%s\",\n", v.originalName, v.originalName)
	}
	g.Printf("}\n")
	g.buildMap(typeName)
}

// format returns the gofmt-ed contents of the Generator's buffer.
func (g *Generator) format() []byte {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		// Should never happen, but can arise when developing this code.
		// The user can compile the output to see the error.
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return g.buf.Bytes()
	}
	return src
}

// Value represents a declared constant.
type Value struct {
	originalName string // The name of the constant.
	name         string // The name with trimmed prefix.
	// The value is stored as a bit pattern alone. The boolean tells us
	// whether to interpret it as an int64 or a uint64; the only place
	// this matters is when sorting.
	// Much of the time the str field is all we need; it is printed
	// by Value.String.
	value  uint64 // Will be converted to int64 when needed.
	signed bool   // Whether the constant is a signed type.
	str    string // The string representation given by the "go/constant" package.
}

func (v *Value) String() string {
	return v.str
}

// genDecl processes one declaration clause.
func (f *File) genDecl(node ast.Node) bool {
	decl, ok := node.(*ast.GenDecl)
	if !ok || decl.Tok != token.CONST {
		// We only care about const declarations.
		return true
	}
	// The name of the type of the constants we are declaring.
	// Can change if this is a multi-element declaration.
	typ := ""
	// Loop over the elements of the declaration. Each element is a ValueSpec:
	// a list of names possibly followed by a type, possibly followed by values.
	// If the type and value are both missing, we carry down the type (and value,
	// but the "go/types" package takes care of that).
	for _, spec := range decl.Specs {
		vspec := spec.(*ast.ValueSpec) // Guaranteed to succeed as this is CONST.
		if vspec.Type == nil && len(vspec.Values) > 0 {
			// "X = 1". With no type but a value. If the constant is untyped,
			// skip this vspec and reset the remembered type.
			typ = ""

			// If this is a simple type conversion, remember the type.
			// We don't mind if this is actually a call; a qualified call won't
			// be matched (that will be SelectorExpr, not Ident), and only unusual
			// situations will result in a function call that appears to be
			// a type conversion.
			ce, ok := vspec.Values[0].(*ast.CallExpr)
			if !ok {
				continue
			}
			id, ok := ce.Fun.(*ast.Ident)
			if !ok {
				continue
			}
			typ = id.Name
		}
		if vspec.Type != nil {
			// "X T". We have a type. Remember it.
			ident, ok := vspec.Type.(*ast.Ident)
			if !ok {
				continue
			}
			typ = ident.Name
		}
		if typ != f.typeName {
			// This is not the type we're looking for.
			continue
		}
		// We now have a list of names (from one line of source code) all being
		// declared with the desired type.
		// Grab their names and actual values and store them in f.values.
		for _, name := range vspec.Names {
			if name.Name == "_" {
				continue
			}
			// This dance lets the type checker find the values for us. It's a
			// bit tricky: look up the object declared by the name, find its
			// types.Const, and extract its value.
			obj, ok := f.pkg.defs[name]
			if !ok {
				log.Fatalf("no value for constant %s", name)
			}
			info := obj.Type().Underlying().(*types.Basic).Info()
			if info&types.IsInteger == 0 {
				log.Fatalf("can't handle non-integer constant type %s", typ)
			}
			value := obj.(*types.Const).Val() // Guaranteed to succeed as this is CONST.
			if value.Kind() != constant.Int {
				log.Fatalf("can't happen: constant is not an integer %s", name)
			}
			i64, isInt := constant.Int64Val(value)
			u64, isUint := constant.Uint64Val(value)
			if !isInt && !isUint {
				log.Fatalf("internal error: value of %s is not an integer: %s", name, value.String())
			}
			if !isInt {
				u64 = uint64(i64)
			}
			v := Value{
				originalName: name.Name,
				value:        u64,
				signed:       info&types.IsUnsigned == 0,
				str:          value.String(),
			}
			if c := vspec.Comment;  c != nil && len(c.List) == 1 {
				v.name = strings.TrimSpace(c.Text())
			}
			f.values = append(f.values, v)
		}
	}
	return false
}


// buildMap handles the case where the space is so sparse a map is a reasonable fallback.
// It's a rare situation but has simple code.
func (g *Generator) buildMap(typeName string) {
	g.Printf("\n")
	g.Printf(stringMap, typeName)
}

// Argument to format is the type name.
const stringMap = `func (i %[1]s) String() string {
	if str, ok := %[1]sMapString[i]; ok {
		return str
	}
	return "%[1]s(" + strconv.FormatInt(int64(i), 10) + ")"
}
`

```
<!--stackedit_data:
eyJoaXN0b3J5IjpbLTI5ODUyNTgwMywtMTQ2OTA1NTc2OSwtMT
A0NDAyNjIwNywyODQyNDM0NDNdfQ==
-->