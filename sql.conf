CREATE TABLE IF NOT EXISTS rooms
(
    id SERIAL,
    number INTEGER NOT NULL UNIQUE,
    params TEXT,
    beds INTEGER,
    CONSTRAINT rooms_pkey PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS guests
(
    id SERIAL,
    name TEXT NOT NULL,
    passport TEXT NOT NULL UNIQUE,
    room_id INTEGER NOT NULL,
    CONSTRAINT guests_pkey PRIMARY KEY(id)
);
