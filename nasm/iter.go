package nasm

// LineIterator is used for iterating over lines.
type LineIterator struct {
	lines   Lines
	lineIdx int
}

func NewLineIterator(lines Lines) *LineIterator {
	return &LineIterator{
		lines:   lines,
		lineIdx: -1,
	}
}

// Next moves the iterator forward.
func (l *LineIterator) Next() bool {
	l.lineIdx++
	return l.lineIdx < len(l.lines)
}

// At returns the line at the given number.
func (l *LineIterator) At(i int) Line {
	if i < 0 || i >= len(l.lines) {
		return Line{}
	}
	return l.lines[i]
}

// LineNum returns the current line's number.
func (l *LineIterator) LineNum() int {
	return l.lineIdx
}

// Before returns the line before this one.
func (l *LineIterator) Before() Line { return l.At(l.lineIdx - 1) }

// Current returns the current line.
func (l *LineIterator) Current() Line { return l.At(l.lineIdx) }

// After returns the line after this one.
func (l *LineIterator) After() Line { return l.At(l.lineIdx + 1) }
