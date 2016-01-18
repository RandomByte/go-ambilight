## Performance analysis

### Tooling
    go test -v -run=^$ -bench=. -benchtime=2s -cpuprofile=prof.cpu | tee perfAnalysis/cpuX
    benchcmp perfAnalysis/cpu0 perfAnalysis/cpu1
    profcpu


0. Converting YCbCr image to RGBA: **-20.58%** (cpu-1 to cpu0)
1. Skipping ````resize.Thumbnail```` in dominantcolor.go: **+53.55%** (cpu0 to cpu1)
2. ~~Replacing ````.At```` in dominantcolor.go access with direct ````.Pix```` access: -88.65% (cpu0 to cpu2)~     | *did actually make all clusters black*
3. ~~Doing #2 again at another place in dominantcolor.go: -81.57% (cpu2 to cpu3); -97.91% (cpu0 to cpu3)~~          | *did actually make the result just black*
4. Fixed dominantcolor.go to do the right thing again (#2 and #3 got invalidated here): **-68.90%** (cpu0 to cpu4)
