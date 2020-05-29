package main

import (
	"bytes"
	"encoding/gob"

	"github.com/prologic/bitcask"
)

type Archive string

// TODO: add delete method.

// Put saves the Bucket into the database.
func (a Archive) Put(key string, b Bucket) error {
	var buf bytes.Buffer
	var enc = gob.NewEncoder(&buf)

	err := enc.Encode(b)
	if err != nil {
		return err
	}

	db, err := bitcask.Open(string(a))
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Put([]byte(key), buf.Bytes())
}

// Get retrieves a bucket from the database.
func (a Archive) Get(key string) (Bucket, error) {
	var buf bytes.Buffer
	var bt Bucket
	var dec = gob.NewDecoder(&buf)

	db, err := bitcask.Open(string(a))
	if err != nil {
		return bt, err
	}
	defer db.Close()

	b, err := db.Get([]byte(key))
	if err != nil {
		return bt, err
	}

	_, err = buf.Write(b)
	if err != nil {
		return bt, err
	}

	err = dec.Decode(&bt)
	return bt, err
}

// Returns all the buckets' keys.
func (a Archive) Keys() (chan []byte, error) {
	db, err := bitcask.Open(string(a))
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return db.Keys(), nil
}
