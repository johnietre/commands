package main
/*
 * Print files types that have templates in help printout
 * Allow multiple files to be input using bracket notation like with unix
 * Use "os/exec" and "syscall" to open things like vim and VSCode instead of C
 * Create custom usage function
 * In creating both header and source files (header as base), the include in
    the source includes the cpp file, not the hpp file
 * Allow for exact searches (ex: search for "car" match ".car" and " car " but not "cars")
 * Print "error: " before errors
 * Make #include in source of header and source created with -b include .h and not .c
 * Allow matching whole words
 * Add "-e" flag to create (and overwrite) an empty file (restart a file from scratch)
*/

// #include <stdlib.h>
import "C"

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"
)

const (
	tempPath string = "/home/johnierodgers/.commands/start/templates/%s.txt"
)

func main() {
	both := flag.Bool("b", false, "Create both a header and source file (C/C++)")
  overwrite := flag.Bool("w", false, "Overwrite existing file if it exists")
	atom := flag.Bool("a", false, "Open file in Atom")
	code := flag.Bool("c", false, "Open file in VSCode")
	open := flag.Bool("o", false, "Open file in default app")
	vim := flag.Bool("v", false, "Open file in Vim")

	flag.Parse()

	filepath := ""
	for _, arg := range flag.Args() {
		switch arg {
		case "-b":
			*both = true
    case "-w":
      *overwrite = true
		case "-a":
			*atom = true
		case "-c":
			*code = true
		case "-o":
			*open = true
		case "-v":
			*vim = true
		default:
			if arg[0] == '-' {
				fmt.Println("Invalid flag:", arg)
				os.Exit(1)
			}
			filepath = arg
		}
	}

	editor := ""
	if *atom {
		editor = "atom"
	} else if *code {
		editor = "code"
	} else if *open {
		editor = "open"
	} else if *vim {
		editor = "vim"
	}

  if !(*overwrite) {
    if _, err := os.Open(filepath); err == nil {
      fmt.Fprintln(os.Stderr, "File already exists")
      if editor != "" {
        openEditor(editor, filepath)
      }
      return
    }
  }

	ext := path.Ext(filepath)
	name := path.Base(filepath)
	if ext == "" || len(ext) == len(name) {
		_, err := os.Create(filepath)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		os.Exit(0)
	}
	name = name[:len(name) - len(ext)]
	hidden := (name[0] == '.')
	if hidden {
		name = name[1:]
	}

	filetype, replace := "", ""
	switch ext {
	case ".c":
		filetype = "c"
	case ".cc", ".cpp", ".c++":
		filetype = "cc"
	case ".f", ".f77", ".f90", ".f95":
		filetype = "f"
		replace = strings.ToLower(name)
	case ".go":
		filetype = "go"
	case ".h", ".hpp", ".h++":
		filetype = "h"
		replace = strings.ReplaceAll(strings.ToUpper(name + ext), ".", "_")
	case ".htm", ".html":
		filetype = "htm"
		replace = strings.ReplaceAll(strings.Title(name), "_", " ")
	case ".jav", ".java":
		filetype = "jav"
		replace = name
  case ".proto":
    filetype = "proto"
  case ".pl":
    filetype = "pl"
	case ".py":
		filetype = "py"
	case ".rs":
		filetype = "rs"
	default:
		_, err := os.Create(filepath)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
    openEditor(editor, filepath)
		os.Exit(0)
	}

	temp, err := template.ParseFiles(fmt.Sprintf(tempPath, filetype))
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	file, err := os.Create(filepath)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	if temp.Execute(file, replace); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
  if filetype == "py" || filetype == "pl" {
    cmd := C.CString("chmod a+x " + filepath)
		C.system(cmd)
  }
	file.Close()

	if *both {
		if filetype == "h" {
			if hidden {
				name = "." + name
			}
			ext = strings.ReplaceAll(ext, "h", "c")
			file, err := os.Create(name + ext)
			if err != nil {
				fmt.Println(err)
				os.Exit(2)
			}
			fmt.Fprintf(file, "#include \"%s\"\n", name + ext)
			file.Close()
		} else if filetype == "c" || filetype == "cc" {
			if ext == "cc" {
				ext = "hpp"
			} else {
				ext = strings.ReplaceAll(ext, "c", "h")
			}
			replace := strings.ToUpper(name + "_" + ext[1:])
			if hidden {
				name = "." + name
			}
			temp, err := template.ParseFiles(fmt.Sprintf(tempPath, "h"))
			if err != nil {
				fmt.Println(err)
				os.Exit(2)
			}
			file, err := os.Create(name + ext)
			if err != nil {
				fmt.Println(err)
				os.Exit(2)
			}
			if err := temp.Execute(file, replace); err != nil {
				fmt.Println(err)
				os.Exit(2)
			}
			file.Close()
		}
	}

  openEditor(editor, filepath)
}

func openEditor(editor, filepath string) {
  if editor != "" {
    cmd := C.CString(editor + " " + filepath)
    C.system(cmd)
  }
}
