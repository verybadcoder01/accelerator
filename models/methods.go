package models

import (
	"strconv"
	"strings"
	"time"
)

func NewOwner(per *Person) Owner {
	return Owner{Per: *per, Hist: []History{}}
}

func NewContactType(s string) ContactType {
	switch strings.ToLower(s) {
	case "phone":
		return _PHONE
	case "email":
		return _EMAIL
	case "telegram":
		return _TELEGRAM
	case "whatsapp":
		return _WHATSAPP
	case "mail":
		return _MAIL
	case "other":
		return _OTHER
	default:
		return _OTHER
	}
}

func NewContact(typeof string, link string) Contact {
	return Contact{TypeOf: NewContactType(typeof), Link: link}
}

func (b *Brand) GetId() int {
	return b.Id
}

func (b *Brand) AppendProduct(p ...Product) {
	b.Products = append(b.Products, p...)
}

func (b *Brand) AppendOwner(o ...Owner) {
	b.Owners = append(b.Owners, o...)
}

func (b *Brand) AppendStat(s ...StatisticMeasure) {
	b.Statistics = append(b.Statistics, s...)
}
func (b *Brand) AppendContact(c ...Contact) {
	b.Contacts = append(b.Contacts, c...)
}

func (b *Brand) GetBulkInsertStatementContacts(core string) string {
	for i := range b.Contacts {
		core += " ("
		core += "'" + b.Contacts[i].TypeOf.String() + "'"
		core += ", "
		core += "'" + b.Contacts[i].Link + "'"
		core += "),"
	}
	if len(b.Contacts) > 0 {
		core = core[:len(core)-1]
	}
	return core
}

func (b *Brand) GetBulkInsertStatementOwners(core string) string { // core is 'INSERT INTO ... VALUES'
	for i := range b.Owners {
		core += " ("
		core += "'" + b.Owners[i].Per.Name + "'"
		core += ", "
		core += "'" + b.Owners[i].Per.Surname + "'"
		core += ", "
		core += "'" + b.Owners[i].Per.Fathername + "'"
		core += ", "
		core += "'" + b.Owners[i].Per.BioInfo + "'"
		core += "),"
	}
	if len(b.Owners) > 0 {
		core = core[:len(core)-1]
	}
	return core
}

func (b *Brand) GetBulkInsertStatementStatistics(core string) string {
	for i := range b.Statistics {
		core += " ("
		core += "'" + strings.Split(b.Statistics[i].StartPeriod.Format(time.RFC3339), "T")[0] + "'" // RFC3339 is practically ISO 8601 + T<time>, and I only need ISO 8601
		core += ", "
		core += "'" + strings.Split(b.Statistics[i].StartPeriod.Format(time.RFC3339), "T")[0] + "'"
		core += ", "
		core += "'" + b.Statistics[i].Name + "'"
		core += ", "
		core += "'" + b.Statistics[i].Description + "'"
		core += ", "
		core += strconv.FormatFloat(float64(b.Statistics[i].Value), 'f', -1, 32)
		core += ", "
		core += strconv.Itoa(b.Id)
		core += "),"
	}
	if len(b.Statistics) > 0 {
		core = core[:len(core)-1]
	}
	return core
}

func (b *Brand) GetBulkInsertStatementPrices(core string) string {
	for i := range b.Products {
		core += " ("
		core += strconv.Itoa(b.Products[i].Price.LowEnd)
		core += ", "
		core += strconv.Itoa(b.Products[i].Price.HighEnd)
		core += ", "
		core += "'" + b.Products[i].Price.Currency + "'"
		core += "),"
	}
	if len(b.Products) > 0 {
		core = core[:len(core)-1]
	}
	return core
}

func (b *Brand) GetBulkInsertStatementProducts(core string, priceIds []int) string {
	for i := range b.Products {
		core += " ("
		core += "'" + b.Products[i].Name + "'"
		core += ", "
		core += "'" + b.Products[i].Description + "'"
		core += ", "
		core += strconv.Itoa(priceIds[i])
		core += ", "
		core += strconv.Itoa(b.Id)
		core += "),"
	}
	if len(b.Products) > 0 {
		core = core[:len(core)-1]
	}
	return core
}

func (p *Product) GetPrice() Price {
	return p.Price
}

func (p *Product) SetPrice(pr *Price) {
	p.Price = *pr
}

func (o *Owner) SetPersonData(pr *Person) {
	o.Per = *pr
}

func (c ContactType) String() string {
	switch c {
	case _PHONE:
		return "phone"
	case _EMAIL:
		return "email"
	case _TELEGRAM:
		return "telegram"
	case _WHATSAPP:
		return "whatsapp"
	case _MAIL:
		return "mail"
	case _OTHER:
		return "other"
	default:
		return "other"
	}
}
