package nasmfmt

import (
	"io"
	"strings"
	"text/tabwriter"

	"github.com/diamondburned/nasmfmt/v2/nasm"
)

// FormatConfig is the configuration for the formatter.
type FormatConfig struct {
	// InstructionIndent is the number of spaces to indent instructions by.
	InstructionIndent int
	// CommentIndent is the number of spaces to indent comments by.
	CommentIndent int
}

// Format formats the NASM assembly code from src and writes it to dst.
// It formats it using the default settings.
func Format(dst io.Writer, src io.Reader, cfg FormatConfig) error {
	lines, err := nasm.Parse(src)
	if err != nil {
		return err
	}

	blocks := []nasm.Lines{nil} // slice of 1, intentionally nil!
	addBlock := func() {
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
			addBlock()
			continue
		}

		if _, ok := line.Token.(nasm.SectionToken); ok {
			addBlock()
			addToBlock(line)
			addBlock()
			continue
		}

		addToBlock(line)
	}

	for _, block := range blocks {
		if err := writeBlock(dst, block, cfg); err != nil {
			return err
		}
	}

	return nil
}

func writeBlock(dst io.Writer, block nasm.Lines, cfg FormatConfig) error {
	lines := writeLinesNoComment(block, cfg)

	// Vertical align the lines.
	lines = strings.Split(valign(lines), "\n")

	// Ugly hack to add comments after we tab-align the columns before the
	// comments are added. We're only doing this for the sake of keeping a fixed
	// indentation before inline comments.
	//
	// I actually hate this so much.
	for i, s := range lines {
		if i >= len(block) {
			break
		}

		line := block[i]
		if line.Comment == (nasm.CommentToken{}) {
			continue
		}

		if line.Token != nil {
			switch line.Token.(type) {
			case nasm.InstructionToken:
				indent := cfg.CommentIndent - (len(s) + 1)
				if indent < 1 {
					indent = 1
				}
				s += strings.Repeat(" ", indent)
			default:
				s += "\t"
			}
		} else if i > 0 && lines[i-1] != "" {
			commentIx := strings.Index(nasm.NoQuotes(lines[i-1], "x"), ";")
			if commentIx != -1 {
				s += strings.Repeat(" ", commentIx)
			}
		}

		s += line.Comment.String()
		lines[i] = s
	}

	// Re-vertically align the lines.
	out := valign(lines)

	_, err := dst.Write([]byte(out))
	return err
}

func writeLinesNoComment(lines nasm.Lines, cfg FormatConfig) []string {
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
				s.WriteString(strings.Repeat(" ", cfg.InstructionIndent))
			}

			s.WriteString(line.Token.String())
		}

		strs[iter.LineNum()] = s.String()
	}

	return strs
}

func valign(lines []string) string {
	var buf strings.Builder
	tabw := tabwriter.NewWriter(&buf, 1, 0, 1, ' ', 0)

	for _, line := range lines {
		tabw.Write([]byte(line))
		tabw.Write([]byte("\n"))
	}

	tabw.Flush()

	return buf.String()
}
