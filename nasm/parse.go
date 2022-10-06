package nasm

import (
	"bufio"
	"fmt"
	"io"
)

// Parser is used by token parsers to help parse Assembly lines.
type Parser struct {
	// Lines contains the lines that are already scanned.
	Lines Lines

	scan *bufio.Scanner
	curr string
	next *string
}

// NewParser returns a new Parser for the given reader.
func NewParser(r io.Reader) *Parser {
	return &Parser{
		scan: bufio.NewScanner(r),
	}
}

// PrevLine returns the previous line, if any.
func (p *Parser) PrevLine() *Line {
	if len(p.Lines) == 0 {
		return nil
	}
	return &p.Lines[len(p.Lines)-1]
}

// Scan moves the line iterator forward one line.
func (p *Parser) Scan() bool {
	if p.next != nil {
		p.curr = *p.next
		p.next = nil
		return true
	}

	scan := p.scan.Scan()
	p.curr = p.scan.Text()
	return scan
}

// Text returns the current line.
func (p *Parser) Text() string {
	return p.curr
}

// Peek allows the parser to peek to the next line. Peek can only peek to the
// next line; calling it multiple times will return that same line.
func (p *Parser) Peek() (string, bool) {
	if p.next != nil {
		return *p.next, true
	}

	if p.scan.Scan() {
		text := p.scan.Text()
		p.next = &text
		return text, true
	}

	return "", false
}

// Err returns any scanning error.
func (p *Parser) Err() error {
	if p.next != nil {
		// can't be non-nil if there's no error when peaking
		return nil
	}
	return p.scan.Err()
}

// Parse parses multiple lines (i.e. a whole file).
func Parse(r io.Reader) (Lines, error) {
	parser := NewParser(r)

	for lineIdx := 0; parser.Scan(); lineIdx++ {
		line, err := parseLine(parser)
		if err != nil {
			return parser.Lines, fmt.Errorf("error at line %d: %w", lineIdx+1, err)
		}
		parser.Lines = append(parser.Lines, line)
	}

	return parser.Lines, parser.Err()
}

func parseLine(scanner *Parser) (Line, error) {
	line := scanner.Text()
	if line == "" {
		return Line{}, nil
	}

	var token Token
	var comment CommentToken

	for _, parser := range TokenParsers {
		token, line = parser(scanner, line)

		if _, isComment := token.(CommentToken); isComment {
			comment = token.(CommentToken)
			continue
		}

		if token != nil || line == "" {
			break
		}
	}

	if line != "" {
		if token == nil {
			return Line{}, fmt.Errorf("unknown token %q", line)
		}
		return Line{}, fmt.Errorf("excess text %q", line)
	}

	return Line{
		Token:   token,
		Comment: comment,
	}, nil
}
