package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// sets up the database connection and routes for the app
func (a *App) Initialize(user, password, dbname string) {
	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s", user, password, dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

// Run starts the app and serves on the specified addr
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(":8000", a.Router))

}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/rooms", a.getRooms).Methods("GET")
	a.Router.HandleFunc("/room", a.createRoom).Methods("POST")
	a.Router.HandleFunc("/room/{id:[0-9]+}", a.getRoom).Methods("GET")
	a.Router.HandleFunc("/room/{id:[0-9]+}", a.updateRoom).Methods("PUT")
	a.Router.HandleFunc("/room/{id:[0-9]+}", a.deleteRoom).Methods("DELETE")

	a.Router.HandleFunc("/guests", a.getGuests).Methods("GET")
	a.Router.HandleFunc("/guest", a.createGuest).Methods("POST")
	a.Router.HandleFunc("/guest/{id:[0-9]+}", a.getGuest).Methods("GET")
	a.Router.HandleFunc("/guest/{id:[0-9]+}", a.updateGuest).Methods("PUT")
	a.Router.HandleFunc("/guest/{id:[0-9]+}", a.deleteGuest).Methods("DELETE")
}

// *** ROOMS ***//

func (a *App) getRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := GetAllRoomsWithGuests(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, rooms)
}

func (a *App) createRoom(w http.ResponseWriter, r *http.Request) {
	var room Room
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&room); err != nil {
		fmt.Println(err.Error())
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := room.createRoom(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, room)
}

func (a *App) getRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	room := Room{ID: id}
	if err := room.getRoom(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Room not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, room)
}

func (a *App) updateRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	var room Room
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&room); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	room.ID = id

	if err := room.updateRoom(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, room)
}

func (a *App) deleteRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	room := Room{ID: id}
	if err := room.deleteRoom(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// *** GUESTS ***//

func (a *App) getGuests(w http.ResponseWriter, r *http.Request) {
	guests, err := GetAllGuests(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, guests)
}

func (a *App) createGuest(w http.ResponseWriter, r *http.Request) {
	var g Guest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&g); err != nil {
		fmt.Println(err.Error())
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := g.createGuest(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, g)
}

func (a *App) getGuest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid guest ID")
		return
	}

	g := Guest{ID: id}
	if err := g.getGuest(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Guest not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, g)
}

func (a *App) updateGuest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid guest ID")
		return
	}

	var g Guest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&g); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	g.ID = id

	if err := g.updateGuest(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, g)
}

func (a *App) deleteGuest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid guest ID")
		return
	}

	g := Guest{ID: id}
	if err := g.deleteGuest(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
