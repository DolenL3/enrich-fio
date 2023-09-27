package models

import "github.com/google/uuid"

// FilterConfig is config with filters, that should be applyed during storage search.
type FilterConfig struct {
	ID          uuid.UUID
	Name        string
	Surname     string
	Patronymic  string
	Age         FilterAge
	Gender      Gender
	Nationality string
}

// FilterAge is filter for age.
type FilterAge struct {
	Min int
	Max int
}
