# Vehicle Rental Microservices Architecture

This project is designed using **Microservices Architecture** to manage a vehicle rental system. The system is split into distinct services to ensure scalability, maintainability, and independent deployment. The architecture supports user management, vehicle reservation, billing and payment processing, and real-time communication between services.

---

## Table of Contents
1. [Design Considerations of the Microservices](#design-considerations)
2. [Architecture Diagram](#architecture-diagram)
3. [Instructions for Setting Up and Running Microservices](#setup-instructions)
4. [Why Microservices Architecture?](#why-microservices)

---

## Design Considerations of the Microservices

### 1. Service Decomposition

The system is decomposed into distinct services that handle specific parts of the vehicle rental process. Each service operates independently and interacts with others via **RESTful APIs**.

- **User Service**:
  - Manages user registration, authentication, membership tiers, and profile management.
  - Users can register via email or phone, manage their membership levels (Basic, Premium, VIP), and track their rental history.

- **Vehicle Service**:
  - Handles vehicle availability, booking, modification, and cancellation.
  - Users can view available vehicles in real-time, make reservations, and modify/cancel them.

- **Billing Service**:
  - Handles real-time cost calculations based on membership tiers and rental duration.
  - Includes payment processing via secure methods, invoice generation, and email receipts.

**Benefits of Service Decomposition**:
- **Scalability**: Individual services can be scaled based on the load. For instance, the **Billing Service** can be scaled independently during peak times without affecting other services.
- **Maintainability**: Each service has a single responsibility, making it easier to update or maintain a specific part of the system without impacting others.

### 2. Inter-Service Communication

Services communicate using **RESTful APIs** with clear API contracts. This is implemented with HTTP methods like `GET`, `POST`, `PUT`, and `DELETE`, ensuring that the system is loosely coupled and easy to integrate with the frontend or other services.

- **User Service API**: Manages user-related endpoints like `/api/v1/users/register` for user registration and `/api/v1/users/login` for authentication.
- **Vehicle Service API**: Provides endpoints for vehicle operations like `/api/v1/vehicles` for viewing and reserving vehicles.
- **Billing Service API**: Handles payment processing and invoice generation with endpoints like `/api/v1/payment/confirm`.

By using **RESTful APIs**, the services are decoupled and can evolve independently, making the system more flexible and reducing the impact of changes.

### 3. Database Per Service

Each service manages its own database, ensuring complete data isolation. For example:
- **User Service** stores user data, including registration, profiles, and membership tiers.
- **Vehicle Service** stores vehicle details, availability, and booking information.
- **Billing Service** stores transaction details, payment status, and invoices.

This separation of data ensures that each service is responsible for its own domain, preventing data leakage and improving fault tolerance.

### 4. Error Handling and Logging

Each service includes robust error handling. Errors are logged with detailed messages to assist in debugging. Services return meaningful HTTP responses with appropriate error codes (e.g., 400 for invalid input, 500 for server errors). Logs are collected for monitoring and troubleshooting purposes.

---

## Architecture Diagram

Below is the architecture diagram for the Vehicle Rental System. It illustrates the microservices and their communication through RESTful APIs.

```plaintext
                             +------------------+
                             |   User Service   |
                             | (Authentication, |
                             | User Profiles,    |
                             | Membership)       |
                             +--------+---------+
                                      |
                            RESTful API |  RESTful API
                                      |
            +-------------------------+------------------------+
            |                                                  |
+-----------+------------+                              +-------+----------+
|  Vehicle Service       |                              |  Billing Service |
|  (Vehicle Registration,|                              |  (Payment, Invoice|
|  Rental Management)    |                              |  Generation)      |
+------------------------+                              +-------------------+
            |                                                  |
         RESTful API                                           RESTful API
            |                                                  |
    +-------+--------+                                     +---+-------------+
    |  Database      |                                     |  Database       |
    |  (User Data)   |                                     |  (Vehicle Data) |
    +----------------+                                     +-----------------+
This diagram highlights:

Service Decomposition: Each service is self-contained with its own database.
Loose Coupling: Services interact via RESTful APIs, ensuring minimal dependencies.
Scalability: Services can be scaled independently to meet varying demand.
4.1.3.3 Instructions for Setting Up and Running Microservices
Prerequisites:
Install Go version 1.16 or higher.
Install Docker if you plan to containerize the services.
Install a relational database like PostgreSQL or MySQL to run the services.
1. Clone the Repository
bash
Copy code
git clone https://github.com/your-repo/vehicle-rental-system.git
cd vehicle-rental-system
2. Configure Environment Variables
Create a .env file to define the necessary environment variables for the services:

bash
Copy code
USER_SERVICE_PORT=8081
VEHICLE_SERVICE_PORT=8082
BILLING_SERVICE_PORT=8083
DB_HOST=localhost
DB_PORT=5432
DB_USER=your-db-user
DB_PASSWORD=your-db-password
DB_NAME=vehicle_rental_db
3. Run the Database
Use Docker to set up a PostgreSQL or MySQL database, or configure it locally.
Run the following Docker command to set up a PostgreSQL container (for example):
bash
Copy code
docker run --name vehicle_rental_db -e POSTGRES_PASSWORD=yourpassword -p 5432:5432 -d postgres
4. Database Setup
Run the database migrations for each service to create the necessary tables. You can use a tool like Go Migrations for this purpose.

5. Run the Services
User Service:

bash
Copy code
cd user-service
go run main.go
Vehicle Service:

bash
Copy code
cd vehicle-service
go run main.go
Billing Service:

bash
Copy code
cd billing-service
go run main.go
Each service will now be running on its respective port:

User Service: http://localhost:8081
Vehicle Service: http://localhost:8082
Billing Service: http://localhost:8083

6. Test the Services
You can now test the services by interacting with the endpoints:

User Service: POST /api/v1/users/register for registration and POST /api/v1/users/login for authentication.
Vehicle Service: GET /api/v1/vehicles to view vehicle availability and POST /api/v1/vehicles/book to make bookings.
Billing Service: POST /api/v1/payment/confirm to confirm payments and generate invoices.

Why Microservices Architecture?
1. Scalability
Microservices can be independently scaled, so each service can handle varying loads. For example, the Billing Service can be scaled more during high traffic periods (e.g., during sale seasons) without affecting the Vehicle Service or User Service.

2. Fault Isolation
Since each service is isolated, failure in one service does not bring down the entire system. For example, if the User Service goes down, the Vehicle Service and Billing Service can still operate independently, ensuring better system reliability.

3. Technology Agnostic
Each microservice can use the technology that best suits its function. For example, the User Service might use a relational database, while the Vehicle Service might use a NoSQL database.

4. Faster Development
With separate services, multiple teams can work on different parts of the system simultaneously. This leads to faster development and easier maintenance, as each service can be deployed independently.

5. Flexibility
Microservices allow for more flexibility in terms of deployment and updates. If you need to add a new feature, you can do so by deploying just the affected service, without needing to redeploy the entire system.

Conclusion
This architecture ensures that each part of the system is decoupled, fault-tolerant, and independently scalable. The use of RESTful APIs allows the system to grow seamlessly, and the modular design makes it easier to maintain and extend. By adopting Microservices Architecture, we achieve greater flexibility and scalability for the future growth of the vehicle rental system.

