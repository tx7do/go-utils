package bank_card

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed bank_card.db
var embeddedDatabase []byte

type Database struct {
	db *sql.DB
}

func NewDatabase(openFile bool) *Database {
	db := &Database{}

	if openFile {
		db.openFromFile()
	} else {
		_ = db.openFromEmbed()
	}

	return db
}

// openFromFile 从文件打开数据库
func (d *Database) openFromFile() {
	db, err := sql.Open("sqlite3", "bank_card.db")
	if err != nil {
		panic("failed to connect database")
	}

	d.db = db

	d.initTables()
}

// openFromEmbed 从内嵌文件打开数据库
func (d *Database) openFromEmbed() error {
	db, err := sql.Open("sqlite3", "file::memory:?mode=ro&cache=shared")
	if err != nil {
		panic("failed to connect database")
		return err
	}

	conn, err := db.Conn(context.Background())
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer conn.Close()

	if err = conn.Raw(func(raw interface{}) error {
		return raw.(*sqlite3.SQLiteConn).Deserialize(embeddedDatabase, "")
	}); err != nil {
		log.Fatal(err)
		return err
	}
	conn.Close()

	d.db = db

	return nil
}

func (d *Database) Close() {
	if d.db != nil {
		_ = d.db.Close()
	}
}

// initTables 创建表
func (d *Database) initTables() {
	strSql := `
CREATE TABLE IF NOT EXISTS bank_cards
(
    bin         INTEGER PRIMARY KEY,
    bank_code   TEXT,
    card_name   TEXT,
    card_type   TEXT,
    card_length INTEGER
);

CREATE TABLE IF NOT EXISTS banks
(
    id        INTEGER PRIMARY KEY autoincrement,
    bank_code TEXT,
    bank_name TEXT
);
`
	_, err := d.db.Exec(strSql)
	if err != nil {
		log.Println("initial database table failed:", err)
	}
}

func (d *Database) insertDataToBankTable(data *Bank) {
	strSql := fmt.Sprintf("INSERT INTO banks (bank_code, bank_name) VALUES ('%s', '%s');", data.BankCode, data.BankName)
	_, err := d.db.Exec(strSql)
	if err != nil {
		log.Println(err)
	}
}

func (d *Database) insertDataToBankCardTable(data *BankCard) {
	strSql := fmt.Sprintf("INSERT INTO bank_cards (bin, bank_code, card_name, card_type, card_length) VALUES (%d, '%s', '%s', '%s', %d);",
		data.BIN, data.BankCode, data.CardName, data.CardType, data.CardLength)
	_, err := d.db.Exec(strSql)
	if err != nil {
		log.Println(err)
	}
}

func (d *Database) UpdateBankCardTableCardName(bin uint32, cardName string) {
	strSql := fmt.Sprintf("UPDATE bank_cards SET card_name = '%s' WHERE bin = '%d';", cardName, bin)
	_, err := d.db.Exec(strSql)
	if err != nil {
		log.Println(err)
	}
}

// queryBank 查询银行信息
func (d *Database) queryBank(bankCode string) *Bank {
	strSql := fmt.Sprintf("SELECT id, bank_code, bank_name FROM banks WHERE bank_code = '%s' LIMIT 1;", bankCode)
	rows, err := d.db.Query(strSql)
	if err != nil {
		log.Fatal("query bank failed:", err)
		return nil
	}
	defer rows.Close()

	rows.Next()
	var bank Bank
	if err = rows.Scan(&bank.Id, &bank.BankCode, &bank.BankName); err != nil {
		log.Fatal("scan bank failed:", err)
		return nil
	}

	//fmt.Printf("bank = %v\n", bank)

	return &bank
}

// queryBankCard 查询银行卡信息
func (d *Database) queryBankCard(bin uint32) *BankCard {
	strSql := fmt.Sprintf("SELECT bin, bank_code, card_name, card_type, card_length FROM bank_cards WHERE bin = '%d' LIMIT 1;", bin)
	rows, err := d.db.Query(strSql)
	if err != nil {
		log.Fatal("query bank card failed:", err)
		return nil
	}
	defer rows.Close()

	rows.Next()
	var bankCard BankCard
	if err = rows.Scan(&bankCard.BIN, &bankCard.BankCode, &bankCard.CardName, &bankCard.CardType, &bankCard.CardLength); err != nil {
		log.Fatal("scan bank card failed:", err)
		return nil
	}
	rows.Close()

	bank := d.queryBank(bankCard.BankCode)
	if bank != nil {
		bankCard.BankName = bank.BankName
	}

	//fmt.Printf("bankCard = %v\n", bankCard)

	return &bankCard
}
