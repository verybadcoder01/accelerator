package models

import "time"

type History struct {
	Id int `json:"-"`
}

type Person struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Fathername string `json:"fathername"`
	BioInfo    string `json:"bioInfo"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

type Owner struct {
	Per  Person    `json:"person"`
	Hist []History `json:"history"`
}

type StatisticMeasure struct {
	Id          int       `json:"-"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartPeriod time.Time `json:"startPeriod"`
	EndPeriod   time.Time `json:"endPeriod"`
	Value       float32   `json:"value"`
}

type Price struct {
	Id       int    `json:"-"`
	LowEnd   int    `json:"lowEnd"`
	HighEnd  int    `json:"highEnd"`
	Currency string `json:"currency"`
}

type Product struct {
	Id          int    `json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       Price  `json:"price"`
}

type ContactType int

const (
	_PHONE = iota
	_EMAIL
	_TELEGRAM
	_WHATSAPP
	_MAIL
	_OTHER
)

type Contact struct {
	typeof ContactType
	link   string
}

type Brand struct {
	Id          int                `json:"-"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Location    string             `json:"location"`
	IsOpen      bool               `json:"-"`
	Owners      []Owner            `json:"owners"`
	Contacts    []Contact          `json:"contacts"`
	Statistics  []StatisticMeasure `json:"statistics"`
	Products    []Product          `json:"products"`
}
