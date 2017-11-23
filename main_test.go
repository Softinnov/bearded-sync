package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestDeletefileOK(t *testing.T) {
	// create a temporary file
	tickets, err := ioutil.TempFile("", "tempticket")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tickets.Name())

	err = deletefileOK(path.Dir(tickets.Name()), path.Base(tickets.Name()))
	if err != nil {
		log.Fatal(err)
	}
	if _, err = os.Stat(tickets.Name()); err == nil {
		log.Fatalf("File not removed, got %v", tickets.Name())
	}
}

func TestRenamefileKO(t *testing.T) {
	// create a temporary file
	tickets, err := ioutil.TempFile("", "tempticket")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tickets.Name())
	b := path.Base(tickets.Name())
	d := path.Dir(tickets.Name())

	err = renamefileKO(d, b)
	if err != nil {
		log.Fatal(err)
	}
	exp := filepath.Join(d, ".err_"+b)
	if _, err = os.Stat(exp); os.IsNotExist(err) {
		log.Fatalf("expected file %v exists\n", exp)
	}
	err = os.Remove(exp)
	if err != nil {
		log.Fatalf("Can't remove file %v \n", exp)
	}
}

func TestLoop(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO vente \(id, prod\) values\(1, 123\);`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// create temporary directory
	tickets, err := ioutil.TempDir("", "tic")
	if err != nil {
		t.Fatalf("error was not expected while creating temporary directory: %s", err)
	}
	defer os.RemoveAll(tickets)

	// create fake sql file
	sqlfake := []byte("INSERT INTO vente (id, prod) values(1, 123);")
	sqlfile := tickets + "/tic.sql"
	err = ioutil.WriteFile(sqlfile, sqlfake, 0777)
	if err != nil {
		t.Fatalf("error was not expected while writing fake content to sql file: %s", err)
	}

	// now we execute our method
	manage(&tickets, db)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}
