package file

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
)

type RepairInfo struct {
	Lines []int
}

func PrintRepair(repaired *ast.File, repairInfo RepairInfo) ([]byte, error) {
	fset := token.NewFileSet()
	start := repaired.Pos()
	end := repaired.End()
	fset.AddFile(`doesntmatter.go`, int(start), int(end-start+1)).SetLines(repairInfo.Lines)

	cfg := printer.Config{
		Mode:     printer.TabIndent,
		Tabwidth: 8,
	}
	b := bytes.NewBuffer(nil)
	err := cfg.Fprint(b, fset, repaired)
	if err != nil {
		return nil, err
	}

	return format.Source(b.Bytes())
}
