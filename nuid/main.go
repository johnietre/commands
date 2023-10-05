package main

import (
  "bytes"
  "crypto/sha1"
  "crypto/sha256"
  "crypto/sha512"
  "encoding/base64"
  "flag"
  "fmt"
  "hash"
  "io"
  "log"
  "net/url"
  "os"
  "strings"

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
    printf("  sha224\t\thash using sha224\n")
    printf("  sha256\t\thash using sha256\n")
    printf("  sha384\t\thash using sha384\n")
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
  case "sha224":
    handleSha224(args[1:])
  case "sha256":
    handleSha256(args[1:])
  case "sha384":
    handleSha384(args[1:])
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

type base64Encoding struct {
  *base64.Encoding
}

func (be *base64Encoding) Set(val string) error {
  switch val {
  case "std":
    be.Encoding = base64.RawStdEncoding
  case "url":
    be.Encoding = base64.RawURLEncoding
  case "pstd":
    be.Encoding = base64.StdEncoding
  case "purl":
    be.Encoding = base64.URLEncoding
  default:
    return fmt.Errorf("invalid value: "+val)
  }
  return nil
}

func (be base64Encoding) String() string {
  if be.Encoding == base64.RawStdEncoding {
    return "std"
  } else if be.Encoding == base64.RawURLEncoding {
    return "url"
  } else if be.Encoding == base64.StdEncoding {
    return "pstd"
  } else if be.Encoding == base64.URLEncoding {
    return "purl"
  }
  return "unknown"
}

func handleBase64(args []string) {
  flagSet := flag.NewFlagSet("", flag.ExitOnError)
  encoding := &base64Encoding{Encoding: base64.RawStdEncoding}
  inFilePath := flagSet.String("f", "", "Use the given file as input")
  input := flagSet.String(
    "what", "",
    "Encode/decode given string using base64 (or use as floating last argument)",
  )
  bufferStdin := flagSet.Bool(
    "buffer", true,
    "When reading from stdin, will buffer everything until all input is read" +
    " (useful since data is encoded/decoded and output in realtime)",
  )
  outFilePath := flagSet.String("o", "", "Use the given file as output")
  decode := flagSet.Bool("d", false, "Decode base64")
  flagSet.Var(
    encoding, "enc",
    "Encoding to use (std = standard, url = URL, p[std/url] = padded)",
  )
  flagSet.Parse(args)

  var writer io.Writer
  if *outFilePath == "" {
    writer = os.Stdout
  } else {
    f, err := os.Create(*outFilePath)
    if err != nil {
      log.Fatalf("error creating output file: %v", err)
    }
    defer f.Close()
    writer = f
  }

  var reader io.Reader
  if *inFilePath == "" {
    if trailing := flagSet.Args(); len(trailing) != 0 {
      reader = strings.NewReader(trailing[0])
    } else if *input != "" {
      reader = strings.NewReader(*input)
    } else {
      if *bufferStdin {
        data, err := io.ReadAll(os.Stdin)
        if err != nil {
          log.Fatalf("error reading input: %v", err)
        }
        reader = bytes.NewReader(data)
      } else {
        reader = os.Stdin
      }
    }
  } else {
    f, err := os.Open(*inFilePath)
    if err != nil {
      log.Fatalf("error opening input file: %v", err)
    }
    reader = f
    defer f.Close()
  }

  if *decode {
    decoder := base64.NewDecoder(encoding.Encoding, reader)
    if _, err := io.Copy(writer, decoder); err != nil {
      log.Fatalf("error decoding: %v", err)
    }
  } else {
    encoder := base64.NewEncoder(encoding.Encoding, writer)
    if _, err := io.Copy(encoder, reader); err != nil {
      log.Fatalf("error encoding: %v", err)
    }
    encoder.Close()
  }
  writer.Write([]byte{'\n'})
}

func handleSha1(args []string) {
  flagSet := flag.NewFlagSet("sha1", flag.ExitOnError)
  inFilePath := flagSet.String("f", "", "Use the given file as input")
  input := flagSet.String(
    "what", "",
    "Hash given string using SHA1 (or use as floating last argument)",
  )
  flagSet.Parse(args)

  var reader io.Reader
  if *inFilePath == "" {
    if trailing := flagSet.Args(); len(trailing) != 0 {
      reader = strings.NewReader(trailing[0])
    } else if *input != "" {
      reader = strings.NewReader(*input)
    } else {
      reader = os.Stdin
    }
  } else {
    f, err := os.Open(*inFilePath)
    if err != nil {
      log.Fatalf("error optning input file: %v", err)
    }
    defer f.Close()
    reader = f
  }
  h := sha1.New()
  if _, err := io.Copy(h, reader); err != nil {
    log.Fatalf("error encoding: %v", err)
  }
  fmt.Printf("%x\n", h.Sum(nil))
}

func handleSha224(args []string) {
  flagSet := flag.NewFlagSet("sha224", flag.ExitOnError)
  inFilePath := flagSet.String("f", "", "Use the given file as input")
  input := flagSet.String(
    "what", "",
    "Hash given string using SHA224 (or use as floating last argument)",
  )
  flagSet.Parse(args)

  var reader io.Reader
  if *inFilePath == "" {
    if trailing := flagSet.Args(); len(trailing) != 0 {
      reader = strings.NewReader(trailing[0])
    } else if *input != "" {
      reader = strings.NewReader(*input)
    } else {
      reader = os.Stdin
    }
  } else {
    f, err := os.Open(*inFilePath)
    if err != nil {
      log.Fatalf("error optning input file: %v", err)
    }
    defer f.Close()
    reader = f
  }
  h := sha256.New224()
  if _, err := io.Copy(h, reader); err != nil {
    log.Fatalf("error encoding: %v", err)
  }
  fmt.Printf("%x\n", h.Sum(nil))
}

func handleSha256(args []string) {
  flagSet := flag.NewFlagSet("sha256", flag.ExitOnError)
  inFilePath := flagSet.String("f", "", "Use the given file as input")
  input := flagSet.String(
    "what", "",
    "Hash given string using SHA256 (or use as floating last argument)",
  )
  flagSet.Parse(args)

  var reader io.Reader
  if *inFilePath == "" {
    if trailing := flagSet.Args(); len(trailing) != 0 {
      reader = strings.NewReader(trailing[0])
    } else if *input != "" {
      reader = strings.NewReader(*input)
    } else {
      reader = os.Stdin
    }
  } else {
    f, err := os.Open(*inFilePath)
    if err != nil {
      log.Fatalf("error optning input file: %v", err)
    }
    defer f.Close()
    reader = f
  }
  h := sha256.New()
  if _, err := io.Copy(h, reader); err != nil {
    log.Fatalf("error encoding: %v", err)
  }
  fmt.Printf("%x\n", h.Sum(nil))
}

func handleSha384(args []string) {
  flagSet := flag.NewFlagSet("sha384", flag.ExitOnError)
  inFilePath := flagSet.String("f", "", "Use the given file as input")
  input := flagSet.String(
    "what", "",
    "Hash given string using SHA-384 (or use as floating last argument)",
  )
  flagSet.Parse(args)

  var reader io.Reader
  if *inFilePath == "" {
    if trailing := flagSet.Args(); len(trailing) != 0 {
      reader = strings.NewReader(trailing[0])
    } else if *input != "" {
      reader = strings.NewReader(*input)
    } else {
      reader = os.Stdin
    }
  } else {
    f, err := os.Open(*inFilePath)
    if err != nil {
      log.Fatalf("error optning input file: %v", err)
    }
    defer f.Close()
    reader = f
  }
  h := sha512.New384()
  if _, err := io.Copy(h, reader); err != nil {
    log.Fatalf("error encoding: %v", err)
  }
  fmt.Printf("%x\n", h.Sum(nil))
}

func handleSha512(args []string) {
  flagSet := flag.NewFlagSet("sha512", flag.ExitOnError)
  inFilePath := flagSet.String("f", "", "Use the given file as input")
  input := flagSet.String(
    "what", "",
    "Hash given string using SHA-512 (or use as floating last argument)",
  )
  variant := flagSet.String(
    "var", "",
    "SHA-512 variant to use. Accepted values: 224 (SHA-512/224), " +
    "256 (SHA-512/256). Nothing means standard SHA-512",
  )
  flagSet.Parse(args)

  var reader io.Reader
  if *inFilePath == "" {
    if trailing := flagSet.Args(); len(trailing) != 0 {
      reader = strings.NewReader(trailing[0])
    } else if *input != "" {
      reader = strings.NewReader(*input)
    } else {
      reader = os.Stdin
    }
  } else {
    f, err := os.Open(*inFilePath)
    if err != nil {
      log.Fatalf("error optning input file: %v", err)
    }
    defer f.Close()
    reader = f
  }
  var h hash.Hash
  switch *variant {
  case "":
    h = sha512.New()
  case "224":
    h = sha512.New512_224()
  case "256":
    h = sha512.New512_256()
  default:
    log.Fatal("invalid SHA-512 variant: " + *variant)
  }
  if _, err := io.Copy(h, reader); err != nil {
    log.Fatalf("error encoding: %v", err)
  }
  fmt.Printf("%x\n", h.Sum(nil))
}

func handleURL(args []string) {
  flagSet := flag.NewFlagSet("url", flag.ExitOnError)
  input := flagSet.String(
    "what", "",
    "Escape given string for URLs (or use as floating last argument)",
  )
  unescape := flagSet.Bool("d", false, "Unescape URL-escaped string")
  flagSet.Parse(args)

  if trailing := flagSet.Args(); len(trailing) != 0 {
    *input = trailing[0]
  }
  if *unescape {
    s, err := url.QueryUnescape(*input)
    if err != nil {
      log.Fatalf("error unescaping: %v", err)
    }
    fmt.Println(s)
  } else {
    fmt.Println(url.QueryEscape(*input))
  }
}
