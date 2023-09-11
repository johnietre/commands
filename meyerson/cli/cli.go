package cli

// TODO: Allow tasks to be killed/restart
// TODO: All environment variables to be read from file (from config and CLI)

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
  "io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
  "strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
  app *App
  errProcRunning = fmt.Errorf("process running already")
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
		app = AppFromConfig(config)
		if *outDir != "" {
			app.outDir = *outDir
		}
		Println("Starting processes...")
		app.StartProcs()
    handleInput()
		app.Wait()
		return
	}

	app = &App{outDir: *outDir}

	// Create the processes
	for i := 1; true; {
    proc := &Process{app: app, num: i}
		proc.Name = readline(fmt.Sprintf("Process %d Name: ", i))
		if proc.Name == "" {
			break
		}
		proc.Program = readline("Program: ")

		for i := 1; true; i++ {
			if arg := readline(fmt.Sprintf("Arg %d: ", i)); arg != "" {
				proc.Args = append(proc.Args, arg)
			} else {
				break
			}
		}

		proc.Env = os.Environ()
		for {
			if kv := readline(fmt.Sprintf("Env Var (key=val): ")); kv != "" {
				proc.Env = append(proc.Env, kv)
			} else {
				break
			}
		}

		proc.OutFilename = readline(
			"Stdout output filename (- = process number, % = name): ",
		)
		proc.ErrFilename = readline(
			"Stderr output filename (- = process number, % = name): ",
		)

		conf := strings.ToLower(readline("Ok [Y/n]? "))
		if conf != "y" && conf != "yes" {
			continue
		}

		app.AddProc(proc)

		conf = strings.ToLower(readline("Start now [Y/n]? "))
		if conf == "y" || conf == "yes" {
			Printf("Starting process %d (%s)\n", proc.num, proc.Name)
			if err := proc.Start(); err != nil {
				Printf(
					"error starting process %d (%s): %v\n", proc.num, proc.Name, err,
				)
			}
		}

		Println("====================")
		i++
	}
	Println("========================================")

	// Check for any deletions
	for {
		if name := readline("Delete any procs (enter name)? "); name == "" {
			break
		} else if !app.RemoveProc(name) {
			Println("No process with name: ", name)
		}
	}
	Println("========================================")

	// Start the processes
	Println("Starting (remaining) processes...")
	app.StartProcs()
  handleInput()
	app.Wait()
}

func handleInput() {
  Println("Pause output (p) to enter commands | Ctrl-C to quit")

  // Wait for input
  printChoices := func() {
    fmt.Println("Options")
    fmt.Println("1) Print Options (Print This)")
    fmt.Println("2) Print Processes")
    fmt.Println("3) Restart Process (Kill)")
    fmt.Println("4) Restart Process (Interrupt)")
    fmt.Println("5) Kill Process")
    fmt.Println("6) Interrupt Process")
    fmt.Println("0) Resume Output")
    fmt.Println("-1) Wait for procs and quit")
  }
InputLoop:
  for {
    line := strings.ToLower(readline())
    if line != "p" && line != "P" {
      continue
    }
    stdout.Lock()
    printChoices()
    for {
      for {
        choice, err := strconv.Atoi(readline("Choice: "))
        if err != nil {
          fmt.Println("Invalid choice")
          continue
        }
        switch choice {
        case 1:
          printChoices()
        case 2:
          printProcesses()
        case 3:
          restartProcessKill()
        case 4:
          restartProcessInterrupt()
        case 5:
          killProcess()
        case 6:
          interruptProcess()
        case 0:
          stdout.Unlock()
          continue InputLoop
        case -1:
          stdout.Unlock()
          break InputLoop
        default:
          fmt.Println("Invalid choice")
          continue
        }
        break
      }
    }
    stdout.Unlock()
  }
}

func printProcesses() {
  for _, proc := range app.procs {
    fmt.Printf(
      "Process #%d (%s): %s\n",
      proc.num, proc.Name, statusString(proc.status.Load()),
    )
  }
}

func restartProcessKill() {
  for {
    num, err := strconv.Atoi(readline("Process # (-1 = Back): "))
    if err != nil {
      fmt.Println("Invalid number")
    }
    if num == -1 {
      return
    }
    proc := app.GetProcByNum(num)
    if proc == nil {
      fmt.Println("No process with num", num)
      continue
    }
    if err := proc.kill(); err != nil {
      fmt.Println("Error killing process:", err)
      continue
    }
    if err := startProc(proc); err != nil {
      fmt.Println("Error starting process:", err)
    }
  }
}

