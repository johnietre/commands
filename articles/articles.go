package main

import (
  "database/sql"
  "fmt"
  "log"
  "path"
  "runtime"
  "strings"

  _ "github.com/mattn/go-sqlite3"
)

type article struct {
  ID int64
  Name string
  URL string
  Read bool
  Favorite bool
}

func main() {
  // Set the log flags
  log.SetFlags(0)

  // Get the source file path
  _, filePath, _, ok := runtime.Caller(0)
  if !ok {
    log.Fatal("error getting file path")
  }
  // Get the directory of the source file
  dir := path.Dir(filePath)
  // Get the db path and open the database
  dbPath := path.Join(dir, "articles.db")
  db, err := sql.Open("sqlite3", dbPath)
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  // Turn on foriegn keys when using SQLITE3
  if res, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
    log.Fatalf("error turning on foreign keys: %v", err)
  } else if n, err := res.RowsAffected(); err != nil {
    log.Fatalf("error turning on foreign keys: %v", err)
  } else if n == 0 {
    log.Fatalf("error turning on foreign keys")
  }
}

func newArticle(db *sql.DB, name, url string, read, favorite bool, collections ...string) error {
  stmt := `INSERT INTO articles (name, url, read, favorite) VALUES (?,?,?,?)`
  res, err := db.Exec(stmt, name, url, read, favorite)
  if err != nil {
    if strings.Contains(err.Error(), "unique constraint failed") {
      return fmt.Errorf("article url already exists")
    }
    return fmt.Errorf("error adding article: %v", err)
  }
  id, err := res.LastInsertId()
  if err != nil {
    if collections == nil {
      return nil
    }
    return fmt.Errorf("error getting article id: %v", err)
  }
  for _, col := range collections {
    stmt = fmt.Sprintf(`INSERT INTO %s (article_id) VALUES (?)`, col)
    if _, err := db.Exec(stmt, id); err != nil {
      if strings.Contains(err.Error(), "no such table") {
        log.Println("error adding article to collection: collection %s doesn't exist", col)
      } else {
        log.Printf("error adding article to collection: %v", err)
      }
    }
  }
  return nil
}

/* TODO: Finish */
func addArticleToCollections(db *sql.Db, articleName, articleURL string, collections ...string) error {
  var id int64
  if articleURL != "" {
    stmt := fmt.Sprintf(`SELECT id FROM articles WHERE url="%s"`, articleURL)
    row := db.QueryRow(stmt)
    if err := row.Scan(&id); err != nil {
      if errors.Is(err, sql.ErrNoRows) {
        if articleName == "" {
          return fmt.Errorf("no article with given url")
        }
      }
    }
  }
}

func deleteArticle(db *sql.DB, name, url string) error {
  return nil
}

func newCollection(db *sql.DB, name string) error {
  stmt := `CREATE TABLE %s (
    article_id INTEGER,
    FOREIGN KEY(article_id) REFERENCES articles(id)
  )`
  stmt := fmt.Sprintf(stmt, name)
  if _, err := db.Exec(stmt); err != nil {
    if strings.Contains(err.Error(), "already exists") {
      return fmt.Errorf("collection already exists")
    }
    return fmt.Errorf("error creating collection: %v", err)
  }
  return nil
}

func deleteCollection(db *sql.DB, name string) error {
  _, err := db.Exec(fmt.Sprintf(`DROP TABLE %s`, name))
  if err != nil [
    err = fmt.Errorf("error deleting collection: %v", err)
  }
  return err
}
