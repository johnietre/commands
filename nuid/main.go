package main

import (
  "flag"
  "fmt"
  "log"

  uuidpkg "github.com/google/uuid"
  "golang.org/x/crypto/bcrypt"
)

func main() {
  log.SetFlags(0)

  bcryptPwd := flag.String("bcrypt", "", "Hash given string using bcrypt")
  bcryptCost := flag.Int("bcrypt-cost", bcrypt.DefaultCost, "Cost to use for bcrypt")
  flag.Parse()

  if bcryptPwd != nil && *bcryptPwd != "" {
    hash, err := bcrypt.GenerateFromPassword([]byte(*bcryptPwd), *bcryptCost)
    if err != nil {
      log.Fatalln("error generating from bcrypt:", err)
    }
    fmt.Println(string(hash))
    return
  }

  uuid, err := uuidpkg.NewRandom()
  if err != nil {
    log.Fatalln("error generating UUID:", err)
  }
  fmt.Println(uuid.String())
}
