package main

import (
	"bytes"
	"encoding/gob"

	"github.com/prologic/bitcask"
)

type Archive string

func (a Archive) Put(key string, c Concept) error {
	var buf bytes.Buffer
	var enc = gob.NewEncoder(&buf)

	err := enc.Encode(c)
	if err != nil { return err }

	db, err := bitcask.Open(string(a))
	if err != nil { return err }
	defer db.Close()

	return db.Put([]byte(key), buf.Bytes())
}

func (a Archive) Get(key string) (Concept, error) {
	var buf bytes.Buffer
	var c Concept
	var dec = gob.NewDecoder(&buf)

	db, err := bitcask.Open(string(a))
	if err != nil { return Concept{}, err }
	defer db.Close()

	err = dec.Decode(&c)
	return c, err
}
