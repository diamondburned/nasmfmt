package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/diamondburned/nasmfmt/v2/nasmfmt"
)

var (
	insIndent     int
	commentIndent int
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [params] [files...]\nParameters:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.IntVar(&insIndent, "ii", 8, "Indentation for instructions in spaces")
	flag.IntVar(&commentIndent, "ci", 40, "Indentation for comments in spaces")
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	for _, file := range flag.Args() {
		if err := formatFile(file); err != nil {
			log.Fatalf("cannot format file %q: %v", file, err)
		}
	}
}

func formatFile(file string) error {
	cfg := nasmfmt.FormatConfig{
		InstructionIndent: insIndent,
		CommentIndent:     commentIndent,
	}

	if file == "-" {
		return nasmfmt.Format(os.Stdout, os.Stdin, cfg)
	}

	src, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("cannot open: %w", err)
	}
	defer src.Close()

	dst, err := os.CreateTemp(filepath.Dir(file), ".~*"+filepath.Ext(file))
	if err != nil {
		return fmt.Errorf("cannot create temp: %w", err)
	}
	defer os.Remove(dst.Name())
	defer dst.Close()

	dstbuf := bufio.NewWriter(dst)
	defer dstbuf.Flush()

	if err := nasmfmt.Format(dstbuf, src, cfg); err != nil {
		return err
	}

	if err := dstbuf.Flush(); err != nil {
		return fmt.Errorf("cannot flush write buffer: %w", err)
	}

	if err := dst.Close(); err != nil {
		return fmt.Errorf("cannot close written temp: %w", err)
	}

	if err := os.Rename(dst.Name(), file); err != nil {
		return fmt.Errorf("cannot mv to commit write: %w", err)
	}

	return nil
}
