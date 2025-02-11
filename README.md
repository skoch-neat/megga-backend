# MEGGA (Backend)

## **Project Overview**
MEGGA (Monitoring Economic Goods & Government Advocacy) is designed to automate political advocacy by monitoring changes in common household goods and economic indicators as reported by the Bureau of Labor Statistics API. Users can create thresholds, and when these are triggered, emails are automatically sent to their specified political representatives. Users can also opt in to receive an email notification when a threshold is met, encouraging further advocacy efforts.

### **Proof of Concept & Security Considerations**
This project is a **proof of concept**, meaning that the full email automation system is not configured to send emails to actual government representatives.

The MEGGA backend is built with Go and provides APIs for user authentication, threshold management, and data tracking. It integrates with PostgreSQL as the database and utilizes Gorilla Mux for routing.

---

## Key Features

- **User Authentication**: Signup and login via AWS Cognito and JSON Web Tokens.
- **Threshold Management**: Full CRUD operations for thresholds.
- **Data Tracking**: Integrates with third-party APIs (BLS) to fetch and store data.
- **Notification System**: Alerts users who opt in when thresholds are breached.
- **RESTful API**: Built with Gorilla Mux for structured routing.
- **PostgreSQL Database**: Reliable persistent storage.
- **Environment Configurations**: Customizable via `.env` files.
- **Content Security Policy (CSP)**: Configurable for API security.
- **Cross-Origin Resource Sharing (CORS)**: Allows frontend-backend communication.
- **Development Utilities**: Database migration and seeding for testing.
- **Extensible Design**: Easily adaptable for future needs.

---

## **Project Structure**

```
megga-backend/
│   ├── cmd/
│   │   ├── devutils/
│   │   │   ├── main.go
│   │   ├── web/
│   │   │   ├── main.go
│   ├── handlers/
│   │   ├── data.go
│   │   ├── notifications.go
│   │   ├── recipients.go
│   │   ├── threshold_recipients.go
│   │   ├── thresholds.go
│   │   ├── users.go
│   ├── internal/
│   │   ├── config/
│   │   │   ├── bls.go
│   │   │   ├── env.go
│   │   ├── database/
│   │   │   ├── database.go
│   │   │   ├── interface.go
│   │   │   ├── pgx_db_wrapper.go
│   │   ├── devutils/
│   │   │   ├── migrate.go
│   │   │   ├── seeder.go
│   │   ├── middleware/
│   │   │   ├── cognito.go
│   │   │   ├── cors.go
│   │   │   ├── csp.go
│   │   │   ├── logging.go
│   │   ├── models/
│   │   │   ├── data.go
│   │   │   ├── notification.go
│   │   │   ├── recipient.go
│   │   │   ├── threshold_recipient.go
│   │   │   ├── threshold.go
│   │   │   ├── user.go
│   │   ├── router/
│   │   │   ├── router.go
│   │   ├── routes/
│   │   │   ├── routes.go
│   │   ├── services/
│   │   │   ├── bls.go
│   │   │   ├── data.go
│   │   │   ├── notification.go
│   │   │   ├── threshold_monitor.go
│   │   ├── templates/
│   │   │   ├── recipient_notification_bad.txt
│   │   │   ├── recipient_notification_good.txt
│   │   │   ├── user_notification.txt
│   │   ├── utils/
│   │   │   ├── utils.go
│   ├── tests/
│   │   ├── database_test/
│   │   │   ├── database_test.go
│   │   ├── handlers_test/
│   │   │   ├── data_test.go
│   │   │   ├── notifications_test.go
│   │   │   ├── recipients_test.go
│   │   │   ├── thresholds_test.go
│   │   │   ├── users_test.go
│   │   ├── routes_test/
│   │   │   ├── routes_test.go
│   │   ├── services_test/
│   │   │   ├── bls_service_test.go
│   │   │   ├── data_service_test.go
│   ├── testutils/
│   │   ├── mockdb_wrapper.go
│   │   ├── mockjwt_token.go
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── LICENSE
├── README
├── render.yaml
```

---

## **Setup Instructions**

### **1. Prerequisites**
Ensure you have the following installed:
- Go (1.23.4 or higher)
- PostgreSQL (14.0 or higher)

---

### **2. Clone the Repository**

- On GitHub, navigate to the main page of the repository. Above the list of files, click `<> Code`.
- Copy the URL for the repository.
- Open your terminal and change to the location where you want the cloned directory.
- Run the following commands to clone the repository and navigate to the project directory:

```sh
git clone <repository-url>
cd megga-backend
```

- Press Enter to create your local clone.

---

### **3. Install Dependencies**
Install required dependencies for the project:

    go mod tidy

---

### **4. Configure Environment Variables**
Copy the `.env.example` file and configure your variables:

