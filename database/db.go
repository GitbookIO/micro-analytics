package database

import (
    "database/sql"
)

type Database struct {
    Name        string
    Directory   string
    Conn        *sql.DB
}