func restartProcessInterrupt() {
  for {
    num, err := strconv.Atoi(readline("Process # (-1 = Back): "))
    if err != nil {
      fmt.Println("Invalid number")
    }
    if num == -1 {
      return
    }
    proc := app.GetProcByNum(num)
    if proc == nil {
      fmt.Println("No process with num", num)
      continue
    }
    if err := proc.interrupt(); err != nil {
      fmt.Println("Error interrupt process:", err)
      continue
    }
    if err := startProc(proc); err != nil {
      fmt.Println("Error starting process:", err)
    }
  }
}

func killProcess() {
  for {
    num, err := strconv.Atoi(readline("Process # (-1 = Back): "))
    if err != nil {
      fmt.Println("Invalid number")
    }
    if num == -1 {
      return
    }
    proc := app.GetProcByNum(num)
    if proc == nil {
      fmt.Println("No process with num", num)
      continue
    }
    if err := proc.interrupt(); err != nil {
      fmt.Println("Error killing process:", err)
    }
  }
}

func interruptProcess() {
  for {
    num, err := strconv.Atoi(readline("Process # (-1 = Back): "))
    if err != nil {
      fmt.Println("Invalid number")
    }
    if num == -1 {
      return
    }
    proc := app.GetProcByNum(num)
    if proc == nil {
      fmt.Println("No process with num", num)
      continue
    }
    if err := proc.interrupt(); err != nil {
      fmt.Println("Error interrupt process:", err)
    }
  }
}

func startProc(proc *Process) error  {
  // TODO: Do we need to wait?
  for i := 0; i < 5; i++ {
    time.Sleep(time.Second)
    if err := proc.Start(); err != errProcRunning {
      return err
    }
  }
  return fmt.Errorf("failed to start too many times")
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
    procEnv := make([]string, len(env), len(env)+len(proc.Env))
    copy(procEnv, env)
    proc.Env = append(procEnv, proc.Env...)
	}
	return app
}

func (a *App) AddProc(p *Process) {
	p.num = len(a.procs) + 1
	a.procs = append(a.procs, p)
}

func (a *App) GetProcByNum(num int) *Process {
  for _, proc := range a.procs {
    if proc.num == num {
      return proc
    }
  }
  return nil
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
			Printf(
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

func (p *Process) populateCmd() {
  if p.status.Load() == statusRunning {
    return
  }
  p.cmd = exec.Command(p.Program, p.Args...)
  p.cmd.Env = p.Env
}

func (p *Process) kill() error {
  if p.status.CompareAndSwap(statusRunning, statusFinished) {
    return nil
  }
  return p.cmd.Process.Signal(os.Kill)
}

func (p *Process) interrupt() error {
  if p.status.CompareAndSwap(statusRunning, statusFinished) {
    return nil
  }
  return p.cmd.Process.Signal(os.Interrupt)
}

func (p *Process) Start() error {
	if p.status.Load() == statusRunning {
		return errProcRunning
	}
  p.populateCmd()
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
			Printf(
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
			Printf(
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
		Printf("Error starting %s: %v\n", p.Name, err)
		// Delete the created files
		if p.outFile != nil {
			p.outFile.Close()
			if err := os.Remove(p.outFile.Name()); err != nil {
				Printf("Error removing stdout file for %s: %v\n", p.Name, err)
			}
		}
		if p.errFile != nil {
			p.errFile.Close()
			if err := os.Remove(p.errFile.Name()); err != nil {
				Printf("Error removing stderr file for %s: %v\n", p.Name, err)
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
		Printf("%s terminated with error: %v\n", p.Name, err)
	} else {
		Println(p.Name, "finished")
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

func statusString(u uint32) string {
  switch u {
  case statusNotStarted:
    return "NOT STARTED"
  case statusRunning:
    return "RUNNING"
  case statusFinished:
    return "FINISHED"
  default:
    return "UNKNOWN"
  }
}

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

type LockedWriter struct {
  w io.Writer
  sync.Mutex
}

func NewLockedWriter(w io.Writer) *LockedWriter {
  return &LockedWriter{w: w}
}

func (s *LockedWriter) Write(p []byte) (int, error) {
  s.Lock()
  defer s.Unlock()
  return s.LockedWrite(p)
}

func (s *LockedWriter) LockedWrite(p []byte) (int, error) {
  return s.w.Write(p)
}

var (
  stdout = NewLockedWriter(os.Stdout)
  stderr = NewLockedWriter(os.Stderr)
)

func Print(args ...any) (int, error) {
  return fmt.Fprint(stdout, args...)
}

func Printf(format string, args ...any) (int, error) {
  return fmt.Fprintf(stdout, format, args...)
}

func Println(args ...any) (int, error) {
  return fmt.Fprintln(stdout, args...)
}
