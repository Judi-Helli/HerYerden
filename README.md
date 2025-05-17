
# HerYerden Backend

This is the backend API for the HerYerden project.

## Features

- User registration and login (with JWT authentication)
- Place orders
- List available orders
- Accept orders (driver functionality)
- MySQL database integration


- Set up the database:
- Create a MySQL database named `HerYerden`.
  
run the following SQL to create the tables:
 CREATE TABLE users (
    user_id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(100) UNIQUE,
    password VARCHAR(255),
    role ENUM('user', 'admin', 'driver'),
    phone_number VARCHAR(15),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE orders (
    order_id INT AUTO_INCREMENT PRIMARY KEY,
    customer_id INT,
    product_photo VARCHAR(255),
    description TEXT,
    location VARCHAR(255),
    status ENUM('pending', 'accepted', 'completed') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES users(user_id)
);

CREATE TABLE driver_orders (
    order_id INT,
    driver_id INT,
    status ENUM('accepted', 'completed'),
    PRIMARY KEY (order_id, driver_id),
    FOREIGN KEY (order_id) REFERENCES orders(order_id),
    FOREIGN KEY (driver_id) REFERENCES users(user_id)
); 

- Update your DB credentials in `main.go` 

- Run the server
