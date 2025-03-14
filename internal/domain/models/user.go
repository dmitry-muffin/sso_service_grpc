package models

type User struct {
	ID       int64
	Name     string
	Email    string
	PassHash []byte
}
