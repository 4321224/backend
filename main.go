package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Customer struct {
	ID           int     `json:"id" bson:"id"`
	NIK          string  `json:"nik" bson:"nik"`
	FullName     string  `json:"full_name" bson:"full_name"`
	Username     string  `json:"legalname" bson:"legalname"`
	BirthPlace   string  `json:"birth_place" bson:"birth_place"`
	BirthDate    time.Time  `json:"birth_date" bson:"birth_date"`
	Salary       float64 `json:"salary" bson:"salary"`
	PasswordHash string  `json:"password_hash" bson:"password_hash"`
	KTPImage     []byte  `json:"ktp_image" bson:"ktp_image"`
	SelfieImage  []byte  `json:"selfie_image" bson:"selfie_image"`
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

var db *sql.DB

func init() {
	var err error

	db, err = sql.Open("mysql", "root:@/xyz_multifinance")
	if err != nil {
		panic(err.Error())
	}
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Connected to the database")
}

func main() {
	r := gin.Default()

	r.POST("/register", registerHandler)
	r.POST("/login", loginHandler)

	auth := r.Group("/")

	auth.Use(authMiddleware())
	{
		auth.GET("/customers", customersHandler)
		auth.POST("/transactions", transactionsCreateHandler)
		auth.GET("/transaction/list", transactionsListHandler)
		auth.GET("/transactions/:id", transactionByIDHandler)
	}

	r.Run(":8080")
}

func registerHandler(c *gin.Context) {
	var customer Customer
	if err := c.BindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(customer.PasswordHash), bcrypt.DefaultCost)

	query := "INSERT INTO Customers(NIK, FullName, LegalName, BirthPlace, BirthDate, Salary) VALUES(?, ?, ?, ?, ?, ?)"

	result, err := db.Exec(query,
		customer.NIK,
		customer.FullName,
		customer.Username,
		customer.BirthPlace,
		customer.BirthDate.Format("2006-01-02"),
		hashedPassword)

	if err != nil {
		c.JSON(http.StatusInternalServerError, "Failed to insert new user.")
		return
	}

	lastInsertedId, errLastInsertedId := result.LastInsertId()

	if errLastInsertedId != nil {
		c.JSON(http.StatusInternalServerError, "Failed to get last inserted id.")
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Registration successful",
		"user_id": lastInsertedId})

}

func loginHandler(c *gin.Context) {
	var customer Customer

	errBindJSON := c.ShouldBindJSON(&customer)

	if errBindJSON != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	usernameFromRequest := customer.NIK
	passwordFromRequest := customer.PasswordHash

	var customerInDB Customer

	row := db.QueryRow("SELECT * FROM Customers WHERE NIK = ?", usernameFromRequest)
	errScan := row.Scan(&customerInDB.ID, &customerInDB.NIK, &customerInDB.FullName,
		&customerInDB.Username, &customerInDB.BirthPlace,
		&customerInDB.BirthDate, &customerInDB.Salary,
		&customerInDB.KTPImage, &customerInDB.SelfieImage)

	if errScan != nil {
		c.JSON(http.StatusUnauthorized, "Invalid credentials")
		return
	}

	errComparePwd := bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash), []byte(passwordFromRequest))

	if errComparePwd != nil && errComparePwd == bcrypt.ErrMismatchedHashAndPassword {

		c.JSON(http.StatusUnauthorized, "Invalid credentials")
		return

	}

	tokenString, errGenerateToken := generateToken(customer)

	if errGenerateToken != nil {
		log.Println(errGenerateToken)
		c.JSON(http.StatusInternalServerError, "Failed to generate token.")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"token":  tokenString})
}

var jwtKey = []byte("your_secret_key")

func generateToken(user Customer) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.NIK,
	})

	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// var jwtKey = []byte("your_secret_key")

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No Authorization header provided"})
			c.Abort()
			return
		}

		bearerToken := strings.Split(authHeader, " ")

		if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, "Invalid or expired JWT token")
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		username := claims["username"].(string)

		var customerInDB Customer

		row := db.QueryRow("SELECT * FROM Customers WHERE NIK = ?", username)
		errScan := row.Scan(&customerInDB.ID, &customerInDB.NIK, &customerInDB.FullName,
			&customerInDB.Username, &customerInDB.BirthPlace,
			&customerInDB.BirthDate, &customerInDB.Salary,
			&customerInDB.KTPImage, &customerInDB.SelfieImage)

		if errScan != nil {
			c.JSON(http.StatusUnauthorized, "Invalid credentials")
			c.Abort()
			return
		}
		c.Next()
	}
}

func customersHandler(c *gin.Context) {
	// Create an empty slice to hold the customers.
	var customers []Customer

	// Query the database for all customer records.
	rows, err := db.Query("SELECT * FROM Customers")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while fetching the customer data."})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var customer Customer
		err := rows.Scan(&customer.ID, &customer.NIK, &customer.FullName,
			&customer.Username, &customer.BirthPlace,
			&customer.BirthDate, &customer.Salary,
			&customer.KTPImage, &customer.SelfieImage)

		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while fetching the customer data."})
			return
		}

		customers = append(customers, customer)
	}

	c.JSON(http.StatusOK, gin.H{"customers": customers})
}

func transactionsCreateHandler(c *gin.Context) {
	var transaction Transaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	query := "INSERT INTO Transactions (CustomerID, ContractNo, OnTheRoad, AdminFee, MonthlyInstallment, Interest, ItemName) VALUES (?, ?, ?, ?, ?, ?, ?)"
	result, err := db.Exec(query,
		transaction.CustomerID,
		transaction.ContractNo,
		transaction.OnTheRoad,
		transaction.AdminFee,
		transaction.MonthlyInstallment,
		transaction.Interest,
		transaction.ItemName)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while creating the transaction."})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while getting last inserted ID."})
		return
	}

	transaction.ID = int(id)

	c.JSON(http.StatusCreated, gin.H{"transaction": transaction})
}

func transactionsListHandler(c *gin.Context) {
	var transactions []Transaction

	rows, err := db.Query("SELECT * FROM Transactions")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while fetching the transaction data."})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var transaction Transaction
		err := rows.Scan(&transaction.ID,
			&transaction.CustomerID,
			&transaction.ContractNo,
			&transaction.OnTheRoad,
			&transaction.AdminFee,
			&transaction.MonthlyInstallment,
			&transaction.Interest,
			&transaction.ItemName)

		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while fetching the transaction data."})
			return
		}

		transactions = append(transactions, transaction)
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

func transactionByIDHandler(c *gin.Context) {
	id := c.Param("id")

	var transaction Transaction

	err := db.QueryRow("SELECT * FROM Transactions WHERE ID = ?", id).Scan(
		&transaction.ID,
		&transaction.CustomerID,
		&transaction.ContractNo,
		&transaction.OnTheRoad,
		&transaction.AdminFee,
		&transaction.MonthlyInstallment,
		&transaction.Interest,
		&transaction.ItemName)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "No record found with provided ID"})
			return
		}

		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred while fetching the record."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transaction})
}
