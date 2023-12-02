package models

import "strings"

func NewOwner(per *Person) Owner {
	return Owner{per: *per, hist: []History{}}
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
	return b.id
}

func (b *Brand) AppendProduct(p ...Product) {
	b.products = append(b.products, p...)
}

func (b *Brand) AppendOwner(o ...Owner) {
	b.owners = append(b.owners, o...)
}

func (b *Brand) AppendStat(s ...StatisticMeasure) {
	b.statistics = append(b.statistics, s...)
}
func (b *Brand) AppendContact(c ...Contact) {
	b.contacts = append(b.contacts, c...)
}

func (p *Product) GetPrice() Price {
	return p.price
}

func (p *Product) SetPrice(pr *Price) {
	p.price = *pr
}

func (o *Owner) SetPersonData(pr *Person) {
	o.per = *pr
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
