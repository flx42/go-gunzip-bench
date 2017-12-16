# go-gunzip-bench

Benchmark multiple ways of decompressing a `gzip` file in golang.  

Published as a support for https://github.com/golang/go/issues/23154

## Test configuration
```sh
$ grep -m 1 'model name' /proc/cpuinfo
model name      : Intel(R) Core(TM) i7-5930K CPU @ 3.50GHz

$ go version
go version go1.9.2 linux/amd64
```

## Preparation
```sh
$ go build -ldflags="-s -w"

$ mkdir -p tmp

$ sudo mount -t tmpfs tmpfs tmp/

$ wget -nv https://developer.download.nvidia.com/compute/redist/cudnn/v7.0.5/cudnn-9.0-linux-x64-v7.tgz -O tmp/cudnn.tgz
2017-12-15 18:03:22 URL:https://developer.download.nvidia.com/compute/redist/cudnn/v7.0.5/cudnn-9.0-linux-x64-v7.tgz [348817823/348817823] -> "tmp/cudnn.tgz" [1]

# Note: we use `runtime.GOMAXPROCS(1)`
$ echo performance | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
performance
```

## [Method 0](https://github.com/flx42/go-gunzip-bench/blob/master/main.go#L25-L39)
Don't use "compress/gzip", pipe to gunzip(1)
```sh
$ ./go-gunzip-bench 0 tmp/cudnn.tgz
tmp/cudnn.tgz:  4.426812529s

36b429f6f780ab46d6dfd5888918968cd5882ef6b6f4cbd97d596a2da211a4c7  tmp/cudnn.tar
```

## [Method 1](https://github.com/flx42/go-gunzip-bench/blob/master/main.go#L41-L56)
Chain two readers, low memory usage, most idiomatic solution
```sh
$ ./go-gunzip-bench 1 tmp/cudnn.tgz
tmp/cudnn.tgz:  7.84219111s

36b429f6f780ab46d6dfd5888918968cd5882ef6b6f4cbd97d596a2da211a4c7  tmp/cudnn.tar
```

## [Method 2](https://github.com/flx42/go-gunzip-bench/blob/master/main.go#L58-L72)
Read the whole file in-memory, stream decompress/write to output file.
```sh
$ ./go-gunzip-bench 2 tmp/cudnn.tgz
tmp/cudnn.tgz:  7.692913915s

36b429f6f780ab46d6dfd5888918968cd5882ef6b6f4cbd97d596a2da211a4c7  tmp/cudnn.tar
```

## [Method 3](https://github.com/flx42/go-gunzip-bench/blob/master/main.go#L75-L87)
Read the whole file in-memory, and decompress the whole file in-memory.
```sh
$ ./go-gunzip-bench 3 tmp/cudnn.tgz
tmp/cudnn.tgz:  7.941778804s

36b429f6f780ab46d6dfd5888918968cd5882ef6b6f4cbd97d596a2da211a4c7  tmp/cudnn.tar
```

## [Method 4](https://github.com/flx42/go-gunzip-bench/blob/master/main.go#L89-L104)
Method 1 but using [cgzip](https://github.com/youtube/vitess/tree/master/go/cgzip), a golang wrapper for [zlib](https://www.zlib.net) (using cgo).
```sh
$ ./go-gunzip-bench 4 tmp/cudnn.tgz
tmp/cudnn.tgz:  3.403516352s

36b429f6f780ab46d6dfd5888918968cd5882ef6b6f4cbd97d596a2da211a4c7  tmp/cudnn.tar
```

## [`gunzip(1)`](https://www.gnu.org/software/gzip/manual/gzip.html)
`gunzip(1)` for read, decompress, write.
```sh
$ /usr/bin/time gunzip --keep --force tmp/cudnn.tgz && sha256sum tmp/cudnn.tar
4.21user 0.21system 0:04.42elapsed 99%CPU (0avgtext+0avgdata 1788maxresident)k
0inputs+0outputs (0major+157minor)pagefaults 0swaps
36b429f6f780ab46d6dfd5888918968cd5882ef6b6f4cbd97d596a2da211a4c7  tmp/cudnn.tar
```

## Cleanup
```sh
$ sudo umount tmp/

$ rm go-gunzip-bench
```