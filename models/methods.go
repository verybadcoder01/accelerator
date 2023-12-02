package models

import "strings"

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
	return Contact{typeof: NewContactType(typeof), link: link}
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
