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
- **Content Security Policy (CSP)**: Configurable for API security.
- **Cross-Origin Resource Sharing (CORS)**: Allows frontend-backend communication.
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
  - env.go (Environment variable handling)
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

#### Variables expected in the `.env`:

- **Database Configuration**:
  - `DATABASE_URI=postgres://<username>:<password>@<host>:<port>/<database_name>`

- **Server Configuration**:
  - `PORT=8080`

- **Application URLs**:
  - `API_BASE_URL=<backend_url>` (e.g., `http://localhost:8080` for local development or `https://api.yourdomain.com` for production)
  - `FRONTEND_URL=<frontend_url>` (e.g., `http://localhost:5173` for local development or `https://www.yourdomain.com` for production)

- **Cognito Configuration**:
  - `COGNITO_CLIENT_ID=<your_cognito_client_id>`
  - `COGNITO_DOMAIN=https://<your_cognito_domain>`
  - `COGNITO_IDP_URL=https://<your_cognito_idp_url>`
  - `COGNITO_TOKEN_URL=https://<your_cognito_token_url>`

- **API Keys**:
  - `BLS_API_KEY=<your_bls_api_key>`
  - `FRED_API_KEY=<your_fred_api_key>`

**Tip**: The `.env.example` file contains placeholders for all required variables. Copy it to `.env` and replace placeholders with your actual configuration values.

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

## **Security**

### **Environment Variables**
Keep your `.env` file secure and do not commit it to version control. Use `.env.example` as a template for collaborators.

### **CSP and CORS**
- CSP headers are dynamically configured in `router.go` to secure API responses.
- CORS middleware is enabled to allow requests from the frontend.

---

### **6. Run the Server**
Start the server with the following command:

    go run main.go

The server will start at `http://localhost:8080` by default. The port can be changed in the `.env` file.

---

## **API Endpoints**

### **User Routes**
- `GET /api/users`: Retrieve a list of all users.
- `POST /api/users`: Create a new user.

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
