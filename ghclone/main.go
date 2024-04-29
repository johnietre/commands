package main

import (
  "log"
  "net/url"
  "os"
  "os/exec"
  "path"
  "path/filepath"
  "strings"
)

func main() {
  log.SetFlags(0)

  if len(os.Args) != 2 {
    log.Fatal(
"Usage: ghclone <REPO URL/NAME>\n"+
"NOTE: Set default github repo user slug using GHCLONE_SLUG environment variable",
)
  }
  ghurl, err := url.Parse("https://github.com")
  if err != nil {
    log.Fatal("bad URL, the developer's an idiot, I guess")
  }
  if strings.Contains(os.Args[1], "/") {
    ghurl.Path = path.Join(os.Getenv("GHCLONE_SLUG"), os.Args[1])
  } else {
    dir, err := os.Getwd()
    if err != nil {
      log.Fatal("error getting directory: ", err)
    }
    ghurl.Path = path.Join(filepath.Base(dir), os.Args[1])
  }
  cmd := exec.Command("git", "clone", ghurl.String())
  cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
  if err := cmd.Run(); err != nil {
    log.Fatal("error running: ", err)
  }
}
