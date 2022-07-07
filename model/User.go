package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ID     string `json:"ID"`
	Email  string `json:"Email" gorm:"column:email"`
	Name   string `json:"Name" gorm:"column:name"`
	Locale string `json:"Locale" gorm:"column:locale"`
}
