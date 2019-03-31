package main

import (
	"log"
	"time"

	bolt "go.etcd.io/bbolt"
)

var bucketName = []byte("foobar")
var db *bolt.DB

func init() {
	var err error
	db, err = bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalln(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	if err != nil {
		log.Fatalln(err)
	}
}

func DBGet(k string) (string, error) {
	var ret []byte
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		ret = bucket.Get([]byte(k))
		return nil
	})
	return string(ret), err
}

func DBSet(k, v string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		return bucket.Put([]byte(k), []byte(v))
	})
}

func DBDel(k string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		return bucket.Delete([]byte(k))
	})
}
