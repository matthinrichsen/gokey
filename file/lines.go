package file

import (
	"go/token"
)

func getLineTokenPositions(tokenFile *token.File) []token.Pos {
	lines := map[int]struct{}{}
	linePos := []token.Pos{}

	for i := 0; i < tokenFile.Size(); i++ {
		l := tokenFile.Line(token.Pos(tokenFile.Base() + i))
		if _, ok := lines[l]; !ok {
			lines[l] = struct{}{}
			linePos = append(linePos, token.Pos(tokenFile.Base()+i)-1)
		}
	}

	return linePos
}

func getLineOffsets(lines []token.Pos, tokenFile *token.File) []int {
	adjustedLines := make([]int, len(lines))
	for i, l := range lines {
		adjustedLines[i] = int(l) - tokenFile.Base() + 1
	}
	return adjustedLines
}
