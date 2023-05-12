package main

import (
  "fmt"
  "net/url"
  "os"
  "os/exec"
  "path/filepath"
)

func main() {
  if len(os.Args) != 2 {
    die("Usage: ghclone <REPO URL/NAME>")
  }
  ghurl, err := url.Parse("https://github.com")
  if err != nil {
    die("bad URL, the developer's an idiot, I guess")
  }
  if !contains(os.Args[1], '/') {
    dir, err := os.Getwd()
    if err != nil {
      die("error getting directory: %v", err)
    }
    ghurl.Path = filepath.Base(dir) + "/" + os.Args[1]
  } else {
    ghurl.Path = os.Args[1]
  }
  cmd := exec.Command("git", "clone", ghurl.String())
  cmd.Stdin = os.Stdin
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  if err := cmd.Run(); err != nil {
    die("error running: %v", err)
  }
}

func contains(s string, b byte) bool {
  for i := range s {
    if s[i] == b {
      return true
    }
  }
  return false
}

func die(format string, args ...interface{}) {
  fmt.Fprintf(os.Stderr, format+"\n", args...)
  os.Exit(1)
}
