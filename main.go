package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type Customer struct {
	ID          int
	NIK         string
	FullName    string
	LegalName   string
	BirthPlace  string
	BirthDate   string
	Salary      float64
	KTPImage    []byte 
	SelfieImage []byte 
}

type Transaction struct {
	ID                 int
	CustomerID         int 
	ContractNo         string
	OnTheRoad          float64
	AdminFee           float64
	MonthlyInstallment float64
	Interest           float64
	ItemName           string
}


func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "" 
	dbName := "xyz_multifinance"

	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

func getAllCustomers() []Customer {
	db := dbConn()
	selDB, err := db.Query("SELECT * FROM Customers")
	if err != nil {
		panic(err.Error())
	}
	custs := make([]Customer, 0)
	for selDB.Next() {
		var cust Customer
		err = selDB.Scan(&cust.ID, &cust.NIK, &cust.FullName,
			&cust.LegalName, &cust.BirthPlace, &cust.BirthDate,
			&cust.Salary, &cust.KTPImage, &cust.SelfieImage)
		if err != nil {
			panic(err.Error())
		}
		custs = append(custs, cust)
	}
	defer db.Close()
	return custs

}

func main() {

	fmt.Println("XYZ Multifinance Application")

	customers := getAllCustomers()

	for _, customer := range customers {
		fmt.Println(customer.FullName)

	}

}
