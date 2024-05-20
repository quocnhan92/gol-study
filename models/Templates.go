package models

import (
	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type Templates struct {
	gorm.Model
	Name      string
	HTML      string
	Data      sql.NullString
	Status    string
	StartDate mysql.NullTime
	EndDate   mysql.NullTime
}
