package main

import (
	"database/sql"
	"errors"
	"fmt"
)

type Room struct {
	ID         int     `json:"id"`
	Number     int     `json:"number"`
	Parameters string  `json:"params"`
	Beds       int     `json:"beds"`
	Guests     []Guest `json:"guests,omitempty"`
}

func (r *Room) getRoom(db *sql.DB) error {
	return db.QueryRow("SELECT number, params, beds FROM rooms WHERE id=$1",
		r.ID).Scan(&r.Number, &r.Parameters, &r.Beds)
}

func (r *Room) updateRoom(db *sql.DB) error {
	_, err := db.Exec("UPDATE rooms SET number=$1, params=$2, beds=$3 WHERE id=$4",
		r.Number, r.Parameters, r.Beds, r.ID)
	return err
}

func (r *Room) deleteRoom(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM rooms WHERE id=$1", r.ID)
	return err
}

func (r *Room) createRoom(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO rooms(number, params, beds) VALUES($1, $2, $3) RETURNING id",
		r.Number, r.Parameters, r.Beds).Scan(&r.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r *Room) GetGuests(db *sql.DB) error {
	rows, err := db.Query(
		"SELECT id, name, passport FROM guests WHERE room_id=$1", r.ID)

	if err != nil {
		return err
	}

	defer rows.Close()

	guests := []Guest{}

	for rows.Next() {
		var g Guest
		if err := rows.Scan(&g.ID, &g.Name, &g.Passport); err != nil {
			return err
		}

		guests = append(guests, g)
	}
	r.Guests = guests
	return nil
}

func GetAllRoomsWithGuests(db *sql.DB) ([]Room, error) {
	rows, err := db.Query(
		"SELECT id, number,  params, beds FROM rooms")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	rooms := []Room{}

	for rows.Next() {
		var r Room
		if err := rows.Scan(&r.ID, &r.Number, &r.Parameters, &r.Beds); err != nil {
			return nil, err
		}
		err = r.GetGuests(db)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, r)
	}

	return rooms, nil

}

type Guest struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Passport string `json:"passport"`
	RoomID   int    `json:"room_id,omitempty"`
}

func (g *Guest) getGuest(db *sql.DB) error {
	return db.QueryRow("SELECT name, passport, room_id FROM guests WHERE id=$1",
		g.ID).Scan(&g.Name, &g.Passport, &g.RoomID)
}

func (g *Guest) updateGuest(db *sql.DB) error {
	_, err := db.Exec("UPDATE guests SET name=$1, passport=$2, room_id=$3 WHERE id=$4",
		g.Name, g.Passport, g.RoomID, g.ID)
	return err
}

func (g *Guest) deleteGuest(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM guests WHERE id=$1", g.ID)
	return err
}

func (g *Guest) createGuest(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO guests(name, passport, room_id) VALUES($1, $2, $3) RETURNING id",
		g.Name, g.Passport, g.RoomID).Scan(&g.ID)

	if err != nil {
		return err
	}
	err = g.checkRoom(db)
	if err != nil {
		return err
	}
	return nil
}

// Checks if room is available for guest
func (g *Guest) checkRoom(db *sql.DB) error {
	room := Room{ID: g.RoomID}
	err := room.getRoom(db)
	if err != nil {
		return errors.New(fmt.Sprintf("Room with ID: %d does not exist", room.ID))

	}
	err = room.GetGuests(db)
	if err != nil {
		return err
	}
	if len(room.Guests) > 0 {
		return errors.New(fmt.Sprintf("Room with ID: %d already occupied", room.ID))
	}
	return nil
}
