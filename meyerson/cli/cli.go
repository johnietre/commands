package cli

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func Run(args []string) {
	log.SetFlags(0)

	fs := flag.NewFlagSet("meyerson cli", flag.ExitOnError)
	outDir := fs.String("out-dir", ".", "Directory to put output files in")
	configPath := fs.String("config", "", "Path to config file")
	configTemp := fs.Bool(
		"config-template", false,
		"Generate a template configuration in the current directory",
	)
	fs.Parse(args)

	if *configTemp {
		_, thisFile, _, _ := runtime.Caller(0)
		configPath := filepath.Join(filepath.Dir(thisFile), "meyerson.json")
		data, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatal("error reading config template: ", err)
		}
		if err = os.WriteFile("meyerson.json", data, 0666); err != nil {
			log.Fatal("error writing config template: ", err)
		}
		return
	}

	if *configPath != "" {
		f, err := os.Open(*configPath)
		if err != nil {
			log.Fatal("error opening config file: ", err)
		}
		config := &Config{}
		if err = json.NewDecoder(f).Decode(config); err != nil {
			f.Close()
			log.Fatal("error parsing config file: ", err)
		}
		app := AppFromConfig(config)
		if *outDir != "" {
			app.outDir = *outDir
		}
		fmt.Println("Starting processes...")
		app.StartProcs()
		app.Wait()
		return
	}

	app := &App{outDir: *outDir}

	// Create the processes
	for i := 1; true; {
		name := readline(fmt.Sprintf("Process %d Name: ", i))
		if name == "" {
			break
		}
		prog := readline("Program: ")

		var args []string
		for i := 1; true; i++ {
			if arg := readline(fmt.Sprintf("Arg %d: ", i)); arg != "" {
				args = append(args, arg)
			} else {
				break
			}
		}

		env := os.Environ()
		for {
			if kv := readline(fmt.Sprintf("Env Var (key=val): ")); kv != "" {
				env = append(env, kv)
			} else {
				break
			}
		}

		outFilename := readline(
			"Stdout output filename (- = process number, % = name): ",
		)
		errFilename := readline(
			"Stderr output filename (- = process number, % = name): ",
		)

		conf := strings.ToLower(readline("Ok [Y/n]? "))
		if conf != "y" && conf != "yes" {
			continue
		}

		cmd := exec.Command(prog, args...)
		cmd.Env = env
		proc := &Process{
			app:         app,
			Name:        name,
			num:         i,
			cmd:         cmd,
			OutFilename: outFilename,
			ErrFilename: errFilename,
		}
		app.AddProc(proc)

		conf = strings.ToLower(readline("Start now [Y/n]? "))
		if conf == "y" || conf == "yes" {
			fmt.Printf("Starting process %d (%s)\n", proc.num, proc.Name)
			if err := proc.Start(); err != nil {
				fmt.Printf(
					"error starting process %d (%s): %v\n", proc.num, proc.Name, err,
				)
			}
		}

		fmt.Println("====================")
		i++
	}
	fmt.Println("========================================")

	// Check for any deletions
	for {
		if name := readline("Delete any procs (enter name)? "); name == "" {
			break
		} else if !app.RemoveProc(name) {
			fmt.Println("No process with name: ", name)
		}
	}
	fmt.Println("========================================")

	// Start the processes
	fmt.Println("Starting (remaining) processes...")
	app.StartProcs()
	app.Wait()
}

type Config struct {
	OutDir string     `json:"outDir"`
	Env    []string   `json:"env"`
	Procs  []*Process `json:"procs"`
}

type App struct {
	procs  []*Process
	outDir string
	wg     sync.WaitGroup
}

func AppFromConfig(config *Config) *App {
	app := &App{
		procs:  config.Procs,
		outDir: config.OutDir,
	}
	env := append(os.Environ(), config.Env...)
	for i, proc := range app.procs {
		proc.num = i + 1
		proc.app = app
		cmd := exec.Command(proc.Program, proc.Args...)
		cmd.Env = make([]string, len(env), len(env)+len(proc.Env))
		copy(cmd.Env, env)
		cmd.Env = append(cmd.Env, proc.Env...)
		proc.cmd = cmd
	}
	return app
}

