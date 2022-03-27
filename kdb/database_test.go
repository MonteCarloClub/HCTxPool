package kdb

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestNewDb(t *testing.T) {
	file := filepath.Join("/", "tmp", "testdb")
	db, _ := OpenDatabase(file)
	var a = []byte("hello boy")
	var b = []byte("hi girl")
	db.Put(a, b)
	bs, _ := db.Get(a)
	s := string(bs[:])
	fmt.Println(s)
	db.Close()
}