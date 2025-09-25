package main

import (
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

	"github.com/spf13/cobra"
)

var tempPath string

func init() {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("error getting template directory")
	}
	tempPath = path.Join(path.Dir(thisFile), "templates", "%s.txt")
}

func main() {
	log.SetFlags(0)
	boolFlags := make(map[string]*bool, 8)

	rootCmd := &cobra.Command{
		Use:                   "start [FLAGS] <FILES...>",
		Short:                 "Start a file(s) based on a template",
		Long:                  "Start a file(s) based on a template for a given language. Accepted filetypes can be printed using the --filetypes flag.",
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			if b, _ := cmd.Flags().GetBool("filetypes"); b {
				fmt.Println(strings.Join(getFiletypeHelp(), "\n"))
				return
			}

			editor, _ := cmd.Flags().GetString("editor")
			filepaths := args

			// Set editor argument/variable
			if editor == "" {
				if *boolFlags["-a"] {
					editor = "atom"
				} else if *boolFlags["-c"] {
					editor = "code"
				} else if *boolFlags["-n"] {
					editor = "nano"
				} else if *boolFlags["-o"] {
					editor = "open"
				} else if *boolFlags["-v"] {
					editor = "vim"
				} else if *boolFlags["-V"] {
					editor = "nvim"
				} else if *boolFlags["-E"] {
					editor = "emacs"
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
		},
	}
	flags := rootCmd.Flags()
	flags.Bool("filetypes", false, "Print filetype mappings")
	boolFlags["-b"] = flags.BoolP(
		"both", "b", false,
		"Create both a header and source file (C/C++)",
	)
	boolFlags["-w"] = flags.BoolP("overwrite", "w", false, "Overwrite existing file if it exsts")
	boolFlags["-a"] = flags.BoolP("atom", "a", false, "Open file with Atom")
	boolFlags["-c"] = flags.BoolP("vsc", "c", false, "Open file with VSCode")
	boolFlags["-n"] = flags.BoolP("nano", "n", false, "Open file with Nano")
	boolFlags["-o"] = flags.BoolP("open", "o", false, "Open file with default app")
	boolFlags["-v"] = flags.BoolP("vim", "v", false, "Open file with Vim")
	boolFlags["-V"] = flags.BoolP("neovim", "V", false, "Open file with Neovim")
	boolFlags["-E"] = flags.BoolP("emacs", "E", false, "Open file with Emacs")
	flags.StringP("editor", "e", "", "Editor to open file with")
	boolFlags["-r"] = flags.BoolP(
		"empty", "r", false,
		"Start empty file, clearing old one if it exists",
	)
	rootCmd.MarkFlagsMutuallyExclusive(
		"open", "atom", "vsc",
		"nano", "vim", "neovim", "emacs",
		"editor",
	)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func createFile(filepath string, boolFlags map[string]*bool) error {
	if *boolFlags["-r"] {
		// Create the new empty file regargless of whether it exists or not
		_, err := os.Create(filepath)
		return err
	} else if !*(boolFlags["-w"]) {
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
	hidden := name[0] == '.'
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
	case "erl":
		replace = name
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
	if err := temp.Execute(file, replace); err != nil {
		return err
	}
	// Make the file executable if it's a script
	if scriptExts[filetype] {
		if err := exec.Command("chmod", "u+x", filepath).Run(); err != nil {
			log.Println(err)
		}
	}

	// Create the file for the h or c/cc file if the -b (both) flag is passed
	if *boolFlags["-b"] {
		if filetype == "h" {
			if hidden {
				name = "." + name
			}
			newExt := strings.ReplaceAll(ext, "h", "c")
			otherFile, err := os.Create(name + newExt)
			if err != nil {
				return err
			}
			defer otherFile.Close()
			_, err = fmt.Fprintf(otherFile, "#include \"%s\"\n", name+ext)
			if err != nil {
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
			if err := temp.Execute(otherFile, replace); err != nil {
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

var exts = map[string]string{
	".c":  "c",
	".cc": "cc", ".cpp": "cc", ".c++": "cc", ".cxx": "cc",
	".erl": "erl",
	".f":   "f", ".f77": "f", ".f90": "f", ".f95": "f",
	".go": "go",
	".h":  "h", ".hpp": "h", ".h++": "h", ".hxx": "h",
	".htm": "htm", ".html": "htm",
	".jav": "jav", ".java": "jav",
	".proto": "proto",
	".pl":    "pl",
	".py":    "py",
	".rs":    "rs",
	".sh":    "sh",
}

var scriptExts = map[string]bool{
	".pl": true, ".py": true, ".sh": true,
}

func getFiletypeHelp() []string {
	filetypes, filetypeArr := make(map[string][]string), make([]string, 0)
	for ext, filetype := range exts {
		if _, ok := filetypes[filetype]; !ok {
			filetypeArr = append(filetypeArr, filetype)
		}
		filetypes[filetype] = append(filetypes[filetype], ext)
	}
	sort.Strings(filetypeArr)
	res := []string{}
	for _, filetype := range filetypeArr {
		sort.Strings(filetypes[filetype])
		lang := filetypeToLang(filetype)
		res = append(
			res,
			fmt.Sprintf("%s: %s", lang, strings.Join(filetypes[filetype], ", ")),
		)
	}
	return res
}

func filetypeToLang(ft string) string {
	switch ft {
	case "c":
		return "C"
	case "cc":
		return "C++"
	case "erl":
		return "Erlang"
	case "f":
		return "Fortran"
	case "go":
		return "Go"
	case "h":
		return "C/C++ Header File"
	case "htm":
		return "HTML"
	case "jav":
		return "Java"
	case "proto":
		return "Protobuf"
	case "pl":
		return "Perl"
	case "py":
		return "Python"
	case "rs":
		return "Rust"
	case "sh":
		return "Shell"
	default:
		panic("unknown filetype: " + ft)
	}
}

func usingWSL() bool {
	v := os.Getenv("USING_WSL")
	return v == "1" || v == "on"
}
