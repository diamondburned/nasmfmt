package nasm

// NoQuotes replaces all quoted parts of a string with provided replacement.
// E.g. NoQuotes(`I'm "in love" with donuts`, "x") -> `I'm xxxxxxxxx with donuts`.
//
// It's useful for performing substring searches ignoring quotations.
// Index of a substring in a 'NoQuotes' version with len(rep)==1
// would equal its index in original string.
func NoQuotes(s, rep string) string {
	sr := []rune(s)
	repr := []rune(rep)
	outr := []rune{}

	for len(sr) > 0 {
		// Find first quotation mark
		ind := indexRuneAny(sr, `"'`)
		if ind >= 0 {
			quote := sr[ind]
			// Find its pair
			ind2 := ind + 1 + indexRune(sr[ind+1:], quote)
			// If it has no pair - include it into output string and skip
			if ind2 < ind+1 {
				outr = append(outr, sr[:ind+1]...)
				sr = sr[ind+1:]
				continue
			}
			// If it's paired - replace it with reps
			outr = append(outr, sr[:ind]...)
			outr = append(outr, runesRepeat(repr, ind2-ind+1)...)
			sr = sr[ind2+1:]
		} else {
			// If no quotation marks found -
			// include the rest of input string and break the cycle.
			outr = append(outr, sr...)
			sr = nil
		}
	}

	return string(outr)
}

func indexRune(s []rune, rn rune) int {
	for i := range s {
		if s[i] == rn {
			return i
		}
	}
	return -1
}

func indexRuneAny(s []rune, any string) int {
	anyr := []rune(any)
	for i, rn := range s {
		if indexRune(anyr, rn) >= 0 {
			return i
		}
	}
	return -1
}

func runesRepeat(rep []rune, count int) []rune {
	ret := []rune{}
	for i := 0; i < count; i++ {
		ret = append(ret, rep...)
	}
	return ret
}
