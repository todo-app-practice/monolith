package users

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName string `gorm:"not null" validate:"required"`
	LastName  string `gorm:"not null" validate:"required"`
	Email     string `gorm:"not null" validate:"required,email"`
	Password  string `gorm:"not null" validate:"required"`
}

// BeforeSave : hook before a user is saved
func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	fmt.Println("before save")
	fmt.Println(u.Password)
	if u.Password != "" {
		hash, err := hashPassword(u.Password)
		if err != nil {
			return nil
		}

		u.Password = hash
	}

	fmt.Println(u.Password)
	return
}

// BeforeUpdate : hook before a user is updated
func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	fmt.Println("before update")
	fmt.Println(u.Password)
	if u.Password != "" {
		hash, err := hashPassword(u.Password)
		if err != nil {
			return nil
		}

		u.Password = hash
	}

	fmt.Println(u.Password)
	return
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
