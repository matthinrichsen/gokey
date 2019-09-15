package main

import (
	"fmt"
	"go/ast"
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

		astinfo, allFiles, err := util.CompileFilesInDirectory(directory, fileSet)
		if err != nil {
			return nil
		}

		sn.AddPackage(importDir, astinfo)
		buildOutImports(allFiles, fileSet, sn)

		for filename, f := range allFiles {
			repairInfo, needsRepair := file.Repair(f, importDir, sn, fileSet)
			if needsRepair {
				wd, _ := os.Getwd()
				reportFile, err := filepath.Rel(wd, filename)
				if err != nil {
					reportFile = filename
				}

				fmt.Println(reportFile)

				repairedByes, err := file.PrintRepair(f, repairInfo)
				if err != nil {
					fmt.Printf("issue repairing %s: %v\n", filename, err)
					continue
				}

				ioutil.WriteFile(filename, repairedByes, info.Mode())
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

			info, nextRoundOfFiles, err := util.CompileFilesInDirectory(filepath.Join(os.Getenv("GOPATH"), "src", util.RemoveQuotes(i.Path.Value)), fileSet)
			if err == nil {
				sn.AddPackage(i.Path.Value, info)
				buildOutImports(nextRoundOfFiles, fileSet, sn)
			}
		}
	}
}
