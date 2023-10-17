package bank_card

import (
	"encoding/csv"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestImportBankName(t *testing.T) {
	db := NewDatabase()
	db.OpenFromFile()
	defer db.Close()

	file, err := os.Open("name.csv")
	if err != nil {
		fmt.Println("Read file err, err =", err)
		return
	}
	defer file.Close()

	fileinfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}

	filesize := fileinfo.Size()
	buffer := make([]byte, filesize)

	_, err = file.Read(buffer)
	if err != nil {
		fmt.Println(err)
		return
	}

	r := csv.NewReader(strings.NewReader(string(buffer)))

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if len(record) == 2 && len(record[0]) > 0 && len(record[1]) > 0 {
			db.insertDataToBankTable(&Bank{BankCode: record[0], BankName: record[1]})
		}

		fmt.Println(record)
	}
}

func TestImportBankCard(t *testing.T) {
	db := NewDatabase()
	db.OpenFromFile()
	defer db.Close()

	file, err := os.Open("bin.csv")
	if err != nil {
		fmt.Println("Read file err, err =", err)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}

	filesize := fileInfo.Size()
	buffer := make([]byte, filesize)

	_, err = file.Read(buffer)
	if err != nil {
		fmt.Println(err)
		return
	}

	r := csv.NewReader(strings.NewReader(string(buffer)))

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if len(record) == 4 && len(record[0]) > 0 && len(record[1]) > 0 && len(record[2]) > 0 && len(record[3]) > 0 {
			bin, _ := strconv.Atoi(record[0])
			cardLength, _ := strconv.Atoi(record[3])
			db.insertDataToBankCardTable(&BankCard{
				BIN:        uint32(bin),
				BankCode:   record[1],
				CardType:   record[2],
				CardLength: uint32(cardLength),
			})
		}

		fmt.Println(record)
	}
}

func TestImportBankCardSingle(t *testing.T) {
	db := NewDatabase()
	db.OpenFromFile()
	defer db.Close()

	binStr := "620114|620187|620046"
	strs := strings.Split(binStr, "|")
	var bins []uint32
	for _, str := range strs {
		bin, _ := strconv.Atoi(str)
		bins = append(bins, uint32(bin))
	}

	bankCode := "ABC"
	cardType := "DC"
	cardLength := 13 + len(strs[0])
	for _, data := range bins {
		bankData := &BankCard{
			BIN:        data,
			BankCode:   bankCode,
			CardType:   cardType,
			CardLength: uint32(cardLength),
		}
		db.insertDataToBankCardTable(bankData)
		fmt.Println(bankData)
	}
}

func TestOpenFromEmbed(t *testing.T) {
	db := NewDatabase()
	defer db.Close()

	err := db.OpenFromEmbed()
	assert.Nil(t, err)

	var bank *Bank
	bank = db.queryBank("ABC")
	assert.NotNil(t, bank)
	assert.Equal(t, bank.BankCode, "ABC")
	assert.Equal(t, bank.BankName, "中国农业银行")

	var bankCard *BankCard
	bankCard = db.queryBankCard(620114)
	assert.NotNil(t, bankCard)
	assert.Equal(t, bankCard.BankCode, "ICBC")
	assert.Equal(t, bankCard.BankName, "中国工商银行")
	assert.Equal(t, bankCard.CardType, "PC")
	assert.Equal(t, bankCard.CardName, "")
	assert.Equal(t, bankCard.CardLength, uint32(19))
}
