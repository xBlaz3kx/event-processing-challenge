package casino

import (
	"time"

	"gorm.io/gorm"
)

type Player struct {
	gorm.Model
	Email          string    `json:"email"`
	LastSignedInAt time.Time `json:"last_signed_in_at"`
}

func (p Player) IsZero() bool {
	return p.Email == "" || p.LastSignedInAt.IsZero()
}
