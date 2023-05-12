package main

import (
  "flag"
  "fmt"
  "io/fs"
  "log"
  "os"
  "path/filepath"
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

  size uint64
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

  args := os.Args[1:]
  var path string
  for i, arg := range args {
    if arg[0] != '-' {
      path = arg
      args = append(args[:i], args[i+1:]...)
      break
    }
  }
  if len(path) == 0 {
    log.Fatal("must provide path")
  }
  flag.CommandLine.Parse(args)

  info, err := os.Stat(path)
  if err != nil {
    log.Fatalf("error getting info: %v", err)
  }
  if !info.IsDir() {
    printInfo(info, nil)
    return
  }
  wg.Add(1)
  walkDir(path)
  wg.Wait()
  printInfo(info, &size)
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
      atomic.AddUint64(&size, uint64(info.Size()))
      continue
    } else if recursive {
      wg.Add(1)
      go walkDir(filepath.Join(path, ent.Name()))
    }
  }
}

func printInfo(info fs.FileInfo, sizePtr *uint64) {
  var size uint64
  if sizePtr == nil {
    size = uint64(info.Size())
  } else {
    size = *sizePtr
  }
  var sizeStr string
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
  fmt.Printf(
    "%s\nSize: %s\nLast Mod: %s\n",
    info.Name(), sizeStr, info.ModTime().Format("15:04 Jan 02, 2006"),
  )
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
