package db

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"accelerator/models"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const BcryptCost = 5

type Database struct {
	db  *sql.DB
	log *log.Logger
}

func NewDb(dsn string, log *log.Logger) Database {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Errorln(err)
		return Database{}
	}
	return Database{db: conn, log: log}
}

func (d *Database) CreateTables() error {
	createBrand := `CREATE TABLE IF NOT EXISTS Brands(id SERIAL PRIMARY KEY, name VARCHAR(50), DESCRIPTION VARCHAR(200), city VARCHAR(50), is_open BOOLEAN)`
	createStats := `CREATE TABLE IF NOT EXISTS Statistics(id SERIAL PRIMARY KEY, start_time DATE, end_time DATE, name VARCHAR(50), description VARCHAR(200), value NUMERIC, brand_id INT, CONSTRAINT fk_statistics FOREIGN KEY(brand_id) REFERENCES Brands(id) ON DELETE CASCADE)`
	createPrices := `CREATE TABLE IF NOT EXISTS Prices(id SERIAL PRIMARY KEY, low_end INT, high_end INT, currency VARCHAR(20))`
	createProducts := `CREATE TABLE IF NOT EXISTS Products(id SERIAL PRIMARY KEY, name VARCHAR(50) UNIQUE, description VARCHAR(50), price_id INT, brand_id INT, CONSTRAINT fk_price_product FOREIGN KEY(price_id) REFERENCES Prices(id) ON DELETE CASCADE, CONSTRAINT fk_brand_product FOREIGN KEY(brand_id) REFERENCES Brands(id) ON DELETE CASCADE)`
	createContacts := `CREATE TABLE IF NOT EXISTS Contacts(id SERIAL PRIMARY KEY, type VARCHAR(20), contact VARCHAR(100))`
	createLinkCBrand := `CREATE TABLE IF NOT EXISTS L_Brand_Contacts(id SERIAL PRIMARY KEY, brand_id INT, contact_id INT, CONSTRAINT fk_link_contacts FOREIGN KEY(brand_id) REFERENCES Brands(id) ON DELETE CASCADE, CONSTRAINT fk_contact_id FOREIGN KEY(contact_id) REFERENCES Contacts(id) ON DELETE CASCADE)`
	createHistory := `CREATE TABLE IF NOT EXISTS History(id SERIAL PRIMARY KEY)` // TBD LATER
	createOwners := `CREATE TABLE IF NOT EXISTS Owners(id SERIAL PRIMARY KEY, name VARCHAR(20), surname VARCHAR(50), fathername VARCHAR(50), bio_info VARCHAR(200), history_id INT, CONSTRAINT fk_history_id FOREIGN KEY (history_id) REFERENCES History(id))`
	createLOBrand := `CREATE TABLE IF NOT EXISTS L_Brand_Owners(id SERIAL PRIMARY KEY, brand_id INT, owner_id INT, CONSTRAINT fk_link_owners FOREIGN KEY(brand_id) REFERENCES Brands(id) ON DELETE CASCADE, CONSTRAINT fk_owner_id FOREIGN KEY(owner_id) REFERENCES Owners(id) ON DELETE CASCADE)`
	createUsers := `CREATE TABLE IF NOT EXISTS users(id SERIAL PRIMARY KEY, email VARCHAR(50) UNIQUE, password VARCHAR(200), name VARCHAR(50), surname VARCHAR(50))`
	execList := []string{
		createBrand, createStats, createPrices, createProducts, createContacts, createLinkCBrand, createHistory,
		createOwners,
		createLOBrand,
		createUsers,
	}
	for _, st := range execList {
		_, err := d.db.Exec(st)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Database) getBrandOwners(brandId int) ([]models.Owner, error) {
	getOwnerIds := `SELECT owner_id FROM l_brand_owners WHERE brand_id = $1`
	getOwnerById := `SELECT name, surname, fathername, bio_info FROM owners WHERE id = $1` // email is not meant to be publicly visible
	rows, err := d.db.Query(getOwnerIds, brandId)
	if err != nil {
		return nil, err
	}
	var res []models.Owner
	for rows.Next() {
		id := 0
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		owner, err := d.db.Query(getOwnerById, id)
		if err != nil {
			return nil, err
		}
		cur := models.Person{}
		for owner.Next() {
			err = owner.Scan(&cur.Name, &cur.Surname, &cur.Fathername, &cur.BioInfo)
			if err != nil {
				return nil, err
			}
		}
		err = owner.Close()
		if err != nil {
			return nil, err
		}
		res = append(res, models.NewOwner(&cur))
	}
	err = rows.Close()
	if err != nil {
		return res, err
	}
	return res, nil
}

func (d *Database) getBrandContacts(brandId int) ([]models.Contact, error) {
	getContactIds := `SELECT contact_id FROM l_brand_contacts WHERE brand_id = $1`
	getContactById := `SELECT type, contact FROM contacts WHERE id = $1`
	rows, err := d.db.Query(getContactIds, brandId)
	if err != nil {
		return nil, err
	}
	var res []models.Contact
	for rows.Next() {
		id := 0
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		contact, err := d.db.Query(getContactById, id)
		if err != nil {
			return nil, err
		}
		var curtype string
		var curlink string
		for contact.Next() {
			err = contact.Scan(&curtype, &curlink)
			if err != nil {
				return nil, err
			}
		}
		err = contact.Close()
		if err != nil {
			return nil, err
		}
		res = append(res, models.NewContact(curtype, curlink))
	}
	err = rows.Close()
	if err != nil {
		return res, err
	}
	return res, nil
}

func (d *Database) GetOpenBrands() ([]models.Brand, error) {
	getOpen := `SELECT id, name, description, city, is_open FROM brands WHERE is_open=TRUE`
	rows, err := d.db.Query(getOpen)
	if err != nil {
		return nil, err
	}
	var open []models.Brand
	for rows.Next() {
		cur := models.Brand{}
		err = rows.Scan(&cur.Id, &cur.Name, &cur.Description, &cur.Location, &cur.IsOpen)
		if err != nil {
			return nil, err
		}
		open = append(open, cur)
	}
	d.log.Debug("scanned core info")
	err = rows.Close()
	if err != nil {
		return nil, err
	}
	getProducts := `SELECT id, name, description, price_id FROM products WHERE brand_id = $1`
	getStats := `SELECT id, name, description, start_time, end_time, value FROM statistics WHERE brand_id = $1`
	getPrice := `SELECT low_end, high_end, currency FROM prices WHERE id = $1`
	for i, brand := range open {
		// search for products
		rows, err = d.db.Query(getProducts, brand.GetId())
		if err != nil {
			return open, err
		}
		curPrice := models.Price{}
		for rows.Next() {
			p := models.Product{}
			err = rows.Scan(&p.Id, &p.Name, &p.Description, &p.Price.Id)
			if err != nil {
				return open, err
			}
			err = d.db.QueryRow(getPrice, p.Price.Id).Scan(&curPrice.LowEnd, &curPrice.HighEnd, &curPrice.Currency)
			if err != nil {
				return open, err
			}
			p.Price.LowEnd = curPrice.LowEnd
			p.Price.HighEnd = curPrice.HighEnd
			p.Price.Currency = curPrice.Currency
			// add one by one
			open[i].AppendProduct(p)
		}
		err = rows.Close()
		if err != nil {
			return open, err
		}
		// add owners info
		owners, err := d.getBrandOwners(brand.GetId())
		if err != nil {
			return open, err
		}
		open[i].AppendOwner(owners...)
		// add contacts info
		contacts, err := d.getBrandContacts(brand.GetId())
		if err != nil {
			return open, err
		}
		open[i].AppendContact(contacts...)
		// search for stat info
		rows, err = d.db.Query(getStats, brand.GetId())
		if err != nil {
			return open, err
		}
		for rows.Next() {
			stat := models.StatisticMeasure{}
			err := rows.Scan(&stat.Id, &stat.Name, &stat.Description, &stat.StartPeriod, &stat.EndPeriod, &stat.Value)
			if err != nil {
				return open, err
			}
			// add one by one
			open[i].AppendStat(stat)
		}
		err = rows.Close()
		if err != nil {
			return open, err
		}
	}
	return open, nil
}

func (d *Database) CreateUser(u models.User) error {
	insertUser := `INSERT INTO users(email, password, name, surname) VALUES ($1, $2, $3, $4)`
	passwd, err := bcrypt.GenerateFromPassword([]byte(u.Password), BcryptCost)
	if err != nil {
		return err
	}
	_, err = d.db.Query(insertUser, u.Email, "'"+string(passwd)+"'", u.Name, u.Surname)
	return err
}

func (d *Database) InsertOwner(o models.Owner) error {
	insertOwner := `INSERT INTO owners (name, surname, fathername, bio_info) VALUES ($1, $2, $3, $4)`
	_, err := d.db.Query(insertOwner, o.Per.Name, o.Per.Surname, o.Per.Fathername, o.Per.BioInfo)
	return err
}

func (d *Database) GetPasswordByEmail(mail string) (string, error) {
	getPassword := `SELECT password FROM users WHERE email = $1`
	rows, err := d.db.Query(getPassword, mail)
	if err != nil {
		return "", err
	}
	var passwd string
	for rows.Next() {
		err = rows.Scan(&passwd)
		if err != nil {
			return "", err
		}
	}
	err = rows.Close()
	if err != nil {
		return "", err
	}
	if passwd == "" {
		return "", nil
	}
	parts := strings.Split(passwd, "'") // password is stored as 'hash', because hash can contain random symbols like / or $
	return parts[1], nil
}

func (d *Database) collectIds(rows *sql.Rows) ([]int, error) {
	defer func() { _ = rows.Close() }()
	var res []int
	for rows.Next() {
		id := 0
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		res = append(res, id)
	}
	return res, nil
}

func (d *Database) generateLinkTableStatement(coreStmt, bId string, ids []int) string {
	for _, id := range ids { // we will get insert ... values (strId, c_id_0), (strId, c_id_1), ...
		coreStmt += " ("
		coreStmt += bId
		coreStmt += ", "
		coreStmt += strconv.Itoa(id)
		coreStmt += "),"
	}
	coreStmt = coreStmt[:len(coreStmt)-1]
	return coreStmt
}

// AddBrand this is awfully complex; the best way to refactor is by creating some complex object on pg side, maybe sometime in the future I will do it...
func (d *Database) AddBrand(c context.Context, b *models.Brand) error {
	ctx, cancel := context.WithCancel(c)
	defer cancel()
	addCore := `INSERT INTO brands (name, description, city, is_open) VALUES ($1, $2, $3, $4) RETURNING id`
	addContacts := `INSERT INTO contacts (type, contact) VALUES`
	addContacts = b.GetBulkInsertStatementContacts(addContacts)
	addContacts += ` RETURNING id`
	addOwners := `INSERT INTO owners (name, surname, fathername, bio_info) VALUES`
	addOwners = b.GetBulkInsertStatementOwners(addOwners)
	addOwners += `RETURNING id`

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// add core brand info
	rows, err := tx.Query(addCore, b.Name, b.Description, b.Location, b.IsOpen) // pq disables LastInsertId feature, unfortunately
	if err != nil {
		return err
	}
	id := 0
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			_ = rows.Close()
			return err
		}
	}
	err = rows.Close()
	if err != nil {
		return err
	}
	b.Id = id
	strId := strconv.Itoa(b.Id)
	d.log.Debug("brand added")
	// working with contacts here, add them, add mtm links
	if len(b.Contacts) > 0 {
		rows, err = tx.Query(addContacts)
		if err != nil {
			return err
		}
		d.log.Debug("contacts without links")
		// collect inserted ids
		ids, err := d.collectIds(rows)
		if err != nil {
			return err
		}
		// generate link table statement
		addLinks := `INSERT INTO l_brand_contacts (brand_id, contact_id) VALUES`
		addLinks = d.generateLinkTableStatement(addLinks, strId, ids)
		d.log.Debug(addLinks)
		// adding rows
		_, err = tx.Exec(addLinks)
		if err != nil {
			return err
		}
		d.log.Debug("contacts with links added")
	}
	d.log.Debug("contacts added")
	// working with owners here, add them, add mtm links
	if len(b.Owners) > 0 {
		d.log.Debug(addOwners)
		rows, err = tx.Query(addOwners)
		if err != nil {
			return err
		}
		// collect inserted ids
		ids, err := d.collectIds(rows)
		if err != nil {
			return err
		}
		// generate link table statement
		addLinks := `INSERT INTO l_brand_owners (brand_id, owner_id) VALUES`
		addLinks = d.generateLinkTableStatement(addLinks, strId, ids)
		// adding rows
		_, err = tx.Exec(addLinks)
		if err != nil {
			return err
		}
	}
	d.log.Debug("owners added")
	// just add statistics
	addStatistics := `INSERT INTO statistics (start_time, end_time, name, description, value, brand_id) VALUES`
	addStatistics = b.GetBulkInsertStatementStatistics(addStatistics)
	if len(b.Statistics) > 0 {
		d.log.Debug(addStatistics)
		_, err = tx.Exec(addStatistics)
		if err != nil {
			return err
		}
	}
	d.log.Debug("statistics added")
	// now add products...
	addPrices := `INSERT INTO prices (low_end, high_end, currency) VALUES`
	addPrices = b.GetBulkInsertStatementPrices(addPrices)
	addPrices += ` RETURNING Id`
	if len(b.Products) > 0 {
		// add prices
		d.log.Debug(addPrices)
		rows, err = tx.Query(addPrices)
		if err != nil {
			return err
		}
		// get their ids
		ids, err := d.collectIds(rows)
		if err != nil {
			return err
		}
		addProducts := `INSERT INTO products (name, description, price_id, brand_id) VALUES`
		addProducts = b.GetBulkInsertStatementProducts(addProducts, ids)
		// finally
		_, err = tx.Exec(addProducts)
		if err != nil {
			return err
		}
	}
	d.log.Debug("products added")
	err = tx.Commit()
	return err
}
