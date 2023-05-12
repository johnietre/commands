package main

import (
  "crypto/sha1"
  "crypto/sha256"
  "crypto/sha512"
  "encoding/base64"
  "flag"
  "fmt"
  "io"
  "log"
  "net/url"
  "os"

  uuidpkg "github.com/google/uuid"
  "golang.org/x/crypto/bcrypt"
)

func main() {
  log.SetFlags(0)

  flag.Usage = func() {
    printf := func(format string, args ...interface{}) {
      fmt.Fprintf(flag.CommandLine.Output(), format, args...)
    }
    printf("Usage: %s [SUBCOMMAND]\n", os.Args[0])
    printf("Generates a UUID or executes SUBCOMMAND\n")
    printf("  base64\t\tencode using base64\n")
    printf("  bcrypt\t\thash using bcrypt\n")
    printf("  sha1\t\t\thash using sha1\n")
    printf("  sha256\t\thash using sha256\n")
    printf("  sha512\t\thash using sha512\n")
    printf("  url\t\t\tencode as url\n")
  }
  flag.Parse()

  args := flag.Args()
  if len(args) == 0 {
    handleUUID()
    return
  }
  switch args[0] {
  case "base64":
    handleBase64(args[1:])
  case "bcrypt":
    handleBcrypt(args[1:])
  case "sha1":
    handleSha1(args[1:])
  case "sha256":
    handleSha256(args[1:])
  case "sha512":
    handleSha512(args[1:])
  case "url":
    handleURL(args[1:])
  default:
    log.Fatalf("unknown command: %s", args[0])
  }
}

func handleUUID() {
  uuid, err := uuidpkg.NewRandom()
  if err != nil {
    log.Fatalln("error generating UUID:", err)
  }
  fmt.Println(uuid.String())
}

func handleBcrypt(args []string) {
  flagSet := flag.NewFlagSet("bcrypt", flag.ExitOnError)
  input := flagSet.String(
    "what", "",
    "Hash given string using bcrypt (or use as floating last argument)",
  )
  cost := flagSet.Int(
    "cost", bcrypt.DefaultCost,
    "Cost to use for bcrypt",
  )
  flagSet.Parse(args)

  if bargs := flagSet.Args(); len(args) != 0 {
    *input = bargs[0]
  }

  hash, err := bcrypt.GenerateFromPassword([]byte(*input), *cost)
  if err != nil {
    log.Fatalln("error generating from bcrypt:", err)
  }
  fmt.Println(string(hash))
  return
}

func handleBase64(args []string) {
  flagSet := flag.NewFlagSet("", flag.ExitOnError)
  inFilePath := flagSet.String("f", "", "Use the given file as input")
  outFilePath := flagSet.String("o", "", "Use the given file as output")
  flagSet.Parse(args)

  if *inFilePath == "" {
    log.Fatal("must provide input file")
  }
  inFile, err := os.Open(*inFilePath)
  if err != nil {
    log.Fatalf("error opening input file: %v", err)
  }
  defer inFile.Close()

  var encoder io.WriteCloser
  if *outFilePath == "" {
    encoder = base64.NewEncoder(base64.URLEncoding, os.Stdout)
  } else {
    outFile, err := os.Create(*outFilePath)
    if err != nil {
      log.Fatalf("error creating output file: %v", err)
    }
    defer outFile.Close()
    encoder = base64.NewEncoder(base64.URLEncoding, outFile)
  }

  if _, err := io.Copy(encoder, inFile); err != nil {
    log.Fatalf("error encoding: %v", err)
  }
  if *outFilePath == "" {
    fmt.Println()
  }
  encoder.Close()
}

func handleSha1(args []string) {
  flagSet := flag.NewFlagSet("sha1", flag.ExitOnError)
  inFilePath := flagSet.String("f", "", "Use the given file as input")
  input := flagSet.String(
    "what", "",
    "Hash given string using SHA1 (or use as floating last argument)",
  )
  flagSet.Parse(args)

  if *inFilePath == "" {
    if trailing := flagSet.Args(); len(trailing) != 0 {
      *input = trailing[0]
    }
    fmt.Printf("%x\n", sha1.Sum([]byte(*input)))
  } else {
    f, err := os.Open(*inFilePath)
    if err != nil {
      log.Fatalf("error optning input file: %v", err)
    }
    defer f.Close()

    h := sha1.New()
    if _, err := io.Copy(h, f); err != nil {
      log.Fatalf("error encoding: %v", err)
    }
    fmt.Printf("%x\n", h.Sum(nil))
  }
}

func handleSha256(args []string) {
  flagSet := flag.NewFlagSet("sha256", flag.ExitOnError)
  inFilePath := flagSet.String("f", "", "Use the given file as input")
  input := flagSet.String(
    "what", "",
    "Hash given string using SHA256 (or use as floating last argument)",
  )
  flagSet.Parse(args)

  if *inFilePath == "" {
    if trailing := flagSet.Args(); len(trailing) != 0 {
      *input = trailing[0]
    }
    fmt.Printf("%x\n", sha256.Sum256([]byte(*input)))
  } else {
    f, err := os.Open(*inFilePath)
    if err != nil {
      log.Fatalf("error optning input file: %v", err)
    }
    defer f.Close()

    h := sha256.New()
    if _, err := io.Copy(h, f); err != nil {
      log.Fatalf("error encoding: %v", err)
    }
    fmt.Printf("%x\n", h.Sum(nil))
  }
}

func handleSha512(args []string) {
  flagSet := flag.NewFlagSet("sha512", flag.ExitOnError)
  inFilePath := flagSet.String("f", "", "Use the given file as input")
  input := flagSet.String(
    "what", "",
    "Hash given string using SHA512 (or use as floating last argument)",
  )
  flagSet.Parse(args)

  if *inFilePath == "" {
    if trailing := flagSet.Args(); len(trailing) != 0 {
      *input = trailing[0]
    }
    fmt.Printf("%x\n", sha512.Sum512([]byte(*input)))
  } else {
    f, err := os.Open(*inFilePath)
    if err != nil {
      log.Fatalf("error optning input file: %v", err)
    }
    defer f.Close()

    h := sha512.New()
    if _, err := io.Copy(h, f); err != nil {
      log.Fatalf("error encoding: %v", err)
    }
    fmt.Printf("%x\n", h.Sum(nil))
  }
}

func handleURL(args []string) {
  flagSet := flag.NewFlagSet("url", flag.ExitOnError)
  input := flagSet.String(
    "what", "",
    "Encode given string as URL (or use as floating last argument)",
  )
  flagSet.Parse(args)

  if trailing := flagSet.Args(); len(trailing) != 0 {
    *input = trailing[0]
  }
  fmt.Println(url.QueryEscape(*input))
}