#### Variables expected in the `.env`:

  - `API_BASE_URL=<backend_url>` (e.g. `http://localhost:8080` for local development or `https://api.yourdomain.com` for production)
  - `APP_ENV=development` (any other value will turn off debug mode)
  - `AWS_REGION=<your_aws_region>`
  - `BLS_API_KEY=<your_bls_api_key>`
  - `BLS_API_URL=https://api.bls.gov/publicAPI/v2/timeseries/data/`
  - `BLS_INIT=false` (Set to true to initialize data fetch at startup, false to bypass)
  - `COGNITO_CLIENT_ID=<your_cognito_client_id>`
  - `COGNITO_DOMAIN=https://<your_cognito_domain>`
  - `COGNITO_IDP_URL=https://<your_cognito_idp_url>`
  - `COGNITO_TOKEN_URL=https://<your_cognito_token_url>`
  - `COGNITO_USER_POOL_ID=<your_cognito_user_pool_id>`
  - `DATABASE_URI=postgres://<username>:<password>@<host>:<port>/<database_name>`
  - `FRONTEND_URL=<frontend_url>` (e.g., `http://localhost:5173` for local development or `https://www.yourdomain.com` for production)
  - `MOCK_JWT_TOKEN=<your_mock_json_web_token>`
  - `PORT=8080`

**Tip**: The `.env.example` file contains placeholders for all required variables. Copy it to `.env` and replace placeholders with your actual configuration values.

---

### **5. Set Up the Database**

#### **Run Migrations**
To apply schema migrations, run:

    go run cmd/devutils/main.go --migrate

#### **Seed the Database**
To populate the development database with sample data, run:

    go run cmd/devutils/main.go --seed

**Note**: Utilities are intended for development purposes only.

---

### **6. Run the Server**
Start the server with the following command:

    go run cmd/web/main.go

The server will start at `http://localhost:8080` by default. The port can be customized in the `.env` file.

---

## **Security**

### **Environment Variables**
Keep your `.env` file secure and do not commit it to version control. Use `.env.example` as a template for collaborators.

### **CSP and CORS**
- CSP headers are dynamically configured in `router.go` to secure API responses.
- CORS middleware is enabled to allow requests from the frontend.

---

## **API Endpoints**

### **Data Routes**
- `POST /data` - Create a new data entry.
- `GET /data` - Retrieve all economic data entries.
- `GET /data/{id}` - Fetch a specific data entry by ID.
- `PUT /data/{id}` - Update an existing data entry.
- `DELETE /data/{id}` - Delete a data entry.

---

### **Notifications Routes**
- `POST /notifications` - Create a new notification.
- `GET /notifications` - Retrieve all notifications.
- `GET /notifications/{id}` - Fetch a specific notification by ID.
- `PUT /notifications/{id}` - Update an existing notification.
- `DELETE /notifications/{id}` - Remove a notification.

---

### **Recipients Routes**
- `POST /recipients` - Add a new recipient.
- `GET /recipients` - Retrieve all recipients.
- `GET /recipients/{id}` - Fetch a specific recipient by ID.
- `PUT /recipients/{id}` - Update recipient details.
- `DELETE /recipients/{id}` - Remove a recipient.

---

### **Threshold Recipients Routes**
- `POST /threshold_recipients` - Assign a recipient to a threshold.
- `GET /threshold_recipients` - Retrieve all threshold-recipient relationships.
- `GET /threshold_recipients/{threshold_id}/{recipient_id}` - Get a specific threshold-recipient relationship.
- `PUT /threshold_recipients/{threshold_id}/{recipient_id}` - Update a threshold-recipient relationship.
- `DELETE /threshold_recipients/{threshold_id}/{recipient_id}` - Remove a recipient from a threshold.

---

### **Thresholds Routes**
- `POST /thresholds` - Create a new threshold.
- `GET /thresholds/{id}` - Fetch details of a specific threshold.
- `PUT /thresholds/{id}` - Update an existing threshold.
- `DELETE /thresholds/{id}` - Remove a threshold.

---

### **User Routes**
- `POST /users` - Create a new user.
- `GET /users/{email}` - Fetch user details by email.
- `GET /users/{userId}/thresholds` - Get all thresholds for a specific user.
- `DELETE /users/{userId}/thresholds` - Delete all thresholds for a specific user.

---

## **Development Utilities**

### **Migrate the Database**
To apply schema migrations, run:

    go run cmd/devutils/main.go --migrate

### **Seed the Database**
To populate the database with sample data, run:

    go run cmd/devutils/main.go --seed

### **Note**
Migration and seeding scripts are for development purposes only and should not be run in production.

---

## **Contributing**

At this time, contributions are not being accepted. This project is intended for educational purposes and is shared for review and feedback.

---

## **License**

This project is licensed under the MIT License.