func (a *App) AddProc(p *Process) {
	p.num = len(a.procs) + 1
	a.procs = append(a.procs, p)
}

// Returns true if a process was deleted
func (a *App) RemoveProc(name string) bool {
	for i, proc := range a.procs {
		if proc.Name == name {
			a.procs = append(a.procs[:i], a.procs[i+1:]...)
			return true
		}
	}
	return false
}

func (a *App) StartProcs() {
	for _, proc := range a.procs {
		if proc.Delay != 0 {
			time.Sleep(time.Second * proc.Delay)
		}
		if err := proc.Start(); err != nil {
			fmt.Printf(
				"error starting process %d (%s): %v\n", proc.num, proc.Name, err,
			)
		}
	}
}

func (a *App) Wait() {
	a.wg.Wait()
}

type Process struct {
	Name        string        `json:"name"`
	Program     string        `json:"program"`
	Args        []string      `json:"args"`
	Env         []string      `json:"env"`
	OutFilename string        `json:"outFilename"`
	ErrFilename string        `json:"errFilename"`
	Delay       time.Duration `json:"delay"`

	app              *App
	num              int
	cmd              *exec.Cmd
	outFile, errFile *os.File
	status           atomic.Uint32
}

func (p *Process) Start() error {
	if p.status.Load() != statusNotStarted {
		return nil
	}
	var err error
	// Open the files for output
	if p.OutFilename != "" {
		if p.OutFilename == "-" {
			p.OutFilename = fmt.Sprintf("process%d-stdout.txt", p.num)
		} else if p.OutFilename == "%" {
			p.OutFilename = fmt.Sprintf("%s-stdout.txt", p.Name)
		}
		p.outFile, err = os.Create(filepath.Join(p.app.outDir, p.OutFilename))
		if err != nil {
			fmt.Printf(
				"Error creating stdout output file for %s: %v\n",
				p.Name, err,
			)
		}
		p.cmd.Stdout = p.outFile
	}
	if p.ErrFilename != "" {
		if p.ErrFilename == "-" {
			p.ErrFilename = fmt.Sprintf("process%d-stderr.txt", p.num)
		} else if p.ErrFilename == "%" {
			p.ErrFilename = fmt.Sprintf("%s-stderr.txt", p.Name)
		} else if p.ErrFilename == p.OutFilename {
			// Same file
			p.ErrFilename = p.OutFilename
			goto StartProc
		}
		p.errFile, err = os.Create(filepath.Join(p.app.outDir, p.ErrFilename))
		if err != nil {
			fmt.Printf(
				"Error creating stderr output file for %s: %v\n",
				p.Name, err,
			)
		}
		p.cmd.Stderr = p.errFile
	}
StartProc:
	// Start the process
	if err := p.cmd.Start(); err != nil {
		p.status.Store(statusFinished)
		fmt.Printf("Error starting %s: %v\n", p.Name, err)
		// Delete the created files
		if p.outFile != nil {
			p.outFile.Close()
			if err := os.Remove(p.outFile.Name()); err != nil {
				fmt.Printf("Error removing stdout file for %s: %v\n", p.Name, err)
			}
		}
		if p.errFile != nil {
			p.errFile.Close()
			if err := os.Remove(p.errFile.Name()); err != nil {
				fmt.Printf("Error removing stderr file for %s: %v\n", p.Name, err)
			}
		}
		return err
	}
	p.status.Store(statusRunning)
	p.app.wg.Add(1)
	// Wait for the process to finish
	go func() {
		p.Wait()
	}()
	return nil
}

func (p *Process) Wait() {
	err := p.cmd.Wait()
	p.status.Store(statusFinished)
	if err != nil {
		fmt.Printf("%s terminated with error: %v\n", p.Name, err)
	} else {
		fmt.Println(p.Name, "finished")
	}
	// Close the files
	if p.outFile != nil {
		p.outFile.Close()
	}
	if p.errFile != nil {
		p.errFile.Close()
	}
	p.app.wg.Done()
}

const (
	statusNotStarted uint32 = iota
	statusRunning
	statusFinished
)

var stdinReader = bufio.NewReader(os.Stdin)

func readline(prompt ...string) string {
	if len(prompt) != 0 {
		fmt.Print(prompt[0])
	}
	line, err := stdinReader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(line)
}
