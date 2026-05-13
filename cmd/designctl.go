package main

import (
	"os"
	"path/filepath"
)

func main() {
	rootDir, err := filepath.Abs("../")
	if err != nil {
		panic(err)
	}

	dirs := []string{
		"creational",
		"structural",
		"behavioral",
	}

	for _, dir := range dirs {
		subDir, err := os.ReadDir(filepath.Join(rootDir, dir))
		if err != nil {
			panic(err)
		}
		println("Directory:", filepath.Join(rootDir, dir))
		for _, entry := range subDir {
			if !entry.IsDir() {
				continue
			}
			goDir := filepath.Join(rootDir, dir, entry.Name(), "go")
			if _, err := os.Stat(goDir); os.IsNotExist(err) {
				err = os.MkdirAll(goDir, os.ModePerm)
				if err != nil {
					panic(err)
				}
			}

			goFile := filepath.Join(goDir, entry.Name()+".go")
			if _, err := os.Stat(goFile); !os.IsNotExist(err) {
				continue
			}

			f, err := os.Create(goFile)
			if err != nil {
				panic(err)
			}
			f.WriteString("package main\n\nfunc main() {\n\t// TODO: Implement design pattern\n}\n")
			f.Close()

			goModFile := filepath.Join(goDir, "go.mod")
			if _, err := os.Stat(goModFile); !os.IsNotExist(err) {
				continue
			}

			f, err = os.Create(goModFile)
			if err != nil {
				panic(err)
			}
			f.WriteString("module " + "github.com/hanzhuoxian/design-pattern/" + dir + "/" + entry.Name() + "/go\n\ngo 1.26.3\n")
			f.Close()
			println("  -", goDir)
		}
	}

	println("Root Directory:", rootDir)
}
