package main
<<<<<<< HEAD
=======

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
>>>>>>> b9510c74774f6feee8a904cf441fe9e99ffb931c

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
<<<<<<< HEAD
	"sort"
=======
	"runtime"
>>>>>>> b9510c74774f6feee8a904cf441fe9e99ffb931c
	"strings"
	"syscall"
	"text/template"
)

<<<<<<< HEAD
const tempPath string = "/home/johnierodgers/.commands/start/templates/%s.txt"

var exts = map[string]string{
	".c":  "c",
	".cc": "cc", ".cpp": "cc", ".c++": "cc",
	".f": "f", ".f77": "f", ".f90": "f", ".f95": "f",
	".go": "go",
	".h":  "h", ".hpp": "h", ".h++": "h",
	".htm": "htm", ".html": "htm",
	".jav": "jav", ".java": "jav",
	".proto": "proto",
	".pl":    "pl",
	".py":    "py",
	".rs":    "rs",
	".sh":    "sh",
}

var scriptFiles = map[string]bool{
	".pl": true, ".py": true, ".sh": true,
}
=======
var (
	tempPath string
)
>>>>>>> b9510c74774f6feee8a904cf441fe9e99ffb931c

func init() {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Fprintln(os.Stderr, "error getting template directory")
		os.Exit(1)
	}
	tempPath = path.Join(path.Dir(thisFile), "templates", "%s.txt")
}

func main() {
<<<<<<< HEAD
	log.SetFlags(0)
=======
	both := flag.Bool("b", false, "Create both a header and source file (C/C++)")
	overwrite := flag.Bool("w", false, "Overwrite existing file if it exists")
	atom := flag.Bool("a", false, "Open file in Atom")
	code := flag.Bool("c", false, "Open file in VSCode")
	open := flag.Bool("o", false, "Open file in default app")
	vim := flag.Bool("v", false, "Open file in Vim")
>>>>>>> b9510c74774f6feee8a904cf441fe9e99ffb931c

	// Create the flags
	boolFlags := make(map[string]*bool, 8)
	boolFlags["-b"] = flag.Bool("b", false, "Create both a header and source file (C/C++)")
	boolFlags["-w"] = flag.Bool("w", false, "Overwrite existing file if it exists")
	boolFlags["-a"] = flag.Bool("a", false, "Open file with Atom")
	boolFlags["-c"] = flag.Bool("c", false, "Open file with VSCode")
	boolFlags["-n"] = flag.Bool("n", false, "Open file with Nano")
	boolFlags["-o"] = flag.Bool("o", false, "Open file with default app")
	boolFlags["-v"] = flag.Bool("v", false, "Open file in Vim")
	boolFlags["-r"] = flag.Bool("r", false, "Start empty file, clearing old one if it exists")
	editorPtr := flag.String("e", "", "Editor to open file with")
	flag.Usage = usageFunc
	flag.Parse()

	// Get the files and any flags not picked up after parsing
	filepaths, prev, editor := []string{}, "", *editorPtr
	for _, arg := range flag.Args() {
<<<<<<< HEAD
		if f := boolFlags[arg]; f != nil {
			*f = true
		} else {
			if arg[0] == '-' && arg != "-e" {
				log.Fatalf("invalid flag: %s", arg)
			} else if prev == "-e" {
				if editor == "" {
					editor = arg
				}
			} else {
				filepaths = append(filepaths, arg)
=======
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
>>>>>>> b9510c74774f6feee8a904cf441fe9e99ffb931c
			}
		}
		prev = arg
	}

	// Set editor argument/variable
	if editor == "" {
		if *(boolFlags["-a"]) {
			editor = "atom"
		} else if *(boolFlags["-c"]) {
			editor = "code"
		} else if *(boolFlags["-n"]) {
			editor = "nano"
		} else if *(boolFlags["-o"]) {
			editor = "open"
		} else if *(boolFlags["-v"]) {
			editor = "vim"
		}
	}

<<<<<<< HEAD
	// Create the file(s)
	if len(filepaths) == 1 {
		if err := createFile(filepaths[0], boolFlags); err != nil {
			log.Println(err)
      if !strings.HasSuffix(err.Error(), "already exists") {
        return
      }
		}
		openEditor(editor, filepaths[0])
	} else if len(filepaths) == 0 {
		log.Fatal("must specify file(s)")
	}
	for _, filepath := range filepaths {
		if err := createFile(filepath, boolFlags); err != nil {
			log.Println(err)
		}
	}
}
=======
	if !(*overwrite) {
		if _, err := os.Open(filepath); err == nil {
			fmt.Fprintln(os.Stderr, "File already exists")
			if editor != "" {
				openEditor(editor, filepath)
			}
			return
		}
	}
>>>>>>> b9510c74774f6feee8a904cf441fe9e99ffb931c

