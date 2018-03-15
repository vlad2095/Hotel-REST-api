# REST-API-example

How does it configures and run

<p> 1. In PostrgreSQL, create tables for rooms and guests: </p>

<code>
    CREATE TABLE IF NOT EXISTS rooms
(
    id SERIAL,
    number INTEGER NOT NULL UNIQUE,
    params TEXT,
    beds INTEGER,
    CONSTRAINT rooms_pkey PRIMARY KEY(id)
);
</code>


<code>
    CREATE TABLE IF NOT EXISTS guests
(
    id SERIAL,
    name TEXT NOT NULL,
    passport TEXT NOT NULL UNIQUE,
    room_id INTEGER NOT NULL,
    CONSTRAINT guests_pkey PRIMARY KEY(id)
);
</code>

<p>2. Configure PostgreSQL environment variables: </p>

<code>
    export APP_DB_USERNAME=youruser
    export APP_DB_PASSWORD=yourpassword
    export APP_DB_NAME=yourdbname
</code>

<p>3. Next: </p>

<code>go build</code>
<code>./REST-API-example</code>