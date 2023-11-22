package main

// TODO: Config file with file extentions and programs

import (
	//"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	quiet = true
)

func printlnFunc(args ...any) {
	if !quiet {
		fmt.Println(args...)
	}
}

func main() {
	log.SetFlags(0)

	flag.BoolVar(
		&quiet, "q", false,
		"Quiet (only display errors and output from commands)",
	)
	flag.Parse()

	files := flag.Args()
	fileTypes := make(map[string][]string)
	for _, file := range files {
		ext := filepath.Ext(file)
		fileTypes[ext] = append(fileTypes[ext], file)
	}
	for ft, files := range fileTypes {
		var cmd *exec.Cmd
		switch ft {
		case ".c", ".cpp", ".cc", ".h", ".hpp", ".js":
			cmd = makeCmd("clang-format", append([]string{"-i"}, files...)...)
		case ".go":
			cmd = makeCmd("go", append([]string{"fmt"}, files...)...)
		case ".json":
			cmd = makeCmd("clang-format", append([]string{"-i"}, files...)...)
		case ".rs":
			cmd = makeCmd("rustfmt", append([]string{"--edition=2021"}, files...)...)
		default:
			log.Printf("Unsupported filetype: %q (files: %v)", ft, files)
			continue
		}
		printlnFunc("Running:", strings.Join(cmd.Args, " "))
		if err := cmd.Run(); err != nil {
			log.Printf("Error running command for filetype %s: %v", ft, err)
		}
	}
}

func makeCmd(prog string, args ...string) *exec.Cmd {
	cmd := exec.Command(prog, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd
}
