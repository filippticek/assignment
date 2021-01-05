package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

//Model in table
type device struct {
	ID     int    `json:"id"`
	Status int    `json:"status"`
	Name   string `json:"name"`
}

//Struct for multiple devices
type devices struct {
	Devs []device `json:"devices"`
}

//Struct for database mutex control
type mutexDB struct {
	mutex    *sync.Mutex
	sqliteDB *sql.DB
}

//Handles a PUT request for creating a device
func (md *mutexDB) createDevice(w http.ResponseWriter, r *http.Request) {
	var d device
	//Decode JSON to device struct
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&d)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Synchronize db for write
	md.mutex.Lock()
	defer md.mutex.Unlock()

	log.Println("Creating device")
	//Create INSERT statement
	insert := `INSERT INTO devices(id, status, name) VALUES (?, ?, ?)`
	statement, err := md.sqliteDB.Prepare(insert)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Execute statement and return status code depending on the success
	_, err = statement.Exec(d.ID, d.Status, d.Name)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusConflict)
		return
	}

	//Header and JSON encoding
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(d)
}

//Handles a PUT request for updating a device with /{id}
func (md *mutexDB) updateDevice(w http.ResponseWriter, r *http.Request) {
	var d device
	param := mux.Vars(r)
	id := param["id"]
	//Decode JSON to device struct
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&d)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	//Synchronize db for write
	md.mutex.Lock()
	defer md.mutex.Unlock()

	log.Println("Updating device")
	//Create UPDATE statement
	update := `UPDATE devices SET id=?, status=?, name=? WHERE id=?`
	statement, err := md.sqliteDB.Prepare(update)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Execute statement and return status code depending on the success
	_, err = statement.Exec(d.ID, d.Status, d.Name, id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	//Header and JSON encoding
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(d)
}

//Handles a DELETE request for deleting a device with /{id}
func (md *mutexDB) deleteDevice(w http.ResponseWriter, r *http.Request) {
	param := mux.Vars(r)
	id := param["id"]

	//Synchronize db for write
	md.mutex.Lock()
	defer md.mutex.Unlock()

	log.Println("Deleting device")
	//Create DELETE statement
	del := `DELETE FROM devices WHERE id=?`
	statement, err := md.sqliteDB.Prepare(del)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Execute statement and return status code depending on the success
	_, err = statement.Exec(id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

//Handles a GET request for a device with /{id}
func (md *mutexDB) getDevice(w http.ResponseWriter, r *http.Request) {
	var d device
	param := mux.Vars(r)
	id := param["id"]

	log.Println("Getting device")
	//Create SELECT query and populate device struct
	qSelect := `SELECT id, status, name FROM devices WHERE id=?`
	err := md.sqliteDB.QueryRow(qSelect, id).Scan(&d.ID, &d.Status, &d.Name)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	//Header and JSON encoding
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(d)
}

//Handles a GET request for all devices
func (md *mutexDB) getEveryDevice(w http.ResponseWriter, r *http.Request) {
	log.Println("Getting all devices")
	//Create SELECT query
	qSelect := `SELECT id, status, name FROM devices`
	rows, err := md.sqliteDB.Query(qSelect)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	//Read rows, create device struct and append to devices struct
	var dev = devices{Devs: make([]device, 0)}
	for rows.Next() {
		var d device
		rows.Scan(&d.ID, &d.Status, &d.Name)
		dev.Devs = append(dev.Devs, d)
	}

	//Header and JSON encoding
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dev)
}

func main() {

	os.Remove("devices.db")              //Delete db file for a fresh start
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
	//Create table devices
	statement, err := sqliteDB.Prepare("CREATE TABLE devices (id INTEGER PRIMARY KEY, status INTEGER CHECK (status IN (0,1)), name TEXT)")
	if err != nil {
		log.Fatal(err)
	}
	statement.Exec()

	//Initialize mutex and create mutexDB struct
	var m sync.Mutex
	mutexDB := &mutexDB{mutex: &m, sqliteDB: sqliteDB}

	//Add routes
	router := mux.NewRouter()
	router.HandleFunc("/", mutexDB.getEveryDevice).Methods(http.MethodGet)
	router.HandleFunc("/", mutexDB.createDevice).Methods(http.MethodPut)
	router.HandleFunc("/{id}", mutexDB.updateDevice).Methods(http.MethodPut)
	router.HandleFunc("/{id}", mutexDB.getDevice).Methods(http.MethodGet)
	router.HandleFunc("/{id}", mutexDB.deleteDevice).Methods(http.MethodDelete)

	log.Fatal(http.ListenAndServe(":8080", router))
}
