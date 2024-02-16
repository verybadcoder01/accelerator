package models

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
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

func (p *Product) GetImages(l *log.Logger) []image.Image {
	res := make([]image.Image, len(p.Media))
	for i, img := range p.Media {
		byteImg := []byte(strings.Split(img, ",")[1])
		l.Debug("attempting to decode image")
		byteArr := make([]byte, len(byteImg))
		// decode from base64
		_, err := base64.StdEncoding.Decode(byteArr, byteImg)
		if err != nil {
			l.Errorln(err)
		}
		// parse image
		curimg, err := png.Decode(bytes.NewReader(byteArr))
		// for now, we only allow png
		if err != nil {
			l.Errorln("decoder err: " + err.Error())
			continue
		}
		res[i] = curimg
	}
	return res
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
		core += "'" + strings.Split(b.Statistics[i].EndPeriod.Format(time.RFC3339), "T")[0] + "'"
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

func (b *Brand) GetBulkUpdateStatementContacts(core string, ids []int) string { // core will be like update ... from (values ?) as ... where ...
	stmt := ""
	for i := range b.Contacts {
		stmt += " ("
		stmt += strconv.Itoa(ids[i])
		stmt += ","
		stmt += "'" + b.Contacts[i].TypeOf.String() + "'"
		stmt += ","
		stmt += "'" + b.Contacts[i].Link + "'"
		stmt += "),"
	}
	if len(b.Contacts) > 0 {
		stmt = stmt[:len(stmt)-1]
	}
	core = strings.Replace(core, "?", stmt, 1)
	return core
}

func (b *Brand) GetBulkUpdateStatementOwners(core string, ids []int) string { // same as in the function above
	stmt := ""
	for i := range b.Owners {
		stmt += " ("
		stmt += strconv.Itoa(ids[i])
		stmt += ","
		stmt += "'" + b.Owners[i].Per.Name + "'"
		stmt += ","
		stmt += "'" + b.Owners[i].Per.Surname + "'"
		stmt += ","
		stmt += "'" + b.Owners[i].Per.Fathername + "'"
		stmt += ","
		stmt += "'" + b.Owners[i].Per.BioInfo + "'"
		stmt += "),"
	}
	if len(b.Owners) > 0 {
		stmt = stmt[:len(stmt)-1]
	}
	core = strings.Replace(core, "?", stmt, 1)
	return core
}

func (b *Brand) GetBulkUpdateStatementPrices(core string, ids []int) string { // same format as in update contacts
	stmt := ""
	for i := range b.Products {
		stmt += " ("
		stmt += strconv.Itoa(ids[i])
		stmt += ","
		stmt += strconv.Itoa(b.Products[i].Price.LowEnd)
		stmt += ","
		stmt += strconv.Itoa(b.Products[i].Price.HighEnd)
		stmt += ","
		stmt += "'" + b.Products[i].Price.Currency + "'"
		stmt += "),"
	}
	if len(b.Products) > 0 {
		stmt = stmt[:len(stmt)-1]
	}
	core = strings.Replace(core, "?", stmt, 1)
	return core
}

func (b *Brand) GetBulkUpdateStatementProducts(core string, ids []int) string { // see format in the previous functions...
	stmt := ""
	for i := range b.Products {
		stmt += " ("
		stmt += strconv.Itoa(ids[i])
		stmt += ","
		stmt += "'" + b.Products[i].Name + "'"
		stmt += ","
		stmt += "'" + b.Products[i].Description + "'"
		stmt += "),"
	}
	if len(b.Products) > 0 {
		stmt = stmt[:len(stmt)-1]
	}
	core = strings.Replace(core, "?", stmt, 1)
	return core
}

func (b *Brand) GetBulkUpdateStatementStats(core string, ids []int) string {
	stmt := ""
	for i := range b.Statistics {
		stmt += " ("
		stmt += strconv.Itoa(ids[i])
		stmt += ","
		stmt += "'" + strings.Split(b.Statistics[i].StartPeriod.Format(time.RFC3339), "T")[0] + "'" // as in insert, RFC3339 is basically ISO 8601 + T<time>, and I need only ISO
		stmt += "::date"                                                                            // because it refuses to work otherwise
		stmt += ","
		stmt += "'" + strings.Split(b.Statistics[i].EndPeriod.Format(time.RFC3339), "T")[0] + "'"
		stmt += "::date"
		stmt += ","
		stmt += "'" + b.Statistics[i].Name + "'"
		stmt += ","
		stmt += "'" + b.Statistics[i].Description + "'"
		stmt += ","
		stmt += strconv.FormatFloat(float64(b.Statistics[i].Value), 'f', -1, 32)
		stmt += "),"
	}
	if len(b.Statistics) > 0 {
		stmt = stmt[:len(stmt)-1]
	}
	core = strings.Replace(core, "?", stmt, 1)
	return core
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
