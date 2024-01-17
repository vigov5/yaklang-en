package vulinbox

import (
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

type Session struct {
	gorm.Model
	Uuid     string
	Username string
	Role     string
}

func (v *VulinUser) CreateSession(dbm *dbm) (session Session, err error) {
	session = Session{
		Uuid:     uuid.NewV4().String(),
		Username: v.Username,
		Role:     v.Role,
	}
	// Insert session records into database
	if err := dbm.db.Create(&session).Error; err != nil {
		return session, err
	}

	return session, nil
}
