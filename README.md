# MEGGA Backend

The MEGGA backend is built with Go and provides APIs for user authentication, threshold management, and data tracking. It integrates with PostgreSQL as the database and utilizes Gorilla Mux for routing.

---

## Key Features

- **User Authentication**: Signup and login via AWS Cognito.
- **Threshold Management**: Full CRUD operations for thresholds.
- **Data Tracking**: Integrates with third-party APIs (BLS, FRED) to fetch and store data.
- **Notification System**: Alerts users when thresholds are breached.
- **RESTful API**: Built with Gorilla Mux for structured routing.
- **PostgreSQL Database**: Reliable persistent storage.
- **Environment Configurations**: Customizable via `.env` files.
- **Development Utilities**: Database migration and seeding for testing.
- **Extensible Design**: Easily adaptable for future needs.

---

## **Project Structure**

megga-backend/
- devutils/
  - migrate.go (Handles database migrations)
  - seed.go (Seeds the database with test data)
- models/
  - data.go (Data model for third-party API data)
  - notification.go (Data model for notifications)
  - recipient.go (Data model for recipients)
  - threshold.go (Data model for thresholds)
  - threshold_recipient.go (Join table for thresholds and recipients)
  - user.go (Data model for users)
- routes/
  - user_routes.go (Handlers for user-related API endpoints)
- services/
  - database.go (Database connection logic)
  - router.go (Router initialization logic)
- .env.example (Environment variable template)
- go.mod (Go module definition)
- go.sum (Dependency checksums)
- main.go (Entry point of the backend application)
- README.md (Documentation)

---

## **Setup Instructions**

### **1. Prerequisites**
Ensure you have the following installed:
- Go (1.23.4 or higher)
- PostgreSQL (14.0 or higher)

---

### **2. Clone the Repository**
Run the following commands to clone the repository and navigate to the project directory:

    git clone <repository-url>
    cd megga-backend

---

### **3. Install Dependencies**
Install required dependencies for the project:

    go mod tidy

---

### **4. Configure Environment Variables**
Copy the `.env.example` file to create your `.env` file:

    cp .env.example .env

Variables expected in the .env:
- DATABASE_URI=postgres://username:password@localhost:5432/megga_dev
- PORT=8080
- COGNITO_CLIENT_ID=your-cognito-client-id
- BLS_API_KEY=your-bls-api-key
- FRED_API_KEY=your-fred-api-key

Update the variables with your information as indicated.

---

### **5. Set Up the Database**

#### **Run Migrations**
To apply schema migrations, run:

    go run cmd/devutils/main.go --migrate

#### **Seed the Database**
To populate the database with sample data, run:

    go run cmd/devutils/main.go --seed

**Note**: These utilities are for development purposes only and should not be used in production.

---

### **6. Run the Server**
Start the server with the following command:

    go run main.go

The server will start at `http://localhost:8080` by default. The port can be changed in the `.env` file.

---

## **API Endpoints**

### **User Routes**
- `GET /users`: Retrieve a list of all users.

---

## **Development Utilities**

### **Migrate the Database**
To apply schema migrations, run:

    go run cmd/devutils/main.go --migrate

### **Seed the Database**
To populate the database with sample data, run:

    go run cmd/devutils/main.go --seed

### **Note**
- Migration and seeding scripts are for development purposes only and should not be run in production.

---

## **Contributing**

At this time, contributions are not being accepted. This project is intended for educational purposes and is shared for review and feedback.

---

## **License**

This project is licensed under the MIT License.
