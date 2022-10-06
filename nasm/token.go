package nasm

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type IndentOpts struct {
	Instruction int
	Comment     int
}

var DefaultIndentOpts = IndentOpts{
	Instruction: 8,
	Comment:     40,
}

// Lines consists of multiple lines.
type Lines []Line

func (ls Lines) String() string {
	var b strings.Builder
	for _, l := range ls {
		b.WriteString(l.String())
		b.WriteByte('\n')
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// Line consists of multiple tokens.
type Line struct {
	Token   Token
	Comment CommentToken
}

func (l Line) IsEmpty() bool {
	return l.Token == nil && l.Comment == (CommentToken{})
}

// String formats a line.
func (l Line) String() string {
	var b strings.Builder
	if l.Token != nil {
		b.WriteString(l.Token.String())
		b.WriteByte('\t')
	}
	if l.Comment != (CommentToken{}) {
		b.WriteString(l.Comment.String())
	}
	return strings.TrimSuffix(b.String(), "\t")
}

type Token interface {
	fmt.Stringer
	token()
}

func (CommentToken) token()     {}
func (SectionToken) token()     {}
func (DirectiveToken) token()   {}
func (PseudoToken) token()      {}
func (LabelToken) token()       {}
func (InstructionToken) token() {}

type TokenParser func(*Parser, string) (Token, string)

var TokenParsers = []TokenParser{
	ParseCommentToken,   // trims end of line
	ParseSectionToken,   // matches whole line
	ParseDirectiveToken, // matches whole line
	ParsePseudoToken,    // matches whole line
	ParseLabelToken,
	ParseInstructionToken, // matches whole line
}

type LabelToken struct {
	Label string
}

func ParseLabelToken(parser *Parser, line string) (Token, string) {
	noq := NoQuotes(line, "x")

	idx := strings.Index(noq, ":")
	if idx == -1 {
		return nil, line
	}

	label := strings.TrimSpace(line[:idx])

	rest := line[idx+1:]
	if strings.Contains(rest, "[") && strings.Contains(rest, "]") {
		return nil, line
	}

	line = rest
	return LabelToken{label}, rest
}

func (t LabelToken) String() string {
	return t.Label + ":"
}

type InstructionToken struct {
	Instr string
	Args  []string
}

var instrRe = regexp.MustCompile(`\s*(\w+)`)

func ParseInstructionToken(parser *Parser, line string) (Token, string) {
	line = strings.TrimLeftFunc(line, unicode.IsSpace)
	noq := NoQuotes(line, "x")

	instrIdx := instrRe.FindStringSubmatchIndex(noq)
	if instrIdx == nil {
		return nil, line
	}

	instr := line[instrIdx[2]:instrIdx[3]]
	token := InstructionToken{Instr: instr}

	rest := line[instrIdx[3]:]
	args := strings.Split(rest, ",")

	for i, arg := range args {
		args[i] = strings.TrimSpace(arg)
	}

	token.Args = args
	return token, ""
}

func (t InstructionToken) String() string {
	s := t.Instr
	if len(t.Args) > 0 {
		s += "\t"
		s += strings.Join(t.Args, ", ")
	}
	return s
}

type SectionToken struct {
	Keyword string
	Name    string
}

// sectionRe matches whole line.
var sectionRe = regexp.MustCompile(`^(?i)\s*(section|segment)\s+([^;\s]*)\s*$`)

func ParseSectionToken(parser *Parser, line string) (Token, string) {
	noq := NoQuotes(line, "x")

	ind := sectionRe.FindStringSubmatchIndex(noq)
	if ind == nil {
		return nil, line
	}

	return SectionToken{
		Keyword: line[ind[2]:ind[3]],
		Name:    line[ind[4]:ind[5]],
	}, ""
}

func (t SectionToken) String() string {
	return t.Keyword + " " + t.Name
}

var directiveKeywords = []string{
	"default",
	"absolute",
	"extern",
	"global",
	"common",
	"cpu",
	"float",
}

var directiveRe = regexp.MustCompile(fmt.Sprintf(
	`^(?i)\s*(%s)\s+([^;\s]*)\s*$`,
	strings.Join(directiveKeywords, "|"),
))

type DirectiveToken struct {
	Keyword string
	Text    string
}

func ParseDirectiveToken(parser *Parser, line string) (Token, string) {
	noq := NoQuotes(line, "x")

	ind := directiveRe.FindStringSubmatchIndex(noq)
	if ind == nil {
		return nil, line
	}

	return DirectiveToken{
		Keyword: line[ind[2]:ind[3]],
		Text:    line[ind[4]:ind[5]],
	}, ""
}

func (t DirectiveToken) String() string {
	return t.Keyword + " " + t.Text
}

var pseudoKeywords = []string{
	// How do these have different things?!
	// https://www.tortall.net/projects/yasm/manual/html/nasm-pseudop.html
	// https://github.com/hyqneuron/assembler/blob/master/doc/manual/nasm-language.txt#pseudo-instructions
	// https://www.nasm.us/doc/nasmdoc3.html
	"db", "dw", "dd", "dq", "dt", "ddq", "do",
	"resb", "resw", "resd", "resq", "rest", "resdq", "reso", "resy", "resz",
	"incbin",
	"equ",
	"times",
}

var pseudoRe = regexp.MustCompile(fmt.Sprintf(
	`(?i)(.*?):?(?:\s|^)(%s)(?:\s|")`,
	strings.Join(pseudoKeywords, "|"),
))

type PseudoToken struct {
	Label string
	Instr string
	Text  string
}

func ParsePseudoToken(parser *Parser, line string) (Token, string) {
	noq := NoQuotes(line, "x")

	idx := pseudoRe.FindStringSubmatchIndex(noq)
	if idx == nil {
		return nil, line
	}

	return PseudoToken{
		Label: strings.TrimSpace(line[idx[2]:idx[3]]),
		Instr: line[idx[4]:idx[5]],
		Text:  strings.TrimSpace(line[idx[5]:]),
	}, ""
}

func (t PseudoToken) String() string {
	return t.Label + "\t" + t.Instr + "\t" + t.Text
}

type CommentToken struct {
	Comment string
}

func ParseCommentToken(parser *Parser, line string) (Token, string) {
	noq := NoQuotes(line, "x")

	idx := strings.Index(noq, ";")
	if idx == -1 {
		return nil, line
	}

	cmt := line[idx+1:]
	if cmt != " " {
		cmt = strings.TrimPrefix(cmt, " ")
	}

	return CommentToken{cmt}, line[:idx]
}

func (t CommentToken) String() string {
	return "; " + t.Comment
}
