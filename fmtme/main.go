package main

// TODO: Config file with file extentions and programs

import (
	//"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var (
	quiet, walk                   bool
	extRegex, fileRegex, dirRegex bool
	includeFiles, excludeFiles    []*regexp.Regexp
	includeExts, excludeExts      []*regexp.Regexp
	includeDirs, excludeDirs      []*regexp.Regexp
)

func main() {
	log.SetFlags(0)

	rootCmd := &cobra.Command{
		Use:                   "fmtme [FLAGS] <FILES...>",
		Short:                 "Format various file types",
		Long:                  "Format various file types with a single command. Accepted filetypes (extensions) are the following:\n" + strings.Join(getSortedExts(), ", "),
		DisableFlagsInUseLine: true,
		Run:                   Run,
	}
	flags := rootCmd.Flags()
	flags.BoolVarP(
		&quiet, "quiet", "q", false,
		"Quiet (only display errors and output from commands)",
	)
	flags.BoolVarP(
		&walk, "walk", "r", false,
		"Walk directories (recusive search; only accepted filetypes are considered); defaults to starting in current directory if none are provided",
	)
	flags.StringSlice("iext", []string{}, "Extensions to include (all others are excluded)")
	flags.StringSlice("eext", []string{}, "Extensions to exclude (all others are included)")
	flags.StringSlice("ifile", []string{}, "File names to include (all others are excluded)")
	flags.StringSlice("efile", []string{}, "File names to exclude (all others are included)")
	flags.StringSlice("idir", []string{}, "Directories to include when walking (all others are excluded)")
	flags.StringSlice("edir", []string{}, "Directories to exclude when walking (all others are included)")
	flags.BoolVar(
		&extRegex, "rext", false,
		"Use regex for ext filter; if multiple regexes are passed, an ext only need match one",
	)
	flags.BoolVar(
		&fileRegex, "rfile", false,
		"Use regex for file filter; if multiple regexes are passed, a file only need match one",
	)
	flags.BoolVar(
		&dirRegex, "rdir", false,
		"Use regex for dir filter; if multiple regexes are passed, a dir only need match one",
	)
	rootCmd.MarkFlagsMutuallyExclusive("iext", "eext")
	rootCmd.MarkFlagsMutuallyExclusive("ifile", "efile")
	rootCmd.MarkFlagsMutuallyExclusive("idir", "edir")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func Run(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	iexts, _ := flags.GetStringSlice("iext")
	for _, iext := range iexts {
		if len(iext) != 0 && iext[0] == '.' {
			iext = iext[1:]
		}
		rexpr, err := (*regexp.Regexp)(nil), error(nil)
		if !extRegex {
			rexpr, err = regexp.Compile(`^` + iext + `$`)
		} else {
			rexpr, err = regexp.Compile(iext)
		}
		if err != nil {
			log.Fatalf("error parsing regex %s: %v", iext, err)
		}
		includeExts = append(includeExts, rexpr)
	}

	eexts, _ := flags.GetStringSlice("eext")
	for _, eext := range eexts {
		if len(eext) != 0 && eext[0] == '.' {
			eext = eext[1:]
		}
		rexpr, err := (*regexp.Regexp)(nil), error(nil)
		if !extRegex {
			rexpr, err = regexp.Compile(`^` + eext + `$`)
		} else {
			rexpr, err = regexp.Compile(eext)
		}
		if err != nil {
			log.Fatalf("error parsing regex %s: %v", eext, err)
		}
		excludeExts = append(excludeExts, rexpr)
	}

	ifiles, _ := flags.GetStringSlice("ifile")
	for _, ifile := range ifiles {
		rexpr, err := (*regexp.Regexp)(nil), error(nil)
		if !fileRegex {
			rexpr, err = regexp.Compile(`^` + ifile + `$`)
		} else {
			rexpr, err = regexp.Compile(ifile)
		}
		if err != nil {
			log.Fatalf("error parsing regex %s: %v", ifile, err)
		}
		includeFiles = append(includeFiles, rexpr)
	}

	efiles, _ := flags.GetStringSlice("efile")
	for _, efile := range efiles {
		rexpr, err := (*regexp.Regexp)(nil), error(nil)
		if !fileRegex {
			rexpr, err = regexp.Compile(`^` + efile + `$`)
		} else {
			rexpr, err = regexp.Compile(efile)
		}
		if err != nil {
			log.Fatalf("error parsing regex %s: %v", efile, err)
		}
		excludeFiles = append(excludeFiles, rexpr)
	}

	idirs, _ := flags.GetStringSlice("idir")
	for _, idir := range idirs {
		rexpr, err := (*regexp.Regexp)(nil), error(nil)
		if !dirRegex {
			rexpr, err = regexp.Compile(`^` + idir + `$`)
		} else {
			rexpr, err = regexp.Compile(idir)
		}
		if err != nil {
			log.Fatalf("error parsing regex %s: %v", idir, err)
		}
		includeDirs = append(includeDirs, rexpr)
	}

	edirs, _ := flags.GetStringSlice("edir")
	for _, edir := range edirs {
		rexpr, err := (*regexp.Regexp)(nil), error(nil)
		if !dirRegex {
			rexpr, err = regexp.Compile(`^` + edir + `$`)
		} else {
			rexpr, err = regexp.Compile(edir)
		}
		if err != nil {
			log.Fatalf("error parsing regex %s: %v", edir, err)
		}
		excludeDirs = append(excludeDirs, rexpr)
	}

	if len(args) == 0 && walk {
		args = []string{"."}
	}

	runWithFiles(args)
}

var extFuncs = map[string]func([]string) *Cmd{
	"c":    clangFormat,
	"cc":   clangFormat,
	"cpp":  clangFormat,
	"cxx":  clangFormat,
	"go":   goFmt,
	"h":    clangFormat,
	"hpp":  clangFormat,
	"hxx":  clangFormat,
	"js":   clangFormat,
	"json": clangFormat,
	"rs":   rustFmt,
}

func runWithFiles(files []string) {
	checkInclude := true
	if walk {
		files = walkGetFiles(files)
		checkInclude = false
	}
	fileTypes := make(map[string][]string)
	for _, file := range files {
		ext := filepath.Ext(file)
		if ext != "" {
			ext = ext[1:]
		}
		if checkInclude && !includeFile(filepath.Base(file)) {
			continue
		}
		fileTypes[ext] = append(fileTypes[ext], file)
	}
	for ft, files := range fileTypes {
		cmdFunc := extFuncs[ft]
		if cmdFunc == nil {
			log.Printf("Unsupported filetype: %q (files: %v)", ft, files)
			continue
		}
		cmd := cmdFunc(files)
		if err := cmd.Run(); err != nil {
			log.Printf("Error running command for filetype %s: %v", ft, err)
		}
	}
}

func getSortedExts() []string {
	exts := make([]string, 0, len(extFuncs))
	for ext := range extFuncs {
		exts = append(exts, ext)
	}
	sort.StringSlice(exts).Sort()
	return exts
}

func includeFile(filename string) bool {
	ext := filepath.Ext(filename)
	name := filepath.Base(filename)
	name = name[:len(name)-len(ext)]
	if ext != "" {
		ext = ext[1:]
	}
	return includeFileExt(ext) && includeFileName(name)
}

func includeFileExt(ext string) bool {
	for _, rexpr := range includeExts {
		if rexpr.MatchString(ext) {
			return true
		}
	}
	for _, rexpr := range excludeExts {
		if rexpr.MatchString(ext) {
			return false
		}
	}
	return len(includeExts) == 0
}

func includeFileName(name string) bool {
	for _, rexpr := range includeFiles {
		if rexpr.MatchString(name) {
			return true
		}
	}
	for _, rexpr := range excludeFiles {
		if rexpr.MatchString(name) {
			return false
		}
	}
	return len(includeFiles) == 0
}

func includeDir(name string) bool {
	for _, rexpr := range includeDirs {
		if rexpr.MatchString(name) {
			return true
		}
	}
	for _, rexpr := range excludeDirs {
		if rexpr.MatchString(name) {
			return false
		}
	}
	return len(includeDirs) == 0
}

func walkGetFiles(files []string) []string {
	var res []string
	for _, path := range files {
		filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if includeDir(filepath.Base(path)) {
					return nil
				}
				return filepath.SkipDir
			}
			ext := filepath.Ext(path)
			if ext != "" {
				ext = ext[1:]
			}
			if extFuncs[ext] != nil && includeFile(path) {
				res = append(res, path)
			}
			return nil
		})
	}
	return res
}

func printlnFunc(args ...any) {
	if !quiet {
		fmt.Println(args...)
	}
}
