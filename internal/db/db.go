package db

import (
	"context"
	"database/sql"
	"errors"
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
	createBrand := `CREATE TABLE IF NOT EXISTS Brands(id SERIAL PRIMARY KEY, name VARCHAR(50) UNIQUE, DESCRIPTION VARCHAR(200), city VARCHAR(50), is_open BOOLEAN)`
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

func (d *Database) GetBrandIdByName(name string) (int, error) { // works because names are unique (see table definitions)
	var res int
	err := d.db.QueryRow(`SELECT id FROM brands WHERE name = $1`, name).Scan(&res)
	if errors.Is(err, sql.ErrNoRows) {
		return -1, nil
	}
	return res, err
}

func (d *Database) GetBrandById(id int) (models.Brand, error) {
	var res models.Brand
	getCore := `SELECT name, description, city FROM brands WHERE id = $1`
	getProducts := `SELECT id, name, description, price_id FROM products WHERE brand_id = $1`
	getStats := `SELECT id, name, description, start_time, end_time, value FROM statistics WHERE brand_id = $1`
	getPrice := `SELECT low_end, high_end, currency FROM prices WHERE id = $1`
	// get core info first
	err := d.db.QueryRow(getCore, id).Scan(&res.Name, &res.Description, &res.Location)
	if err != nil {
		return res, err
	}
	// search for products
	rows, err := d.db.Query(getProducts, id)
	if err != nil {
		return res, err
	}
	curPrice := models.Price{}
	for rows.Next() {
		p := models.Product{}
		err = rows.Scan(&p.Id, &p.Name, &p.Description, &p.Price.Id)
		if err != nil {
			return res, err
		}
		err = d.db.QueryRow(getPrice, p.Price.Id).Scan(&curPrice.LowEnd, &curPrice.HighEnd, &curPrice.Currency)
		if err != nil {
			return res, err
		}
		p.Price.LowEnd = curPrice.LowEnd
		p.Price.HighEnd = curPrice.HighEnd
		p.Price.Currency = curPrice.Currency
		// add one by one
		res.AppendProduct(p)
	}
	err = rows.Close()
	if err != nil {
		return res, err
	}
	// add owners info
	owners, err := d.getBrandOwners(id)
	if err != nil {
		return res, err
	}
	res.AppendOwner(owners...)
	// add contacts info
	contacts, err := d.getBrandContacts(id)
	if err != nil {
		return res, err
	}
	res.AppendContact(contacts...)
	// search for stat info
	rows, err = d.db.Query(getStats, id)
	if err != nil {
		return res, err
	}
	for rows.Next() {
		stat := models.StatisticMeasure{}
		err := rows.Scan(&stat.Id, &stat.Name, &stat.Description, &stat.StartPeriod, &stat.EndPeriod, &stat.Value)
		if err != nil {
			return res, err
		}
		// add one by one
		res.AppendStat(stat)
	}
	err = rows.Close()
	if err != nil {
		return res, err
	}
	return res, nil
}

func (d *Database) GetOpenBrands() ([]models.Brand, error) {
	getOpen := `SELECT id FROM brands WHERE is_open=TRUE`
	rows, err := d.db.Query(getOpen)
	if err != nil {
		return nil, err
	}
	openIds, err := d.collectIds(rows)
	if err != nil {
		return nil, err
	}
	var open []models.Brand
	for _, brand := range openIds {
		// search for products
		cur, err := d.GetBrandById(brand)
		if err != nil {
			return open, err
		}
		open = append(open, cur)
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

func (d *Database) UpdateBrand(c context.Context, b *models.Brand) error {
	ctx, cancel := context.WithCancel(c)
	defer cancel()
	var contactIds []int
	getContactIds := `SELECT contact_id FROM l_brand_contacts WHERE brand_id = $1`
	getOwnerIds := `SELECT owner_id FROM l_brand_owners WHERE brand_id = $1`
	getPriceIds := `SELECT price_id FROM products WHERE brand_id = $1`
	getProductIds := `SELECT id FROM products WHERE brand_id = $1`
	getStatsIds := `SELECT id FROM statistics WHERE brand_id = $1`
	updateCoreInfo := `UPDATE brands SET name = $1, description = $2, city = $3, is_open = $4 WHERE id = $5`
	updateContacts := `UPDATE contacts as c SET type = vals.type, contact = vals.contact FROM (VALUES ?) AS vals(id, type, contact) WHERE vals.id = c.id`
	updateOwners := `UPDATE owners as o SET name = vals.name, surname = vals.surname, fathername = vals.fathername, bio_info = vals.bio_info FROM (VALUES ?) AS vals(id, name, surname, fathername, bio_info) WHERE vals.id = o.id`
	updatePrices := `UPDATE prices as p SET low_end = vals.low_end, high_end = vals.high_end, currency = vals.currency FROM (VALUES ?) AS vals(id, low_end, high_end, currency) WHERE vals.id = p.id`
	updateProducts := `UPDATE products as p SET name = vals.name, description = vals.description FROM (VALUES ?) AS vals(id, name, description) WHERE vals.id = p.id`
	updateStatistics := `UPDATE statistics as s SET start_time = vals.start_time, end_time = vals.end_time, name = vals.name, description = vals.description, value = vals.value  FROM (VALUES ?) AS vals(id, start_time, end_time, name, description, value) WHERE vals.id = s.id`
	// transaction begins here
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// update core info
	_, err = tx.Exec(updateCoreInfo, b.Name, b.Description, b.Location, b.IsOpen, b.Id)
	if err != nil {
		return err
	}
	// get contact ids for updating
	rows, err := tx.Query(getContactIds, b.Id)
	if err != nil {
		return err
	}
	contactIds, err = d.collectIds(rows)
	if err != nil {
		return err
	}
	// for simplicity let's assume that...
	if len(b.Contacts) != len(contactIds) {
		return errors.New("insufficient update data: not all contacts are listed")
	}
	updateContacts = b.GetBulkUpdateStatementContacts(updateContacts, contactIds)
	if len(b.Contacts) > 0 {
		_, err = tx.Exec(updateContacts)
		if err != nil {
			return err
		}
	}
	// moving on to owners
	rows, err = tx.Query(getOwnerIds, b.Id)
	if err != nil {
		return err
	}
	ownerIds, err := d.collectIds(rows)
	if err != nil {
		return err
	}
	// again, we assume that
	if len(b.Owners) != len(ownerIds) {
		return errors.New("insufficient update data: not all owners are listed")
	}
	updateOwners = b.GetBulkUpdateStatementOwners(updateOwners, ownerIds)
	if len(b.Owners) > 0 {
		_, err = tx.Exec(updateOwners)
		if err != nil {
			return err
		}
	}
	// now updating prices
	rows, err = tx.Query(getPriceIds, b.Id)
	if err != nil {
		return err
	}
	priceIds, err := d.collectIds(rows)
	if err != nil {
		return err
	}
	// again, we need to have all information listed
	if len(b.Products) != len(priceIds) {
		return errors.New("insufficient update data: not all products are listed")
	}
	updatePrices = b.GetBulkUpdateStatementPrices(updatePrices, priceIds)
	if len(b.Products) > 0 {
		_, err = tx.Exec(updatePrices)
		if err != nil {
			return err
		}
	}
	// update products now (done with prices)
	rows, err = tx.Query(getProductIds, b.Id)
	if err != nil {
		return err
	}
	productIds, err := d.collectIds(rows)
	if err != nil {
		return err
	}
	updateProducts = b.GetBulkUpdateStatementProducts(updateProducts, productIds)
	if len(b.Products) > 0 {
		_, err = tx.Exec(updateProducts)
		if err != nil {
			return err
		}
	}
	// now only thing left is to update statistics
	rows, err = tx.Query(getStatsIds, b.Id)
	if err != nil {
		return err
	}
	statIds, err := d.collectIds(rows)
	if err != nil {
		return err
	}
	if len(b.Statistics) != len(statIds) {
		return errors.New("insufficient update information: not all statistics measures are listed")
	}
	updateStatistics = b.GetBulkUpdateStatementStats(updateStatistics, statIds)
	d.log.Debug(updateStatistics)
	if len(b.Statistics) > 0 {
		_, err = tx.Exec(updateStatistics)
		if err != nil {
			return err
		}
	}
	// transaction committed, done
	err = tx.Commit()
	return err
}
