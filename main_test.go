package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var a App

func TestMain(m *testing.M) {
	a = App{}
	a.Initialize(
		os.Getenv("TEST_DB_USERNAME"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_NAME"))

	ensureTableExistsGuests()
	ensureTableExistsRooms()

	code := m.Run()

	clearTableGuests()
	clearTableRooms()

	os.Exit(code)
}

func TestEmptyTableRooms(t *testing.T) {
	clearTableRooms()

	req, _ := http.NewRequest("GET", "/rooms", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestEmptyTableGuests(t *testing.T) {
	clearTableGuests()

	req, _ := http.NewRequest("GET", "/guests", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentRoom(t *testing.T) {
	clearTableRooms()

	req, _ := http.NewRequest("GET", "/room/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Room not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Room not found'. Got '%s'", m["error"])
	}
}

func TestGetNonExistentGuest(t *testing.T) {
	clearTableGuests()

	req, _ := http.NewRequest("GET", "/guest/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Guest not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Guest not found'. Got '%s'", m["error"])
	}
}

func TestCreateRoom(t *testing.T) {
	clearTableRooms()

	payload := []byte(`{"number":15, "params":"fine", "beds":1}`)

	req, _ := http.NewRequest("POST", "/room", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["params"] != "fine" {
		t.Errorf("Expected params to be 'fine'. Got '%v'", m["params"])
	}

	if m["number"] != 15.0 {
		t.Errorf("Expected number to be '15'. Got '%v'", m["number"])
	}

	// the id and beds is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["beds"] != 1.0 {
		t.Errorf("Expected beds to be '1'. Got '%v'", m["beds"])
	}

	if m["id"] != 1.0 {
		t.Errorf("Expected room ID to be '1'. Got '%v'", m["id"])
	}
}

func TestCreateGuest(t *testing.T) {
	clearTableRooms()
	clearTableGuests()
	addRoom()

	payload := []byte(`{"name":"Nastya", "passport":"7785DF", "room_id":1}`)

	req, _ := http.NewRequest("POST", "/guest", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "Nastya" {
		t.Errorf("Expected guest name to be 'Nastya'. Got '%v'", m["name"])
	}

	if m["passport"] != "7785DF" {
		t.Errorf("Expected guest passport to be '7785DF'. Got '%v'", m["passport"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["room_id"] != 1.0 {
		t.Errorf("Expected guest ID to be '1'. Got '%v'", m["room_id"])
	}

}

func TestCreateGuestWithRoomOccupied(t *testing.T) {
	clearTableRooms()
	clearTableGuests()
	addRoom()
	addGuest()

	payload := []byte(`{"name":"Sara", "passport":"9985DF", "room_id":1}`)

	req, _ := http.NewRequest("POST", "/guest", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusInternalServerError, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Room with ID: 1 already occupied" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Room with ID: 1 already occupied'. Got '%s'", m["error"])
	}
}

func TestGetRoom(t *testing.T) {
	clearTableRooms()
	addRoom()

	req, _ := http.NewRequest("GET", "/room/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestGetGuest(t *testing.T) {
	clearTableGuests()
	addGuest()

	req, _ := http.NewRequest("GET", "/guest/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateRoom(t *testing.T) {
	clearTableRooms()
	addRoom()

	req, _ := http.NewRequest("GET", "/room/1", nil)
	response := executeRequest(req)
	var originalRoom map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalRoom)

	payload := []byte(`{"number":20, "params":"very good(actually not :) )"}`)

	req, _ = http.NewRequest("PUT", "/room/1", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalRoom["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalRoom["id"], m["id"])
	}

	if m["number"] == originalRoom["number"] {
		t.Errorf("Expected the number to change from '%v' to '%v'. Got '%v'", originalRoom["number"], m["number"], m["number"])
	}

	if m["params"] == originalRoom["params"] {
		t.Errorf("Expected the params to change from '%v' to '%v'. Got '%v'", originalRoom["params"], m["params"], m["params"])
	}
}

func TestUpdateGuest(t *testing.T) {
	clearTableRooms()
	addRoom()

	clearTableGuests()
	addGuest()

	req, _ := http.NewRequest("GET", "/guest/1", nil)
	response := executeRequest(req)
	var originalGuest map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalGuest)

	payload := []byte(`{"name":"Dan", "passport":"8870465", "room_id":1}`)

	req, _ = http.NewRequest("PUT", "/guest/1", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalGuest["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalGuest["id"], m["id"])
	}

	if m["name"] == originalGuest["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalGuest["name"], m["name"], m["name"])
	}

	if m["passport"] == originalGuest["passport"] {
		t.Errorf("Expected the passport to change from '%v' to '%v'. Got '%v'", originalGuest["passport"], m["passport"], m["passport"])
	}

	if m["room_id"] != originalGuest["room_id"] {
		t.Errorf("Expected the room_id to remain the same (%v). Got %v", originalGuest["room_id"], m["room_id"])
	}
}

func TestDeleteRoom(t *testing.T) {
	clearTableRooms()
	addRoom()

	req, _ := http.NewRequest("GET", "/room/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/room/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/room/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestDeleteGuest(t *testing.T) {
	clearTableGuests()
	addGuest()

	req, _ := http.NewRequest("GET", "/guest/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/guest/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/guest/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestGetAllRoomsWithGuests(t *testing.T) {
	clearTableGuests()
	clearTableRooms()

	payloadRoom := []byte(`{"number":20, "params":"bad", "beds":1}`)

	reqRoom, _ := http.NewRequest("POST", "/room", bytes.NewBuffer(payloadRoom))
	responseR := executeRequest(reqRoom)

	checkResponseCode(t, http.StatusCreated, responseR.Code)

	payloadGuest := []byte(`{"name":"James", "passport":"9345ZZ", "room_id":1}`)

	reqGuest, _ := http.NewRequest("POST", "/guest", bytes.NewBuffer(payloadGuest))
	responseG := executeRequest(reqGuest)

	checkResponseCode(t, http.StatusCreated, responseG.Code)

	req, _ := http.NewRequest("GET", "/rooms", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var z []Room
	json.Unmarshal(response.Body.Bytes(), &z)

	room := z[0]

	if room.Parameters != "bad" {
		t.Errorf("Expected params to be 'bad'. Got '%v'", room.Parameters)
	}

	if room.Number != 20.0 {
		t.Errorf("Expected number to be '20'. Got '%v'", room.Number)
	}

	if room.Beds != 1.0 {
		t.Errorf("Expected beds to be '1'. Got '%v'", room.Beds)
	}

	if room.ID != 1.0 {
		t.Errorf("Expected room ID to be '1'. Got '%v'", room.ID)
	}

	g := room.Guests[0]

	if g.Name != "James" {
		t.Errorf("Expected guest name to be 'James'. Got '%v'", g.Name)
	}

	if g.Passport != "9345ZZ" {
		t.Errorf("Expected guest passport to be '9345ZZ'. Got '%v'", g.Passport)
	}

}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func ensureTableExistsRooms() {
	if _, err := a.DB.Exec(tableCreationQueryRooms); err != nil {
		log.Fatal(err)
	}
}

func ensureTableExistsGuests() {
	if _, err := a.DB.Exec(tableCreationQueryGuests); err != nil {
		log.Fatal(err)
	}
}

func clearTableRooms() {
	a.DB.Exec("DELETE FROM rooms")
	a.DB.Exec("ALTER SEQUENCE rooms_id_seq RESTART WITH 1")
}

func clearTableGuests() {
	a.DB.Exec("DELETE FROM guests")
	a.DB.Exec("ALTER SEQUENCE guests_id_seq RESTART WITH 1")
}

func addRoom() {
	a.DB.Exec("INSERT INTO rooms(number, params, beds) VALUES($1, $2, $3)", 1, "five stars", 2)
}

func addGuest() {
	a.DB.Exec("INSERT INTO guests(name, passport, room_id) VALUES($1, $2, $3)", "John", "ZZ178567", 1)
}

const tableCreationQueryRooms = `CREATE TABLE IF NOT EXISTS rooms
(
    id SERIAL,
    number INTEGER NOT NULL UNIQUE,
    params TEXT,
    beds INTEGER,
    CONSTRAINT rooms_pkey PRIMARY KEY(id)
);`

const tableCreationQueryGuests = `CREATE TABLE IF NOT EXISTS guests
(
    id SERIAL,
    name TEXT NOT NULL,
    passport TEXT NOT NULL UNIQUE,
    room_id INTEGER NOT NULL,
    CONSTRAINT guests_pkey PRIMARY KEY(id)
);`
