package models

import "github.com/google/uuid"

// Person a result of service's buisness logic.
type Person struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Surname     string    `json:"surname"`
	Patronymic  string    `json:"patronymic"`
	Age         int       `json:"age"`
	Gender      Gender    `json:"gender"`
	Nationality string    `json:"nationality"`
}

// Gender is a type for gender value in a Person.
type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)
