package main

import (
  "time"
)

type Tx struct {
  FromID int // Account id
  ToID int // Account id
  Amount int64
  Description string
  Timestamp int64 // Unix timestamp
}

func NewTx(from, to int, amount int64, desc string) *Tx {
  return &Tx{
    From: from,
    To: to,
    Amount: amount,
    Description: desc,
  }
}
