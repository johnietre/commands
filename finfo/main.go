package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type SizeDenom uint64

func (sd *SizeDenom) Set(s string) error {
  switch strings.ToLower(s) {
  case "b":
    *sd = B
  case "kb":
    *sd = KB
  case "kib":
    *sd = KiB
  case "mb":
    *sd = MB
  case "mib":
    *sd = MiB
  case "gb":
    *sd = GB
  case "gib":
    *sd = GiB
  default:
    n, err := strconv.ParseUint(s, 0, 64)
    if err != nil {
      return err
    }
    *sd = SizeDenom(n)
  }
  return nil
}

func (sd SizeDenom) String() string {
  switch sd {
  case B:
    return "B"
  case KB:
    return "KB"
  case KiB:
    return "KiB"
  case MB:
    return "MB"
  case MiB:
    return "MiB"
  case GB:
    return "GB"
  case GiB:
    return "GiB"
  }
  return fmt.Sprintf("%d B", sd)
}

const (
  B SizeDenom = 1
  KB = 1_000
  KiB = 1 << 10
  MB = 1_000_000
  MiB = 1 << 20
  GB = 1_000_000_000
  GiB = 1 << 30
)

var (
  sizeDenom = B
  recursive = false
  prec = -1

  matchFunc = func(string) bool { return true }
  size uint64
  totalSize uint64
  wg sync.WaitGroup
)

func main() {
  log.SetFlags(0)

  flag.Var(
    &sizeDenom, "size",
    "Size denomination (B/KB/KiB/MB/MiB/GB/GiB/Custom)",
  )
  flag.BoolVar(
    &recursive, "r", false,
    "If directory, get size of all subdirectories (recursive)",
  )
  flag.IntVar(
    &prec, "p", -1,
    "Precision of output decimals (-1 for no limit)",
  )
  total := flag.Bool("total", false, "Calculate the total size of all args")
  // Used only for when looping through the arguments since it can be passed
  // multiple times
  flag.String(
    "path", "", "Explicitly specify path (can be passed multiple times)",
  )
  regexStr := flag.String(
    "regex", "",
    "Regex expression to match paths against when calculating size",
  )
  excludeMatch := flag.Bool(
    "excl", false, "Exclude regular expression matches",
  )
  filesOnly := flag.Bool(
    "files", false,
    "Match only files with regular expressions; shouldn't pass --dirs with this flag",
  )
  dirsOnly := flag.Bool(
    "dirs", false, 
    "Match only directories with regular expressions; shouldn't pass --files with this flag",
  )
  boolFlags := map[string]bool{
    "r": true,
    "total": true,
    "excl": true,
    "files": true,
    "dirs": true,
  }

  // Print help
  if len(os.Args) == 2 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
    flag.Parse()
  }
  var args, paths []string
  // Keeps track of whether the previous arg was a flag with a value yet to be
  // passed
  prevFlag := false
  for i, l := 1, len(os.Args); i < l; i++ {
    arg := os.Args[i]
    if prevFlag {
      prevFlag = false
      args = append(args, arg)
      continue
    } else if arg[0] != '-' {
      paths = append(paths, arg)
      continue
    }
    // Get the flag name
    name, value := "", ""
    l := len(arg)
    if l > 1 && arg[1] == '-' {
      name = arg[2:]
    } else {
      name = arg[1:]
    }
    // Split value from flag
    if index := strings.IndexByte(name, '='); index != -1 {
      name, value = name[:index], name[index+1:]
    } else {
      prevFlag = true
    }
    // Check if it is the path flag
    if name == "path" {
      if value != "" {
        // Add the value if it was defined with the flag
        paths = append(paths, value)
      } else if prevFlag {
        // Get the next arg (the value) if it exists
        if i != l - 1 {
          i++
          paths = append(paths, os.Args[i+1])
        } else {
          // If this is the last flag passed but no value is passed afterwards
          // (checked with `i != l - 1`), then throw an error. This can be done
          // by just including it in the arguments passed to Parse() since it
          // will throw an error when a flag is passed without a defined value.
          // We can break since we know it's the last arg
          args = append(args, arg)
          break
        }
      }
      prevFlag = false
      continue
    }
    // See if the flag is defined
    if flag.Lookup(name) != nil {
      args = append(args, arg)
      // No value will be defined in next arg if this is a boolean flag
      prevFlag = !boolFlags[name] || value == ""
    } else {
      paths = append(paths, arg)
      prevFlag = false
    }
  }

  flag.CommandLine.Parse(args)
  if len(paths) == 0 {
    log.Fatal("must provide path")
  }

  if *regexStr != "" {
    regex, err := regexp.Compile(*regexStr)
    if err != nil {
      log.Fatal("invalid regex: ", err)
    }
    matchFunc = regex.MatchString
  }
  if *excludeMatch {
    f := matchFunc
    matchFunc = func(s string) bool { return !f(s) }
  }
  if *filesOnly {
    f := matchFunc
    matchFunc = func(s string) bool {
      l := len(s)
      return l == 0 || s[l-1] == '/' || f(s)
    }
  }
  if *dirsOnly {
    f := matchFunc
    matchFunc = func(s string) bool {
      l := len(s)
      println(s, l == 0 || s[l-1] != '/' || f(s))
      return l == 0 || s[l-1] != '/' || f(s)
    }
  }

  for i, path := range paths {
    if i != 0 {
      fmt.Println(strings.Repeat("=", 40))
    }
    info, err := os.Stat(path)
    if err != nil {
      //log.Fatalf("error getting info: %v", err)
      log.Printf("error getting info: %v", err)
      continue
    }
    if !info.IsDir() {
      size = uint64(info.Size())
    } else {
      wg.Add(1)
      walkDir(path)
      wg.Wait()
    }
    printInfo(info, size)
    if *total {
      totalSize += size
    }
    size = 0
  }
  if *total {
    fmt.Println(strings.Repeat("=", 40))
    fmt.Println("Total Size:", makeSizeStr(totalSize))
  }
}

