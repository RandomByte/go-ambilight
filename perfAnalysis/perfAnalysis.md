## Performance analysis

### Tooling
    go test -v -run=^$ -bench=. -benchtime=2s -cpuprofile=prof.cpu | tee perfAnalysis/XXX
    benchcmp perfAnalysis/XXX perfAnalysis/YYY
    profcpu

### BenchmarkComputeDominatorColors
    go test -v -run=^$ -bench=^BenchmarkComputeDominatorColors$ -benchtime=2s -cpuprofile=prof.cpu | tee perfAnalysis/cpuX
    benchcmp perfAnalysis/cpu0 perfAnalysis/cpu1

Pre-baseline: cpu-1

Baseline: Converting YCbCr image to RGBA: **-20.58%** (cpu-1 to cpu0)

1. Skipping ````resize.Thumbnail```` in dominantcolor.go: **+53.55%** (cpu0 to cpu1)
2. ~~Replacing ````.At```` in dominantcolor.go access with direct ````.Pix```` access: -88.65% (cpu0 to cpu2)~     | *did actually make all clusters black*
3. ~~Doing #2 again at another place in dominantcolor.go: -81.57% (cpu2 to cpu3); -97.91% (cpu0 to cpu3)~~          | *did actually make the result just black*
4. Fixed dominantcolor.go to do the right thing again (#2 and #3 got invalidated here): **-68.90%** (cpu0 to cpu4)
5. _(verification) Removed the multi threading in goambi: **+105.31%%** (cpu4 to cpu5)_
6. _(verification) Back to multi threading: **-52.21%** (cpu4 to cpu6); -1.88% (cpu4 to cpu6)_

### BenchmarkLoadAndCompute
    benchcmp perfAnalysis/lac.cpu0 perfAnalysis/lac.cpu1
    go test -v -run=^$ -bench=^BenchmarkLoadImageAndComputeDominatorColors$ -benchtime=2s -cpuprofile=prof.cpu | tee perfAnalysis/lac.cpuX

1. Nothing done so far..


### Learned
- GOMAXPROCS default is set to the count of CPUs available on the current system
- Benchmark results are tied to the CPU count (e.g. can't compare a run on two cores to a run on four) - the number after the Benchmark name indicates the count of CPUs used for the benchmark (which is defined by GOMAXPROCS)