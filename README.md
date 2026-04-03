# finance_api_backend

## Finance Management API
A backend solution for finance management built with **Go**, featuring role-based access control (RBAC), rate limiting, JWT authentication, and containerization.

---

## 🚀 Technologies
* **Language:** Go 1.24+
* **Framework:** Chi Router
* **Database:** PostgreSQL 15
* **Migrations:** golang-migrate
* **Containerization:** Docker & Docker Compose

---

## 🛠️ Getting Started
The project is fully dockerized. You don't need Go or PostgreSQL installed locally - only **Docker**.

### 1. Run the application
```bash
docker-compose up --build
```

### What happens automatically:
* **Starts the PostgreSQL 15 container.**
* **Implements a Retry Mechanism** to ensure the database is ready before the API starts.
* **Executes all Database Migrations** from the `migrate/migrations` directory.
* **Starts the API server** on port `3000`.

---


### 🔑 Permission Levels

The system implements a permission structure based on user roles:

* **Viewer:** Read-only access. Authorized to view basic information but **cannot** create, modify, or delete records.
* **Analyst:** Data-focused access. Authorized to read records and view trend analysis.
* **Admin:** Full management access. Authorized to create, update, and delete records, as well as manage user accounts and roles.

---

## 🛣️ API Endpoints
The API is accessible at `http://localhost:3000/v1/`.

### Public & Auth
| Method | Endpoint | Description |
|:-------|:---------|:------------|
| **GET** | `/health` | Check API and Database connectivity |
| **POST** | `/users/register` | Register a new account |
| **POST** | `/users/login` | User authentication and JWT generation |

### User Management (Admin Only)
| Method | Endpoint | Description |
|:-------|:---------|:------------|
| **GET** | `/users/` | Get all users |
| **PATCH** | `/users/{id}/status` | Activate or deactivate user accounts |
| **PATCH** | `/users/{id}/role` | Change user role (Admin, Analyst, Viewer) |
| **DELETE** | `/users/{id}` | Remove user from system |


### Finance Records
| Method | Endpoint | Description | Role |
|:-------|:---------|:------------|:-----|
| **GET** | `/finance/summary` | Get dashboard financial summary | **User** |
| **GET** | `/finance/list_records` | List all financial records. Supports filtering and pagination | **User** |
| **GET** | `/finance/trends` | Get financial trend analysis | **Analyst/Admin** |
| **POST** | `/finance/create_record` | Create a new record | **Admin** |
| **PUT** | `/finance/update_record/{id}` | Update existing record | **Admin** |
| **DELETE** | `/finance/delete_record/{id}` | Delete a record | **Admin** |

> **Note:** Full documentation is available at `http://localhost:3000/swagger/index.html` once the app is running.

> **Note:** Some endpoints would need to be redesigned for production purposes. For example, the endpoint `/finance/trends` currently retrieves data per month. A production version would use `generate_series` in SQL to fill empty months with default values to avoid gaps in the result set.

---

## 📂 Project Structure
* **`/cmd/api`**: Application entry point (`main.go`), handlers, and middleware logic.
* **`/migrate/migrations`**: SQL schema files (Up/Down).
* **`/internal`**: Core layers (Database, Authenticator, Rate-limiters, Config parser).
* **`/docs`**: Swagger documentation and API specs.

---

## 💡 Key Features
* **RBAC (Role-Based Access Control):** Middleware to enforce permissions (`Admin`, `Analyst`, `Viewer`).
* **Rate Limiting:** Protects endpoints using a Fixed Window and Token rate limiters.
* **Automated Migrations:** Database schema is automatically updated on every startup.
* **Validation:** Strict input validation using `go-playground/validator`.

---

## 🧪 Testing
To run the test suite locally:

```bash
go test -v ./...
```

<br>

*Developed by Drazen Petrovic*