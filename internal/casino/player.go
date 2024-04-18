package casino

import (
	"time"
)

type Player struct {
	ID             int       `json:"id" gorm:"column:id;primary_key"`
	Email          string    `json:"email" gorm:"column:email"`
	LastSignedInAt time.Time `json:"last_signed_in_at" gorm:"column:last_signed_in_at"`
}

func (Player) TableName() string {
	return "players"
}

func (p Player) IsZero() bool {
	return p.Email == "" || p.LastSignedInAt.IsZero()
}
