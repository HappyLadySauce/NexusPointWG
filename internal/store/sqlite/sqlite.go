package sqlite

import (
	"gorm.io/gorm"
	
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
)

type datastore struct {
	db *gorm.DB
}

func (ds *datastore) Users() store.UserStore {
	return newUsers(ds)
}
