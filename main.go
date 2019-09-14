package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/matthinrichsen/gokey/file"
	"github.com/matthinrichsen/gokey/util"
)

func main() {
	fixDirectory(``)
}

func fixDirectory(path string) {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	fileSet := token.NewFileSet()
	sn := util.NewStructManager()

	_ = filepath.Walk(path, func(directory string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		if _, folderName := filepath.Split(directory); folderName == `vendor` {
			return filepath.SkipDir
		}

		importDir, err := filepath.Rel(filepath.Join(os.Getenv(`GOPATH`), `src`), directory)
		if err != nil {
			importDir = directory
		}

		astinfo, allFiles, err := utils.CompileFilesInDirectory(directory, fileSet)
		if err != nil {
			return nil
		}

		sn.AddPackage(importDir, astinfo)
		buildOutImports(allFiles, fileSet, sn)

		for filename, f := range allFiles {
			if file.Repair(f, importDir, sn) {
				wd, _ := os.Getwd()
				reportFile, err := filepath.Rel(wd, filename)
				if err != nil {
					reportFile = filename
				}
				fmt.Println(reportFile)
				b := &bytes.Buffer{}
				printer.Fprint(b, fileSet, f)

				formatted, err := format.Source(b.Bytes())
				if err != nil {
					formatted = b.Bytes()
				}

				ioutil.WriteFile(filename, formatted, info.Mode())
			}
		}
		return nil
	})
}

func buildOutImports(files map[string]*ast.File, fileSet *token.FileSet, sn util.StructManager) {
	for _, f := range files {
		for _, i := range f.Imports {
			if sn.HasPackage(i.Path.Value) {
				continue
			}

			info, nextRoundOfFiles, err := utils.CompileFilesInDirectory(filepath.Join(os.Getenv("GOPATH"), "src", util.RemoveQuotes(i.Path.Value)), fileSet)
			if err == nil {
				sn.AddPackage(i.Path.Value, info)
				buildOutImports(nextRoundOfFiles, fileSet, sn)
			}
		}
	}
}
