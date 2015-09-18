// Copyright 2015 Johnnie Kearse III
// Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/codegangsta/cli"
)

// main() inintalizes cli and kicks it off
func main() {
	a := cli.NewApp()
	a.Name = appName
	a.Usage = appUsage
	a.Version = appVersion
	a.Author = appAuthor
	a.Email = appEmail
	a.Commands = appCommands

	a.Run(os.Args)

}

// set application variables for main()
var (
	appName         = "todocli"
	appUsage        = "Keep track of things you can't remember. What else?"
	appVersion      = "0.1.0alpha"
	appAuthor       = "Johnnie Kearse III"
	appEmail        = "jkearse3@gmail.com"
	appCommands     = appMainCommands
	appMainCommands = []cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "Add an item.",
			Action:  addItem,
		},
		{
			Name:    "remove",
			Aliases: []string{"r"},
			Usage:   "Remove an item.",
			Action:  removeItem,
		},
		{
			Name:    "show",
			Aliases: []string{"s"},
			Usage:   "Show all items",
			Action:  showItems,
		},
	}
)

func addItem(c *cli.Context) {
	if len(c.Args()) == 0 {
		fmt.Printf("please add text for your todo-list item.\n")
		return
	}
	db, err := openDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	for _, v := range c.Args() {
		id, err := generateID(db)
		if err != nil {
			log.Fatal(err)
		}
		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("todos"))
			err := b.Put(id, []byte(v))
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("added item: todo number %s - %s\n", id, v)
	}
}

func removeItem(c *cli.Context) {
	if len(c.Args()) == 0 {
		fmt.Printf("please select a todolist item.\n")
		return
	}
	db, err := openDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	for _, v := range c.Args() {
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("todos"))
			err := b.Delete([]byte(v))
			if err != nil {
				return err
			}
			return nil
		})
		fmt.Printf("removed item: %s\n", v)
	}
}

func showItems(c *cli.Context) {
	db, err := openDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("todos"))
		fmt.Printf("Showing items in the list:\n")
		err := b.ForEach(func(k, v []byte) error {
			fmt.Printf("id# %s: %s\n", k, v)
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})
}

func openDB() (db *bolt.DB, err error) {
	db, err = bolt.Open("todocli.db", 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Batch(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("todos"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("idgen"))
		return nil
	})
	return db, nil
}

func generateID(db *bolt.DB) ([]byte, error) {
	var id []byte
	err := db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("idgen"))
		if b == nil {
			log.Fatal(fmt.Errorf("Empty idgen"))
		}
		cid := b.Get([]byte("id"))
		if cid == nil {
			// If no ID, create ID starting at 1
			err := b.Put([]byte("id"), []byte("1"))
			if err != nil {
				return err
			}
			cid = b.Get([]byte("id"))
		} else {
			// Increment ID by 1 if it already exists
			cidInt, err := strconv.Atoi(string(cid))
			if err != nil {
				return err
			}
			cidInt = cidInt + 1
			cid = []byte(strconv.Itoa(cidInt))
			b.Put([]byte("id"), cid)
		}
		id = cid
		return nil
	})
	if err != nil {
		return nil, err
	}
	return id, nil
}
