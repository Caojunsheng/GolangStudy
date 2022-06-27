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
 "go/token" "strings"  
 "golang.org/x/tools/go/analysis" "golang.org/x/tools/go/analysis/singlechecker")  
  
func main() {  
   singlechecker.Main(Analyzer)  
}  
  
var Analyzer = &analysis.Analyzer{  
   Name: "smerror correct check",  
  Doc:  "Checks that smerror first param type is Context",  
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
 break  }  
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
      if smerrCall, ok := stmt.Results[i].(*ast.CallExpr); ok {  
         if smerrNew, ok := smerrCall.Fun.(*ast.SelectorExpr); ok {  
            if pkgName, ok := smerrNew.X.(*ast.Ident); ok && pkgName.Name == "smerror" {  
               if smerrNew.Sel.Name == "New" {  
                  if name, ok := fileConstMap[getIdentName(smerrCall.Args[0])]; ok && name != receiverName {  
                     pass.Reportf(stmt.Pos(), "error.New use fault receiver name\n")  
                  }  
                  if name, ok := constMap[getIdentName(smerrCall.Args[1])]; ok && name != mName {  
                     pass.Reportf(stmt.Pos(), "error.New use fault method name\n")  
                  }  
               } else if smerrNew.Sel.Name == "Wrap" {  
                  if name, ok := fileConstMap[getIdentName(smerrCall.Args[1])]; ok && name != receiverName {  
                     pass.Reportf(stmt.Pos(), "error.Wrap use fault receiver name\n")  
                  }  
                  if name, ok := constMap[getIdentName(smerrCall.Args[2])]; ok && name != mName {  
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
<!--stackedit_data:
eyJoaXN0b3J5IjpbNTk3MDMwMzQxXX0=
-->