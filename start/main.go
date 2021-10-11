package main

/* TODO:
* Runing "start build.sh" creates the file but also outputs:
   "build.sh already exists"
*/

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"text/template"
)

var tempPath string

var exts = map[string]string{
	".c":  "c",
	".cc": "cc", ".cpp": "cc", ".c++": "cc",
	".erl": "erl",
	".f":   "f", ".f77": "f", ".f90": "f", ".f95": "f",
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

func init() {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("error getting template directory")
	}
	tempPath = path.Join(path.Dir(thisFile), "templates", "%s.txt")
}

func main() {
	log.SetFlags(0)

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
	} else {
		for _, filepath := range filepaths {
			if err := createFile(filepath, boolFlags); err != nil {
				log.Println(err)
			}
		}
	}
}

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

	// Get the info of the file name
	ext := path.Ext(filepath)
	name := strings.TrimSuffix(path.Base(filepath), ext)
	if ext == "" || name == "" {
		_, err := os.Create(filepath)
		return err
	}
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
	case "h":
		replace = strings.ReplaceAll(strings.ToUpper(name+ext), ".", "_")
	case "htm":
		replace = strings.ReplaceAll(strings.Title(name), "_", " ")
	case "jav":
		replace = name
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
	return nil
}

func openEditor(editor, filepath string) {
	if editor == "atom" && usingWSL() {
		binPath, err := exec.LookPath("cmd.exe")
		if err != nil {
			log.Fatal(err)
		}
		err = syscall.Exec(binPath, []string{"cmd.exe", "/C", "atom", filepath}, os.Environ())
		if err != nil {
			log.Fatal(err)
		}
	} else if editor != "" {
		binPath, err := exec.LookPath(editor)
		if err != nil {
			log.Fatal(err)
		}
		err = syscall.Exec(binPath, []string{editor, filepath}, os.Environ())
		if err != nil {
			log.Fatal(err)
		}
	}
	os.Exit(0)
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
	}
}

func usingWSL() bool {
	v := os.Getenv("USING_WSL")
	return v == "1" || v == "on"
}
