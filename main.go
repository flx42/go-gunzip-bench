package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/klauspost/pgzip"
	"github.com/youtube/vitess/go/cgzip"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func check(err error) {
	if err != nil {
		log.Panicln("Fatal:", err)
	}
}

// Method 0: don't use "compress/gzip", pipe to gunzip(1)
func method0(dst, src string) {
	f, err := os.Open(src)
	check(err)
	defer f.Close()

	w, err := os.Create(dst)
	check(err)
	defer w.Close()

	cmd := exec.Command("gunzip")
	cmd.Stdin = f
	cmd.Stdout = w
	check(cmd.Run())
}

// Method 1: chain two readers, low memory usage, most idiomatic solution
func method1(dst, src string) {
	f, err := os.Open(src)
	check(err)
	defer f.Close()

	r, err := gzip.NewReader(f)
	check(err)

	w, err := os.Create(dst)
	check(err)
	defer w.Close()

	_, err = io.Copy(w, r)
	check(err)
}

// Method 2: read the whole file in-memory, stream decompress/write to output file.
func method2(dst, src string) {
	fb, err := ioutil.ReadFile(src)
	check(err)

	r, err := gzip.NewReader(bytes.NewReader(fb))
	check(err)

	w, err := os.Create(dst)
	check(err)
	defer w.Close()

	_, err = io.Copy(w, r)
	check(err)
}

// Method 3: read the whole file in-memory, and decompress the whole file in-memory.
func method3(dst, src string) {
	fb, err := ioutil.ReadFile(src)
	check(err)

	r, err := gzip.NewReader(bytes.NewReader(fb))
	check(err)

	rb, err := ioutil.ReadAll(r)
	check(err)

	err = ioutil.WriteFile(dst, rb, 0666)
	check(err)
}

// Method 4: Method 1 but using cgzip, a golang wrapper for zlib (using cgo).
func method4(dst, src string) {
	f, err := os.Open(src)
	check(err)
	defer f.Close()

	r, err := cgzip.NewReader(f)
	check(err)

	w, err := os.Create(dst)
	check(err)
	defer w.Close()

	_, err = io.Copy(w, r)
	check(err)
}

// Method 5: Method 1 but using pgzip: https://github.com/klauspost/pgzip
func method5(dst, src string) {
	f, err := os.Open(src)
	check(err)
	defer f.Close()

	r, err := pgzip.NewReader(f)
	check(err)

	w, err := os.Create(dst)
	check(err)
	defer w.Close()

	_, err = io.Copy(w, r)
	check(err)
}

func warmPageCache(path string) {
	f, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	check(err)
	defer f.Close()

	cmd := exec.Command("cat", path)
	cmd.Stdout = f
	check(cmd.Run())
}

func sha256sum(path string) {
	cmd := exec.Command("sha256sum", path)
	cmd.Stdout = os.Stdout
	check(cmd.Run())
}

func main() {
	// Just in case, to avoid too much interference from the go runtime.
	runtime.GOMAXPROCS(1)
	if len(os.Args) != 3 {
		os.Exit(2)
	}

	method, err := strconv.Atoi(os.Args[1])
	if err != nil {
		os.Exit(2)
	}

	src := os.Args[2]
	dst := strings.Replace(src, ".tgz", ".tar", -1)

	// cat src > /dev/null
	warmPageCache(src)

	start := time.Now()
	switch method {
	case 0:
		method0(dst, src)
	case 1:
		method1(dst, src)
	case 2:
		method2(dst, src)
	case 3:
		method3(dst, src)
	case 4:
		method4(dst, src)
	case 5:
		method5(dst, src)
	default:
		os.Exit(2)
	}
	elapsed := time.Since(start)
	fmt.Printf("%s:  %s\n\n", src, elapsed)

	sha256sum(dst)
}
