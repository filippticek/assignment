package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type device struct {
	ID     int    `json:"id"`
	Status int    `json:"status"`
	Name   string `json:"name"`
}

type devices struct {
	Devs []device `json:"devices"`
}

type mutexDB struct {
	mutex    *sync.Mutex
	sqliteDB *sql.DB
}

func (md *mutexDB) createDevice(w http.ResponseWriter, r *http.Request) {
	var d device
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&d)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	md.mutex.Lock()
	defer md.mutex.Unlock()
	log.Println("Creating device")
	insert := `INSERT INTO devices(id, status, name) VALUES (?, ?, ?)`
	statement, err := md.sqliteDB.Prepare(insert)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = statement.Exec(d.ID, d.Status, d.Name)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(d)
}

func (md *mutexDB) updateDevice(w http.ResponseWriter, r *http.Request) {
	var d device
	param := mux.Vars(r)
	id := param["id"]
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&d)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	md.mutex.Lock()
	defer md.mutex.Unlock()
	log.Println("Updating device")
	update := `UPDATE devices SET id=?, status=?, name=? WHERE id=?`
	statement, err := md.sqliteDB.Prepare(update)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = statement.Exec(d.ID, d.Status, d.Name, id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(d)
}

func (md *mutexDB) deleteDevice(w http.ResponseWriter, r *http.Request) {
	param := mux.Vars(r)
	id := param["id"]

	md.mutex.Lock()
	defer md.mutex.Unlock()
	log.Println("Deleting device")
	del := `DELETE FROM devices WHERE id=?`
	statement, err := md.sqliteDB.Prepare(del)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = statement.Exec(id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (md *mutexDB) getDevice(w http.ResponseWriter, r *http.Request) {
	var d device
	param := mux.Vars(r)
	id := param["id"]

	log.Println("Getting device")
	qSelect := `SELECT id, status, name FROM devices WHERE id=?`
	err := md.sqliteDB.QueryRow(qSelect, id).Scan(&d.ID, &d.Status, &d.Name)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(d)
}

func (md *mutexDB) getEveryDevice(w http.ResponseWriter, r *http.Request) {

	log.Println("Getting all devices")
	qSelect := `SELECT id, status, name FROM devices`
	rows, err := md.sqliteDB.Query(qSelect)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var dev = devices{Devs: make([]device, 0)}
	for rows.Next() {
		var d device
		rows.Scan(&d.ID, &d.Status, &d.Name)
		dev.Devs = append(dev.Devs, d)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dev)
}

func main() {
	fmt.Println("hey")
	os.Remove("devices.db")
	file, err := os.Create("devices.db") // Create SQLite file
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	sqliteDB, err := sql.Open("sqlite3", "./devices.db")
	if err != nil {
		log.Fatal(err)
	}
	defer sqliteDB.Close()
	statement, err := sqliteDB.Prepare("CREATE TABLE devices (id INTEGER PRIMARY KEY, status INTEGER CHECK (status IN (0,1)), name TEXT)")
	if err != nil {
		log.Fatal(err)
	}
	statement.Exec()
	var m sync.Mutex
	mutexDB := &mutexDB{mutex: &m, sqliteDB: sqliteDB}

	router := mux.NewRouter()
	router.HandleFunc("/", mutexDB.getEveryDevice).Methods(http.MethodGet)
	router.HandleFunc("/", mutexDB.createDevice).Methods(http.MethodPut)
	router.HandleFunc("/{id}", mutexDB.updateDevice).Methods(http.MethodPut)
	router.HandleFunc("/{id}", mutexDB.getDevice).Methods(http.MethodGet)
	router.HandleFunc("/{id}", mutexDB.deleteDevice).Methods(http.MethodDelete)

	http.ListenAndServe(":8080", router)

}
