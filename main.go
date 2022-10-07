package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/diamondburned/nasmfmt/v2/nasm"
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
	if file == "-" {
		return format(os.Stdout, os.Stdin)
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

	if err := format(dstbuf, src); err != nil {
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

func format(dst io.Writer, src io.Reader) error {
	lines, err := nasm.Parse(src)
	if err != nil {
		return err
	}

	blocks := []nasm.Lines{nil}

	newBlock := func() {
		if blocks[len(blocks)-1] != nil {
			blocks = append(blocks, nil)
		}
	}

	addToBlockN := func(line nasm.Line, n int) {
		blocks[len(blocks)-n] = append(blocks[len(blocks)-n], line)
	}

	addToBlock := func(line nasm.Line) {
		addToBlockN(line, 1)
	}

	iter := nasm.NewLineIterator(lines)
	for iter.Next() {
		line := iter.Current()

		if line.IsEmpty() {
			newBlock()
			continue
		}

		if _, ok := line.Token.(nasm.SectionToken); ok {
			newBlock()
			addToBlock(line)
			newBlock()
			continue
		}

		addToBlock(line)
	}

	for _, block := range blocks {
		if err := writeBlock(dst, block); err != nil {
			return err
		}
	}

	return nil
}

func writeBlock(dst io.Writer, block nasm.Lines) error {
	lines := writeLinesNoComment(block)

	var buf strings.Builder
	tabw := tabwriter.NewWriter(&buf, 1, 0, 1, ' ', 0)

	for _, line := range lines {
		tabw.Write([]byte(line))
		tabw.Write([]byte("\n"))
	}

	tabw.Flush()

	// Ugly hack to add comments after we tab-align the columns before the
	// comments are added. We're only doing this for the sake of keeping a fixed
	// indentation before inline comments.
	//
	// I actually hate this so much.
	lines = strings.Split(buf.String(), "\n")
	for i, s := range lines {
		if i >= len(block) {
			break
		}

		line := block[i]
		if line.Comment == (nasm.CommentToken{}) {
			continue
		}

		if line.Token != nil {
			var indent int

			_, instr := line.Token.(nasm.InstructionToken)
			if instr {
				indent = commentIndent - (len(s) + 1)
				if indent < 1 {
					indent = 1
				}
			} else {
				indent = 1
			}

			s += strings.Repeat(" ", indent)
		} else if i > 0 && lines[i-1] != "" {
			commentIx := strings.Index(nasm.NoQuotes(lines[i-1], "x"), ";")
			if commentIx != -1 {
				s += strings.Repeat(" ", commentIx)
			}
		}

		s += line.Comment.String()
		lines[i] = s
	}

	s := strings.Join(lines, "\n")
	s += "\n"

	_, err := dst.Write([]byte(s))
	return err
}

func writeLinesNoComment(lines nasm.Lines) []string {
	strs := make([]string, len(lines))

	iter := nasm.NewLineIterator(lines)
	for iter.Next() {
		line := iter.Current()
		if line.IsEmpty() {
			continue
		}

		var s strings.Builder

		if line.Token != nil {
			_, instr := line.Token.(nasm.InstructionToken)
			if instr {
				s.WriteString(strings.Repeat(" ", insIndent))
			}

			s.WriteString(line.Token.String())
		}

		strs[iter.LineNum()] = s.String()
	}

	return strs
}
