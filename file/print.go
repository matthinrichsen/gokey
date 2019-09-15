package file

import (
	"go/ast"
	"go/printer"
	"go/token"
	"io"
)

type RepairInfo struct {
	Lines []int
}

func PrintRepair(out io.Writer, repaired *ast.File, repairInfo RepairInfo) error {
	fset := token.NewFileSet()
	start := repaired.Pos()
	end := repaired.End()
	fset.AddFile(`doesntmatter.go`, int(start), int(end-start+1)).SetLines(repairInfo.Lines)

	cfg := printer.Config{
		Mode:     printer.TabIndent,
		Tabwidth: 8,
	}
	return cfg.Fprint(out, fset, repaired)
}
