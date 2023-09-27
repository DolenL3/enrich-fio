package models

import "github.com/google/uuid"

// ChangeConfig is config with changes, which should be applyed to storage.
type ChangeConfig struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Surname     string    `json:"surname"`
	Patronymic  string    `json:"patronymic"`
	Age         int       `json:"age"`
	Gender      Gender    `json:"gender"`
	Nationality string    `json:"nationality"`
}
