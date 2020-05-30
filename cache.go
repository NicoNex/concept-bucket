package main

import (
	"bytes"
	"encoding/gob"
	"strconv"

	"github.com/prologic/bitcask"
)

type Cache string

// TODO: add delete method.

// Returns the bytes representation of the chatId.
func itob(i int64) []byte {
	return []byte(strconv.FormatInt(i, 10))
}

// Put saves a list of bucket tokens associated to a provided chatId.
func (c Cache) Put(id int64, data []string) error {
	var buf bytes.Buffer
	var enc = gob.NewEncoder(&buf)

	err := enc.Encode(data)
	if err != nil {
		return err
	}

	db, err := bitcask.Open(string(c))
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Put(itob(id), buf.Bytes())
}

// Get retrieves the list of bucket tokens associated to the given chatId.
func (c Cache) Get(id int64) ([]string, error) {
	var buf bytes.Buffer
	var ret []string
	var dec = gob.NewDecoder(&buf)

	db, err := bitcask.Open(string(c))
	if err != nil {
		return ret, err
	}
	defer db.Close()

	b, err := db.Get(itob(id))
	if err != nil {
		return ret, err
	}

	_, err = buf.Write(b)
	if err != nil {
		return ret, err
	}

	err = dec.Decode(&ret)
	return ret, err
}

// Returns all the cached chatId.
func (c Cache) Keys() (chan []byte, error) {
	db, err := bitcask.Open(string(c))
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return db.Keys(), nil
}
