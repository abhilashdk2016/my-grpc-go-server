package database

import (
	"time"

	"github.com/google/uuid"
)

type DummyOrm struct {
	UserId    uuid.UUID `gorm:"primary_key"`
	UserName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (DummyOrm) TableName() string {
	return "dummy"
}
