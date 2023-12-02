package db

import (
	"database/sql"

	"accelerator/models"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

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
	createStats := `CREATE TABLE IF NOT EXISTS Statistics(id SERIAL PRIMARY KEY, start_time TIMESTAMP, end_time TIMESTAMP, name VARCHAR(50), description VARCHAR(200), value NUMERIC, brand_id INT, CONSTRAINT fk_statistics FOREIGN KEY(brand_id) REFERENCES Brands(id) ON DELETE CASCADE)`
	createPrices := `CREATE TABLE IF NOT EXISTS Prices(id SERIAL PRIMARY KEY, low_end INT, high_end INT, currency VARCHAR(20))`
	createProducts := `CREATE TABLE IF NOT EXISTS Products(id SERIAL PRIMARY KEY, name VARCHAR(50), description VARCHAR(50), price_id INT, brand_id INT, CONSTRAINT fk_price_product FOREIGN KEY(price_id) REFERENCES Prices(id) ON DELETE CASCADE, CONSTRAINT fk_brand_product FOREIGN KEY(brand_id) REFERENCES Brands(id) ON DELETE CASCADE)`
	createContacts := `CREATE TABLE IF NOT EXISTS Contacts(id SERIAL PRIMARY KEY, type VARCHAR(20), contact VARCHAR(100))`
	createLinkCBrand := `CREATE TABLE IF NOT EXISTS L_Brand_Contacts(id SERIAL PRIMARY KEY, brand_id INT, contact_id INT, CONSTRAINT fk_link_contacts FOREIGN KEY(brand_id) REFERENCES Brands(id) ON DELETE CASCADE, CONSTRAINT fk_contact_id FOREIGN KEY(contact_id) REFERENCES Contacts(id) ON DELETE CASCADE)`
	createHistory := `CREATE TABLE IF NOT EXISTS History(id SERIAL PRIMARY KEY)` // TBD LATER
	createOwners := `CREATE TABLE IF NOT EXISTS Owners(id SERIAL PRIMARY KEY, name VARCHAR(20), surname VARCHAR(50), fathername VARCHAR(50), bio_info VARCHAR(200), email VARCHAR(50), password VARCHAR(100), history_id INT, CONSTRAINT fk_history_id FOREIGN KEY (history_id) REFERENCES History(id))`
	createLOBrand := `CREATE TABLE IF NOT EXISTS L_Brand_Owners(id SERIAL PRIMARY KEY, brand_id INT, owner_id INT, CONSTRAINT fk_link_owners FOREIGN KEY(brand_id) REFERENCES Brands(id) ON DELETE CASCADE, CONSTRAINT fk_owner_id FOREIGN KEY(owner_id) REFERENCES Owners(id) ON DELETE CASCADE)`
	execList := []string{
		createBrand, createStats, createPrices, createProducts, createContacts, createLinkCBrand, createHistory,
		createOwners,
		createLOBrand,
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
	getOwnerIds := `SELECT owner_id FROM l_brand_owners WHERE brand_id = ?`
	getOwnerById := `SELECT (name, surname, fathername, bio_info) FROM owners WHERE id = ?` // email is not meant to be publicly visible
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
			err = owner.Scan(&cur)
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
	getContactIds := `SELECT contact_id FROM l_brand_contacts WHERE brand_id = ?`
	getContactById := `SELECT (type, contact) FROM contacts WHERE id = ?`
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
	getOpen := `SELECT (id, name, description, city, is_open) FROM brands WHERE is_open=TRUE`
	rows, err := d.db.Query(getOpen)
	if err != nil {
		return nil, err
	}
	var open []models.Brand
	for rows.Next() {
		cur := models.Brand{}
		err = rows.Scan(&cur)
		if err != nil {
			return nil, err
		}
		open = append(open, cur)
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}
	getProducts := `SELECT (id, name, description, price_id) FROM products WHERE brand_id = ?`
	getStats := `SELECT (id, name, description, start_time, end_time, value) FROM statistics WHERE brand_id = ?`
	for i, brand := range open {
		// search for products
		rows, err = d.db.Query(getProducts, brand.GetId())
		if err != nil {
			return open, err
		}
		for rows.Next() {
			p := models.Product{}
			err = rows.Scan(&p)
			if err != nil {
				return open, err
			}
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
			err := rows.Scan(&stat)
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

func (d *Database) InsertOwner(o models.Owner) error {
	insertOwner := `INSERT INTO owners (name, surname, fathername, bio_info, email, password) VALUES ($1, $2, $3, $4, $5, $6)`
	passwd, err := bcrypt.GenerateFromPassword([]byte(o.Per.Password), 17)
	if err != nil {
		return err
	}
	_, err = d.db.Query(insertOwner, o.Per.Name, o.Per.Surname, o.Per.Fathername, o.Per.BioInfo, o.Per.Email, "'"+string(passwd)+"'")
	return err
}
