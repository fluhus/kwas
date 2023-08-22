// Package kmc provides an API over kmc3.
package kmc

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/fluhus/kwas/util"
)

const (
	exePath = "kmc"
)

type reader struct {
	r *io.PipeReader
	c chan int
}

// Read reads from the output of kmc_dump.
func (r *reader) Read(b []byte) (int, error) {
	return r.r.Read(b)
}

// Close finalizes the running instance of kmc and removes temporary files.
func (r *reader) Close() error {
	r.r.Close()
	<-r.c
	return nil
}

func runKMC(fq string, opts ...Option) (io.ReadCloser, error) {
	dir, err := os.MkdirTemp("", "kmc_")
	if err != nil {
		return nil, err
	}
	stderr1 := bytes.NewBuffer(nil)

	// Allow overriding default params.
	paramMap := map[string]string{
		"v": "-v",
		"t": "-t2",
	}
	for _, opt := range opts {
		paramMap[opt.key] = opt.param
	}
	var params []string
	for _, p := range paramMap {
		if p == "" {
			continue
		}
		params = append(params, p)
	}
	params = append(params, fq, filepath.Join(dir, "a"), dir)
	cmd1 := exec.Command(exePath, params...)
	cmd1.Stdout = stderr1
	cmd1.Stderr = stderr1

	r, w := io.Pipe()
	stderr2 := bytes.NewBuffer(nil)
	cmd2 := exec.Command(exePath+"_dump", filepath.Join(dir, "a"), os.Stdout.Name())
	cmd2.Stdout = w
	cmd2.Stderr = stderr2

	result := &reader{r: r, c: make(chan int, 1)}

	go func() {
		defer func() {
			os.RemoveAll(dir)
			result.c <- 1
		}()
		if err := cmd1.Run(); err != nil {
			w.CloseWithError(fmt.Errorf("%s %s", err, stderr1.Bytes()))
			return
		}
		if err := cmd2.Run(); err != nil {
			w.CloseWithError(fmt.Errorf("%s %s", err, stderr2.Bytes()))
			return
		}
		w.CloseWithError(io.EOF)
	}()
	return result, nil
}

func KMC(forEach func(kmer []byte, count int), fq string, opts ...Option) error {
	r, err := runKMC(fq, opts...)
	if err != nil {
		return err
	}
	defer r.Close()
	sc := bufio.NewScanner(r)
	var parts [][]byte
	for sc.Scan() {
		parts = util.SplitBytes(sc.Bytes(), '\t', parts)
		if len(parts) != 2 {
			return fmt.Errorf("bad number of parts in %q: %v, want 2",
				sc.Text(), len(parts))
		}
		count, err := strconv.Atoi(string(parts[1]))
		if err != nil {
			return err
		}
		forEach(parts[0], count)
	}
	return sc.Err()
}

type Option struct {
	key, param string
}

func OptionK(k int) Option {
	return Option{"k", fmt.Sprint("-k", k)}
}

func OptionMin(n int) Option {
	return Option{"ci", fmt.Sprint("-ci", n)}
}

func OptionFasta() Option {
	return Option{"f", "-fm"}
}

func OptionThreads(n int) Option {
	return Option{"t", fmt.Sprint("-t", n)}
}

func OptionMaxCount(n int) Option {
	return Option{"cs", fmt.Sprint("-cs", n)}
}