func walkDir(path string) {
  defer wg.Done()
  ents, err := os.ReadDir(path)
  if err != nil {
    log.Printf("error opening %s: %v", path, err)
    return
  }
  for _, ent := range ents {
    info, err := ent.Info()
    if err != nil {
      log.Printf("error getting info for %s: %v", filepath.Join(path, ent.Name()), err)
      continue
    }
    if !info.IsDir() {
      if matchFunc(ent.Name()) {
        atomic.AddUint64(&size, uint64(info.Size()))
      }
      continue
    } else if recursive {
      name := ent.Name()
      if l := len(name); l == 0 || name[l-1] != '/' {
        name += "/"
      }
      if matchFunc(name) {
        wg.Add(1)
        go walkDir(filepath.Join(path, ent.Name()))
      }
    }
  }
}

func printInfo(info fs.FileInfo, size uint64) {
  sizeStr := makeSizeStr(size)
  fmt.Printf(
    "%s\nSize: %s\nLast Mod: %s\n",
    info.Name(), sizeStr, info.ModTime().Format("15:04 Jan 02, 2006"),
  )
}

func makeSizeStr(size uint64) (sizeStr string) {
  switch sizeDenom {
  case B:
    sizeStr = commas(size) + " B"
  case KB:
    sizeStr = commas(size / KB)
    if rem := size % KB; rem != 0 {
      sizeStr += strconv.FormatFloat(float64(rem) / float64(KB), 'f', prec, 64)[1:]
    }
    sizeStr += " KB"
  case KiB:
    sizeStr = commas(size / KiB)
    if rem := size % KiB; rem != 0 {
      sizeStr += strconv.FormatFloat(float64(rem) / float64(KiB), 'f', prec, 64)[1:]
    }
    sizeStr += " KiB"
  case MB:
    sizeStr = commas(size / MB)
    if rem := size % MB; rem != 0 {
      sizeStr += strconv.FormatFloat(float64(rem) / float64(MB), 'f', prec, 64)[1:]
    }
    sizeStr += " MB"
  case MiB:
    sizeStr = commas(size / MiB)
    if rem := size % MiB; rem != 0 {
      sizeStr += strconv.FormatFloat(float64(rem) / float64(MiB), 'f', prec, 64)[1:]
    }
    sizeStr += " MiB"
  case GB:
    sizeStr = commas(size / GB)
    if rem := size % GB; rem != 0 {
      sizeStr += strconv.FormatFloat(float64(rem) / float64(GB), 'f', prec, 64)[1:]
    }
    sizeStr += " GB"
  case GiB:
    sizeStr = commas(size / GiB)
    if rem := size % GiB; rem != 0 {
      sizeStr += strconv.FormatFloat(float64(rem) / float64(GiB), 'f', prec, 64)[1:]
    }
    sizeStr += " GiB"
  default:
    sd := uint64(sizeDenom)
    sizeStr = commas(size / sd)
    if rem := size % sd; rem != 0 {
      sizeStr += strconv.FormatFloat(float64(rem) / float64(sd), 'f', prec, 64)[1:]
    }
    sizeStr += fmt.Sprintf(" /%d B", sd)
  }
  return
}

func commas(u uint64) string {
  numStr := strconv.FormatUint(u, 10)
  str := ""
  // Track when to place comma with cc
  for i, cc := len(numStr) - 1, -1; i >= 0; i-- {
    cc++
    if cc == 3 {
      str, cc = "," + str, 0
    }
    str = string(numStr[i]) + str
  }
  return str
}
