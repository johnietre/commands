package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	utils "github.com/johnietre/utils/go"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	cmdsDir, binDir string

	installProg       = "sh"
	installFileName   = "install.sh"
	descFileName      = "description"
	uninstallFileName = "uninstall.sh"

	names, paths []string = nil, nil

	width     int
	seperator string
)

func init() {
	log.SetFlags(0)

	_, thisFile, _, _ := runtime.Caller(0)
	cmdsDir = filepath.Dir(thisFile)
	binDir = filepath.Join(cmdsDir, "bin")

	var err error
	width, _, err = term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80
		log.Printf("Error getting terminal width, using %d\n%v", width, err)
	}
	seperator = strings.Repeat("=", width)
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "jtcmds",
		Short: "Manage commands.",
		Long:  "The manager to install, uninstall, etc. various commands.",
		PreRun: func(cmd *cobra.Command, args []string) {
		},
		Run: func(cmds *cobra.Command, args []string) {},
	}
	rootCmd.AddCommand(
		makeListCmd(),
		makeDescCmd(),
		makeInstallCmd(),
		makeUninstallCmd(),
	)
	getNames()
	//cobra.CheckErr(rootCmd.Execute())
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}

func makeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list [all/installed/uninstalled]",
		Aliases: []string{"l"},
		Short:   "List programs.",
		Long:    "List various categories of programs.",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch arg := cmd.Flags().Arg(0); arg {
			case "", "all":
				for _, name := range names {
					fmt.Println(name)
				}
			case "installed":
				ents, err := os.ReadDir(binDir)
				if err != nil {
					log.Fatal("Error reading bin directory: ", err)
				}
				for _, ent := range ents {
					fmt.Println(ent.Name())
				}
			case "uninstalled":
				ents, err := os.ReadDir(binDir)
				if err != nil {
					log.Fatal("Error reading bin directory: ", err)
				}
				var uninstalled []string
				if l := len(ents); l != 0 {
					uninstalled = utils.FilterSlice(names, func(name string) bool {
						return -1 != sort.Search(len(ents), func(i int) bool {
							return ents[i].Name() == name
						})
					})
				} else {
					uninstalled = names
				}
				for _, name := range uninstalled {
					fmt.Println(name)
				}
			default:
				log.Fatal("Unknown argument: ", arg)
			}
		},
	}
}

func makeDescCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "describe <PROGRAMS...>",
		Aliases: []string{"desc"},
		Short:   "Describe program(s).",
		Long:    `Print descriptions of various programs. Use "all" as the only argument to print all descriptions.`,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 1 && args[0] == "all" {
				args = names
			}
			loopThru(args, func(path string) {
				desc := getDesc(path)
				if desc == nil {
					desc = []byte("no description found")
				}
				fmt.Println(getName(path))
				os.Stdout.Write(desc)
				fmt.Println()
			})
		},
	}
}

func makeInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "install <PROGRAMS...>",
		Aliases: []string{"i"},
		Short:   "Install programs.",
		Long:    `Install any number of programs. Use "all" as the only argument to (re)install all programs.`,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 1 && args[0] == "all" {
				args = names
			}
			loopThru(args, func(path string) {
				fmt.Println(getName(path))
				// No need to check install
				installPath := filepath.Join(path, installFileName)
				if err := newCmd("sh", installPath).Run(); err != nil {
					fmt.Printf("Error running install script for %s: %v\n", path, err)
					return
				}
				fmt.Println("INSTALLED")
			})
		},
	}
}

func makeUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "uninstall <PROGRAMS...>",
		Aliases: []string{"u"},
		Short:   "Uninstall programs.",
		Long:    `Uninstall any number of programs. Use "all" as the only argument to uninstall all programs.`,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 1 && args[0] == "all" {
				args = names
			}
			binDir := filepath.Join(cmdsDir, "bin")
			loopThru(args, func(path string) {
				name := getName(path)
				if name == "" {
					log.Println("Missing name for path: ", path)
					return
				}
				fmt.Println(getName(path))
				uninstallPath := filepath.Join(path, uninstallFileName)
				if _, err := os.Lstat(uninstallPath); err != nil {
					if !os.IsNotExist(err) {
						fmt.Println("Error:", err)
						return
					}
					fmt.Println("no uninstall script, just deleting binary...")
					if err := os.Remove(filepath.Join(binDir, name)); err != nil {
						fmt.Printf("Error removing %s binary: %v\n", name, err)
						return
					}
				} else {
					if err := newCmd("sh", uninstallPath).Run(); err != nil {
						fmt.Printf("Error running uninstall script for %s: %v\n", path, err)
						return
					}
				}
				fmt.Println("UNINSTALLED")
			})
		},
	}
}

func getDesc(dirPath string) []byte {
	descPath := filepath.Join(dirPath, descFileName)
	bytes, err := os.ReadFile(descPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		log.Fatal("Error reading descript file (%s): %v", descPath, err)
	}
	return bytes
}

func getNames() {
	if names != nil {
		return
	}

	// Make sure bin directory exists
	if _, err := os.Lstat(binDir); err != nil {
		if !os.IsNotExist(err) {
			log.Fatal("Error creating bin directory: ", err)
		}
		if err := os.Mkdir(binDir, 0755); err != nil {
			log.Fatal("Error creating bin directory: ", err)
		}
	}

	ents, err := os.ReadDir(cmdsDir)
	if err != nil {
		log.Fatal("Error getting commands: ", err)
	}
	names = make([]string, 0, len(ents))
	paths = make([]string, 0, len(ents))
	for _, ent := range ents {
		if !ent.IsDir() {
			continue
		}
		path := filepath.Join(cmdsDir, ent.Name())

		installPath := filepath.Join(path, installFileName)
		if _, err := os.Lstat(installPath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			log.Fatalf("Error reading dir (%s): %v", installPath, err)
		}
		names = append(names, ent.Name())
		paths = append(paths, path)
	}
}

func getPath(name string) string {
	for i, n := range names {
		if n == name {
			return paths[i]
		}
	}
	return ""
}

func getName(path string) string {
	for i, p := range paths {
		if p == path {
			return names[i]
		}
	}
	return ""
}

func loopThru(args []string, f func(path string)) {
	for _, name := range args {
		path := getPath(name)
		if path == "" {
			fmt.Println("no program", name)
			fmt.Println(seperator)
			continue
		}
		f(path)
		fmt.Println(seperator)
	}
}

func newCmd(prog string, args ...string) *exec.Cmd {
	cmd := exec.Command(prog, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	//cmd.Stdin, cmd.Stderr = os.Stdin, os.Stderr
	cmd.Dir = cmdsDir
	return cmd
}
