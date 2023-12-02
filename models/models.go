package models

import "time"

type History struct {
	id int
}

type Person struct {
	name       string
	surname    string
	fathername string
	bioInfo    string
}

type Owner struct {
	per  Person
	hist []History
}

type StatisticMeasure struct {
	id          int
	name        string
	description string
	startPeriod time.Time
	endPeriod   time.Time
	value       float32
}

type Price struct {
	id       int
	lowEnd   int
	highEnd  int
	currency string
}

type Product struct {
	id          int
	name        string
	description string
	price       Price
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
	id          int
	name        string
	description string
	location    string
	isOpen      bool
	owners      []Owner
	contacts    []Contact
	statistics  []StatisticMeasure
	products    []Product
}