func createFile(filepath string, boolFlags map[string]*bool) error {
	if *(boolFlags["-r"]) {
		// Create the new empty file regargless of whether it exists or not
		_, err := os.Create(filepath)
		return err
	} else if !(*(boolFlags["-w"])) {
		// Return an error if the file exists and isn't meant to be overwritten
		if _, err := os.Stat(filepath); err == nil {
			return fmt.Errorf("%s already exists", filepath)
		}
	}
<<<<<<< HEAD

	// Get the info of the file name
	ext := path.Ext(filepath)
	name := strings.TrimSuffix(path.Base(filepath), ext)
	if ext == "" || name == "" {
		_, err := os.Create(filepath)
		return err
	}
=======
	name = name[:len(name)-len(ext)]
>>>>>>> b9510c74774f6feee8a904cf441fe9e99ffb931c
	hidden := (name[0] == '.')
	if hidden {
		name = name[1:]
	}

	// Get the filetype
	// "replace" is used to replace parts of the template if necessary
	var filetype, replace string
	if filetype = exts[ext]; filetype == "" {
		// If there is no template, just create the file
		_, err := os.Create(filepath)
		return err
	}
	// Set the replace variable
	switch filetype {
	case "f":
		replace = strings.ToLower(name)
<<<<<<< HEAD
	case "h":
		replace = strings.ReplaceAll(strings.ToUpper(name+ext), ".", "_")
	case "htm":
=======
	case ".go":
		filetype = "go"
	case ".h", ".hpp", ".h++":
		filetype = "h"
		replace = strings.ReplaceAll(strings.ToUpper(name+ext), ".", "_")
	case ".htm", ".html":
		filetype = "htm"
>>>>>>> b9510c74774f6feee8a904cf441fe9e99ffb931c
		replace = strings.ReplaceAll(strings.Title(name), "_", " ")
	case "jav":
		replace = name
<<<<<<< HEAD
=======
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
>>>>>>> b9510c74774f6feee8a904cf441fe9e99ffb931c
	}

	// Parse the template and create the file
	temp, err := template.ParseFiles(fmt.Sprintf(tempPath, filetype))
	if err != nil {
		return err
	}
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	if temp.Execute(file, replace); err != nil {
		return err
	}
	// Make the file executable if it's a script
	if scriptFiles[filetype] {
		if err := exec.Command("chmod", "u+x", filepath).Run(); err != nil {
			log.Println(err)
		}
	}
<<<<<<< HEAD
=======
	if filetype == "py" || filetype == "pl" {
		cmd := C.CString("chmod a+x " + filepath)
		C.system(cmd)
	}
	file.Close()
>>>>>>> b9510c74774f6feee8a904cf441fe9e99ffb931c

	// Create the file for the h or c/cc file if the -b (both) flag is passed
	if *(boolFlags["-b"]) {
		if filetype == "h" {
			if hidden {
				name = "." + name
			}
			ext = strings.ReplaceAll(ext, "c", "h")
			otherFile, err := os.Create(name + ext)
			if err != nil {
				return err
			}
			defer otherFile.Close()
			if _, err := fmt.Fprintf(file, "#include \"%s\"\n", name+ext); err != nil {
				return err
			}
<<<<<<< HEAD
=======
			fmt.Fprintf(file, "#include \"%s\"\n", name+ext)
			file.Close()
>>>>>>> b9510c74774f6feee8a904cf441fe9e99ffb931c
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
				return err
			}
			otherFile, err := os.Create(name + ext)
			if err != nil {
				return err
			}
			defer otherFile.Close()
			if err := temp.Execute(file, replace); err != nil {
				return err
			}
		}
	}
<<<<<<< HEAD
	return nil
=======

	openEditor(editor, filepath)
>>>>>>> b9510c74774f6feee8a904cf441fe9e99ffb931c
}

func openEditor(editor, filepath string) {
	if editor != "" {
<<<<<<< HEAD
		binPath, err := exec.LookPath(editor)
		if err != nil {
			log.Fatal(err)
		}
		err = syscall.Exec(binPath, []string{editor, filepath}, os.Environ())
		if err != nil {
			log.Fatal(err)
		}
	}
}

func usageFunc() {
	clof := func(format string, args ...interface{}) {
		fmt.Fprintf(flag.CommandLine.Output(), format, args...)
	}
	clof("Usage of %s [files] [options]:\n", os.Args[0])
	flag.PrintDefaults()
	// Print out the template names and extention mappings
	clof("File extensions with templates:")
	filetypes, filetypeArr := make(map[string][]string), make([]string, 1)
	for ext, filetype := range exts {
		if _, ok := filetypes[filetype]; !ok {
			filetypeArr = append(filetypeArr, filetype)
		}
		filetypes[filetype] = append(filetypes[filetype], ext)
	}
	sort.Strings(filetypeArr)
	for _, filetype := range filetypeArr {
		sort.Strings(filetypes[filetype])
		clof("  %s\t%s\n", filetype, strings.Join(filetypes[filetype], ", "))
=======
		cmd := C.CString(editor + " " + filepath)
		C.system(cmd)
>>>>>>> b9510c74774f6feee8a904cf441fe9e99ffb931c
	}
}
