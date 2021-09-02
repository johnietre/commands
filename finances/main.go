package main

import (
  "database/sql"
  "log"
  "path"
)

var (
  logger = log.New(os.Stderr, "", 0)
  db DB
)

func init() {
  _, thisFile, _, ok := runtime.Caller(0)
  if !ok {
    logger.Fatal("error loading source file")
  }
  thisDir := path.Dir(thisFile)

  var err error
  db, err = ConnectDB(path.Join(thisDir, "finances.db"))
  if err != nil {
    logger.Fatalf("error connecting to db: %v", err)
  }
}

func main() {
  defer func() {
    if err := db.Close(); err != nil {
      logger.Fatal(err)
    }
  }()
}
