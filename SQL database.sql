
CREATE USER 'user'@'localhost' IDENTIFIED BY
'password';
GRANT ALL ON *.* TO 'user'@'localhost'

-- Step 1: Create the database
CREATE DATABASE IF NOT EXISTS car_sharing;

-- Step 2: Use the database
USE car_sharing;
select * from users;

-- Step 3: Create the courses table
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'Basic',
    is_verified BOOLEAN DEFAULT FALSE,
    verification_token VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS membership_tiers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    hourly_rate_discount DECIMAL(5, 2) DEFAULT 0.0,
    priority_access BOOLEAN DEFAULT FALSE,
    booking_limit INT DEFAULT 0
);

-- Insert default membership tiers
INSERT INTO membership_tiers (name, hourly_rate_discount, priority_access, booking_limit)
VALUES
('Basic', 0.0, FALSE, 5),
('Premium', 10.0, TRUE, 10),
('VIP', 20.0, TRUE, 20);

ALTER TABLE users ADD COLUMN membership_tier_id INT DEFAULT 1, 
ADD FOREIGN KEY (membership_tier_id) REFERENCES membership_tiers(id);

CREATE TABLE vehicles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    make VARCHAR(50),
    model VARCHAR(50),
    registration_number VARCHAR(20) UNIQUE,
    is_available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE bookings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    vehicle_id INT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    status ENUM('confirmed', 'modified', 'canceled') DEFAULT 'confirmed',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id)
);
ALTER TABLE bookings MODIFY status ENUM('confirmed', 'modified', 'canceled', 'completed') DEFAULT 'confirmed';

CREATE TABLE vehicle_status (
    vehicle_id INT PRIMARY KEY,
    location VARCHAR(255),
    charge_level INT CHECK (charge_level BETWEEN 0 AND 100),
    cleanliness ENUM('clean', 'dirty', 'needs maintenance') DEFAULT 'clean',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id)
);
-- Insert sample data into vehicles table
INSERT INTO vehicles (make, model, registration_number, is_available)
VALUES 
('Tesla', 'Model S', 'ABC123', TRUE),
('BMW', 'X5', 'XYZ789', TRUE),
('Audi', 'A4', 'AUD456', TRUE),
('Toyota', 'Corolla', 'TOY789', TRUE),
('Honda', 'Civic', 'HON123', TRUE);

-- Insert sample data into vehicle_status table
INSERT INTO vehicle_status (vehicle_id, location, charge_level, cleanliness)
VALUES
(1, 'Garage A', 85, 'clean'),
(2, 'Garage B', 60, 'dirty'),
(3, 'Garage C', 95, 'clean'),
(4, 'Garage D', 50, 'needs maintenance'),
(5, 'Garage E', 75, 'clean');

-- We will create a trigger to update vehicle availability based on cleanliness
DELIMITER //

CREATE TRIGGER update_vehicle_availability_before_update
BEFORE UPDATE ON vehicle_status
FOR EACH ROW
BEGIN
    IF NEW.cleanliness = 'needs maintenance' THEN
        UPDATE vehicles SET is_available = FALSE WHERE id = NEW.vehicle_id;
    ELSE
        UPDATE vehicles SET is_available = TRUE WHERE id = NEW.vehicle_id;
    END IF;
END//

DELIMITER ;

-- Payment table to store payment transactions
CREATE TABLE IF NOT EXISTS payments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    payment_status ENUM('pending', 'completed', 'failed') DEFAULT 'pending',
    payment_method VARCHAR(255),
    payment_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
ALTER TABLE payments
ADD COLUMN booking_id INT NOT NULL,
ADD FOREIGN KEY (booking_id) REFERENCES bookings(id);
-- Invoice table to store details of generated invoices
CREATE TABLE IF NOT EXISTS invoices (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    booking_id INT NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    invoice_status ENUM('generated', 'paid') DEFAULT 'generated',
    invoice_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (booking_id) REFERENCES bookings(id)
);

SHOW TABLES;


-- Create an index for `is_available` to optimize availability queries
CREATE INDEX idx_vehicle_is_available ON vehicles(is_available);

-- Enforce no overlapping reservations for the same vehicle
CREATE UNIQUE INDEX idx_no_overlap ON bookings(vehicle_id, start_time, end_time)
WHERE status = 'confirmed';



-- Payment methods table: Store different payment methods for users (e.g., credit card, PayPal)
CREATE TABLE IF NOT EXISTS payment_methods (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    method_type VARCHAR(50) NOT NULL,   -- Type of payment method (e.g., 'Credit Card', 'PayPal')
    method_details TEXT NOT NULL,       -- Encrypted details (e.g., last 4 digits of card, token for PayPal)
    expiration_date DATE,               -- Expiration date for cards, etc.
    is_active BOOLEAN DEFAULT TRUE,     -- Whether the payment method is active
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Refunds table: To track refunds for cancellations or payment issues
CREATE TABLE IF NOT EXISTS refunds (
    id INT AUTO_INCREMENT PRIMARY KEY,
    payment_id INT NOT NULL,            -- The associated payment for which the refund was issued
    amount DECIMAL(10, 2) NOT NULL,     -- Refund amount
    refund_status ENUM('pending', 'completed', 'failed') DEFAULT 'pending', -- Status of refund
    refund_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reason TEXT,                        -- Reason for the refund (optional)
    FOREIGN KEY (payment_id) REFERENCES payments(id)
);

-- Modify payments table to include payment method reference
ALTER TABLE payments
ADD COLUMN payment_method_id INT,   -- Foreign key to payment_methods table
ADD FOREIGN KEY (payment_method_id) REFERENCES payment_methods(id);

-- Modify invoices table to include payment_status and track if it’s been paid
ALTER TABLE invoices
ADD COLUMN payment_status ENUM('unpaid', 'paid', 'refunded') DEFAULT 'unpaid';  -- Track payment status of invoices
