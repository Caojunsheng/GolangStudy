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
