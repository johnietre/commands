package main

import (
  "database/sql"

  _ "github.com/mattn/go-sqlite3"
)

type DB *sql.DB

func ConnectDB(dbPath string) (DB, error) {
  return sql.Open("sqlite3", dbPath)
}

func (db DB) NewAccount(name, desc string, initialBal int64) (*Account, error) {
  res, err := db.Exec(
    `INSERT INTO ACCOUNTS(NAME, DESCRIPTION, BALANCE) VALUES (?, ?, ?)`,
    name, desc, initialBal,
  )
  if err != nil {
    return nil, err
  }
  account := NewAccount(name, desc, initialBal)
  if account.ID, err = res.LastInsertId(); err != nil {
    return nil, err
  }
  return account, nil
}

func (db DB) NewAccountWithTime(name, desc string, initialBal, openedAt int64) (*Account, error) {
  res, err := db.Exec(
    `INSERT INTO ACCOUNTS(NAME, DESCRIPTION, BALANCE, OPENED_AT) VALUES (?, ?, ?, ?)`,
    name, desc, initialBal, openedAt,
  )
  if err != nil {
    return nil, err
  }
  account := NewAccountWithTime(name, desc, initialBal, openedAt)
  if account.ID, err = res.LastInsertId(); err != nil {
    return nil, err
  }
  return account, nil
}

func (db DB) AddAccount(account *Account) error {
  var (
    stmt string
    res sql.Result
    err error
  )
  if account.OpenedAt == 0 {
    stmt = `INSERT INTO ACCOUNTS(NAME, DESCRIPTION, BALANCE) VALUES (?, ?, ?)`
    res, err = db.Exec(
      stmt,
      account.Name, account.Description, account.Balance,
    )
  } else {
    stmt = `INSERT INTO ACCOUNTS(NAME, DESCRIPTION, BALANCE, OPENED_AT) VALUES (?, ?, ?, ?)`
    res, err = db.Exec(
      stmt,
      account.Name, account.Description, account.Balance, account.OpenedAt,
    )
  }
  if err != nil {
    return err
  }
  account.ID, err = res.LastInsertId()
  return err
}

func (db DB) CloseAccount(id int) error {
  _, err := db.Exec(
    `UPDATE ACCOUNTS SET CLOSED_AT=? WHERE ID=? AND WHERE CLOSED_AT NOT NULL`,
    time.Now().Unix(), id,
  )
  if err != nil {
    return err
  }
  return nil
}

func (db DB) NewTx(fromID, toID int, amount int64, desc string) (from *Account, to *Account, err error) {
  _, err = db.Exec(
    `INSERT INTO TRANSACTIONS(FROM_ID, TO_ID, AMOUNT, DESCRIPTION) VALUES (?, ?, ?, ?)`,
    fromID, toID, amount, desc,
  )
  if err != nil {
    return
  }
  stmt, err := db.Prepare(`SELECT 
}

func (db DB) NewTxWithTime(from, to int, amount int64, desc string, timestamp int64) (*Tx, error) {

}
