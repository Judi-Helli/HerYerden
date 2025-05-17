package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"github.com/golang-jwt/jwt/v4"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("secret_key")
// Define the Order struct
type Order struct {
	OrderID      int    `json:"order_id"`
	CustomerID   int    `json:"customer_id"`
	ProductPhoto string `json:"product_photo"`
	Description  string `json:"description"`
	Location     string `json:"location"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}
// Define the Claims struct to hold JWT claims
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}
// Connect to the database
func connectToDatabase() (*sql.DB, error) {
	dsn := "root:password@tcp(localhost:3306)/HerYerden" // Update with your database credentials
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// Generate JWT token for the user
func generateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	return tokenString, err
}


// Handler to register a user
func registerUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var user struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Phone    string `json:"phone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	db, err := connectToDatabase()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Hash the password before storing it
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	query := "INSERT INTO users (username, password, role, phone_number) VALUES (?, ?, ?, ?)"
	_, err = db.Exec(query, user.Username, hashedPassword, user.Role, user.Phone)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	// Generate JWT token after successful registration
	tokenString, err := generateJWT(user.Username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Respond with the token
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully",
		"token":   tokenString,
	})
}

// Handler to log in a user
func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Login handler reached") // Log this to see if the route is hit
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	db, err := connectToDatabase()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var storedHashedPassword string
	query := "SELECT password FROM users WHERE username = ?"
	err = db.QueryRow(query, credentials.Username).Scan(&storedHashedPassword)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Compare the hashed password with the provided one
	if err := bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(credentials.Password)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT token after successful login
	tokenString, err := generateJWT(credentials.Username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Respond with the token
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}


// Handler to place an order
func placeOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var order struct {
		CustomerID  int    `json:"customer_id"`
		Photo       string `json:"photo"`
		Description string `json:"description"`
		Location    string `json:"location"`
	}

	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	db, err := connectToDatabase()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := "INSERT INTO orders (customer_id, product_photo, description, location) VALUES (?, ?, ?, ?)"
	_, err = db.Exec(query, order.CustomerID, order.Photo, order.Description, order.Location)
	if err != nil {
		http.Error(w, "Failed to place order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Order placed successfully")
}

// Handler to get available orders
func getAvailableOrdersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	db, err := connectToDatabase()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := "SELECT order_id, customer_id, product_photo, description, location, status, created_at FROM orders WHERE status = 'pending'"
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Failed to retrieve orders", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.OrderID, &order.CustomerID, &order.ProductPhoto,
			 &order.Description, &order.Location, &order.Status, &order.CreatedAt); err != nil {
			http.Error(w, "Error scanning order", http.StatusInternalServerError)
			return
		}
		orders = append(orders, order)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// Handler to accept an order
func acceptOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		OrderID  int `json:"order_id"`
		DriverID int `json:"driver_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	db, err := connectToDatabase()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("UPDATE orders SET status = 'accepted' WHERE order_id = ?", request.OrderID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update order status", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("INSERT INTO driver_orders (order_id, driver_id, status) VALUES (?, ?, 'accepted')", request.OrderID, request.DriverID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to insert driver order", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Order accepted successfully")
}

func main() {
	http.HandleFunc("/register", registerUserHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/place-order", placeOrderHandler)
	http.HandleFunc("/orders", getAvailableOrdersHandler)
	http.HandleFunc("/accept-order", acceptOrderHandler)
	fmt.Println("Server is running on port 8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
