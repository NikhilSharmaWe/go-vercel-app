package models

type UserDBModel struct {
	Username     string `gorm:"column:username;primaryKey"`
	Email        string `gorm:"column:email"`
	GithubAccess bool   `gorm:"column:github_access"`
}
