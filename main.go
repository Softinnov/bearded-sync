package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// init vars
	var (
		dbip    = flag.String("db", "db", "database ip")
		dbport  = flag.String("port", "3306", "database port")
		tickets = flag.String("tickets", "/tickets", "tickets folder")
		conf    = flag.String("conf", "", "config file")
		err     error
	)
	dbip, dbport, tickets, err = initvars(conf, dbip, dbport, tickets)
	if err != nil {
		log.Fatal(err.Error())
	}
	// init db
	db, err := initdb("admin:admin@(" + *dbip + ":" + *dbport + ")/prod?parseTime=true&multiStatements=true")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	// main loop
	for {
		manage(tickets, db)
		time.Sleep(time.Second * 5)
	}

}

func manage(tickets *string, db *sql.DB) {
	// on va lire le dosser des tickets
	entries, err := ioutil.ReadDir(*tickets)
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".sql") {

			s := filepath.Join(*tickets, name)
			tmp, err := ioutil.ReadFile(s)
			sql := string(tmp)
			transaction, err := db.Begin()
			if err != nil {
				log.Printf("Can't begin transaction, got : %v", err.Error())
				break
			}
			_, err = transaction.Exec(sql)
			if err != nil {
				transaction.Rollback()
				log.Printf("Can't execute sql file %v, got : %v", s, err.Error())

				err = renamefileKO(*tickets, name)
				if err != nil {
					log.Fatalf("Can't rename the file %v while rollbacking, got : %v", name, err.Error())
				}

				break
			}
			err = transaction.Commit()
			if err != nil {
				log.Printf("Can't commit transaction, got : %v", err.Error())
				break
			}
			err = deletefileOK(*tickets, name)
			if err != nil {
				log.Fatalf("Can't delete the file %v, got : %v", name, err.Error())
			}
		}
	}
}

// deletefileOK supprime le fichier pour qu'il ne soit plus retraité
func deletefileOK(folder, file string) error {
	return os.Remove(filepath.Join(folder, file))
}

// renamefileKO renomme le fichier pour qu'il ne soit plus retraité mais flagué en erreur
func renamefileKO(folder, file string) error {
	s := filepath.Join(folder, file)
	d := filepath.Join(folder, ".err_"+file)
	return os.Rename(s, d)
}

// initdb ouvre la connexion mysql
func initdb(url string) (db *sql.DB, err error) {
	db, err = sql.Open("mysql", url)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// initvars fixe les valeurs à partir d'un fichier de config ou des paramètres ou par défaut
func initvars(conf, dbip, dbport, tickets *string) (i, p, t *string, e error) {
	flag.Parse()
	if *conf != "" {
		f, e := os.Open(*conf)
		if e != nil {
			return nil, nil, nil, e
		}
		defer f.Close()
		c := struct {
			Address string
			Port    string
			Tickets string
		}{}
		e = json.NewDecoder(f).Decode(&c)
		if e != nil {
			return nil, nil, nil, e
		}
		dbip = &c.Address
		dbport = &c.Port
		tickets = &c.Tickets
	}
	return dbip, dbport, tickets, nil
}
