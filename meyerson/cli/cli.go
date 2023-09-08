package cli

import (
	"bufio"
  "flag"
	"fmt"
	"os"
	"os/exec"
  "path/filepath"
	"strings"
	"sync"
)

func Run(args []string) {
  var outDir string
  flag.StringVar(&outDir, "out-dir", ".", "Directory to put output files in")
  flag.CommandLine.Parse(args)

  // Create the processes
  var procs []*Process
  for i := 1; true; i++ {
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
    choice := strings.ToLower(readline("Pipe stdout [Y/n]? "))
    pipeStdout := choice == "y" || choice == "yes"
    choice = strings.ToLower(readline("Pipe stderr [Y/n]? "))
    pipeStderr := choice == "y" || choice == "yes"

    cmd := exec.Command(prog, args...)
    cmd.Env = env
    procs = append(procs, &Process{
      name: name,
      cmd: cmd,
      pipeStdout: pipeStdout,
      pipeStderr: pipeStderr,
    })
    fmt.Println("====================")
  }
  fmt.Println("========================================")
  // Check for any deletions
  for {
    name := readline("Delete any procs (enter name)? ")
    if name == "" {
      break
    }
    deleted := false
    for i, proc := range procs {
      if proc.name == name {
        procs = append(procs[:i], procs[i+1:]...)
        deleted = true
        break
      }
    }
    if !deleted {
      fmt.Println("No process with name: ", name)
    }
  }
  fmt.Println("========================================")
  // Start the processes
  fmt.Println("Starting processes...")
  var wg sync.WaitGroup
  for i, proc := range procs {
    var err error
    // Open the files for output
    var stdoutFile, stderrFile *os.File
    if proc.pipeStdout {
      stdoutFile, err = os.Create(
        filepath.Join(outDir, fmt.Sprintf("process%d-stdout.txt", i+1)),
      )
      if err != nil {
        fmt.Printf(
          "Error creating stdout output file for %s: %v\n",
          proc.name, err,
        )
      }
      proc.cmd.Stdout = stdoutFile
    }
    if proc.pipeStderr {
      fmt.Println("piping")
      stderrFile, err = os.Create(
        filepath.Join(outDir, fmt.Sprintf("process%d-stderr.txt", i)),
      )
      if err != nil {
        fmt.Printf(
          "Error creating stderr output file for %s: %v\n",
          proc.name, err,
        )
      }
      proc.cmd.Stderr = stderrFile
    }
    // Start the process
    if err := proc.cmd.Start(); err != nil {
      fmt.Printf("Error starting %s: %v\n", proc.name, err)
      // Delete the created files
      if stdoutFile != nil {
        stdoutFile.Close()
        if err := os.Remove(stdoutFile.Name()); err != nil {
          fmt.Printf("Error removing stdout file for %s: %v\n", proc.name, err)
        }
      }
      if stderrFile != nil {
        stderrFile.Close()
        if err := os.Remove(stderrFile.Name()); err != nil {
          fmt.Printf("Error removing stderr file for %s: %v\n", proc.name, err)
        }
      }
      continue
    }
    // Wait for the process to finish
    wg.Add(1)
    go func(proc *Process, outFile, errFile *os.File) {
      if err := proc.cmd.Wait(); err != nil {
        fmt.Printf("%s terminated with error: %v\n", proc.name, err)
      } else {
        fmt.Println(proc.name, "finished")
      }
      if outFile != nil {
        outFile.Close()
      }
      if errFile != nil {
        errFile.Close()
      }
      wg.Done()
    }(proc, stdoutFile, stderrFile)
  }
  wg.Wait()
}

type Process struct {
  name string
  cmd *exec.Cmd
  pipeStdout, pipeStderr bool
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
