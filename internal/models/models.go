package models

type Person struct {
	Name        string
	Surname     string
	Pantronymic string
	Age         int    `json:"age"`
	Gender      string `json:"gender"`
	Nationality string `json:"country_id"`
}

type Message interface {
}
