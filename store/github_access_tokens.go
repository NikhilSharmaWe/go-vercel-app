package store

import (
	"errors"

	"github.com/NikhilSharmaWe/go-vercel-app/models"
	"gorm.io/gorm"
)

type GithubTokenStore interface {
	CreateTable() error
	Create(fr models.GithubTokenDBModel) error
	Update(updateMap, whereMap map[string]interface{}) error
	Delete(whereMap map[string]interface{}) error
	IsExists(whereQuery string, whereArgs ...interface{}) (bool, error)
	DB() *gorm.DB
}

type githubTokenStore struct {
	db *gorm.DB
}

func NewGithubTokenStore(db *gorm.DB) GithubTokenStore {
	return &githubTokenStore{
		db: db,
	}
}

func (us *githubTokenStore) table() string {
	return "github_token"
}

func (us *githubTokenStore) DB() *gorm.DB {
	return us.db
}

func (us *githubTokenStore) CreateTable() error {
	return us.db.Table(us.table()).AutoMigrate(models.GithubTokenDBModel{})
}

func (us *githubTokenStore) Create(fr models.GithubTokenDBModel) error {
	return us.db.Table(us.table()).Create(fr).Error
}

func (us *githubTokenStore) Update(updateMap, whereMap map[string]interface{}) error {
	return us.db.Table(us.table()).Where(whereMap).Updates(updateMap).Error
}

func (us *githubTokenStore) Delete(whereMap map[string]interface{}) error {
	return us.db.Table(us.table()).Where(whereMap).Delete(nil).Error
}

func (us *githubTokenStore) IsExists(whereQuery string, whereArgs ...interface{}) (bool, error) {
	type Res struct {
		IsExists bool
	}

	var res Res

	if err := us.db.Table(us.table()).Select("1 = 1 us is_exists").Where(whereQuery, whereArgs...).Find(&res).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}

		return false, err
	}

	return res.IsExists, nil
}
