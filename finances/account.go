package main

type Account struct {
  ID int
  Name string
  Description string
  Balance int64
  OpenedAt int64 // Unix timestamp
  ClosedAt int64 // Unix timestamp
}

func NewAccount(name, desc string, initialBal int64) *Account {
  return &Account{
    Name: name,
    Description: desc,
    Balance: initialBal,
  }
}

func NewAccountWithTime(name, desc string, initialBal, openedAt int64) *Account {
  return &Account{
    Name: name,
    Description: desc,
    Balance: initialBal,
    OpenedAt: openedAt,
  }
}
