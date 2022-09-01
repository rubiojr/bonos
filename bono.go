package main

import (
	"errors"
	"time"
)

type Bono struct {
	CreatedAt string   `json:"created_at"`
	Remaining uint32   `json:"remaining"`
	Stamps    []string `json:"stamps"`
	Username  string   `json:"username"`
}

func (b *Bono) Use() error {
	if b.Remaining <= 0 {
		return errors.New("no more passes left")
	}

	b.Remaining -= 1
	b.Stamps = append(b.Stamps, time.Now().Format(time.RFC3339))

	return nil
}

func (b *Bono) IsValid() bool {
	return b.Remaining > 0
}
