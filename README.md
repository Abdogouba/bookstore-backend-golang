# 📚 Bookstore Backend API

A RESTful backend application for an online bookstore built with Go. This project follows Clean Architecture principles and implements authentication, authorization, book management, and order management features.

> **Note:** This project is currently in progress. Most features have been completed, and the remaining order-related functionality is under development. ChatGPT was used as an assistant during the design and implementation process.

---

## 🚀 Tech Stack

* **Go (Golang)**
* **Gin** (HTTP Framework)
* **GORM** (ORM)
* **PostgreSQL** (run in docker)
* **JWT Authentication & Authorization**
* **Docker**
* **Swagger/OpenAPI**
* **Integration Testing**
* **File Upload Handling**
* **Environment Variables**
* **Database Transactions**
* **Database Indexes**
* **Password Hashing**
* **Pagination**
* **DTO Pattern**
* **Clean Architecture**

---

## 🏗️ Architecture

The project follows a Clean Architecture approach:

```
├── cmd
├── config
├── docs
├── internal
│   ├── handlers
│   ├── services
│   ├── repositories
│   ├── models
│   ├── middleware
│   └── dto
├── uploads
├── tests
└── ...
```

Main layers:

* **Handlers** → HTTP layer
* **Services** → Business logic
* **Repositories** → Database access
* **Models** → Database entities
* **Middleware** → Authentication & Authorization

---

## 📋 Business Assumptions

* Only **one admin** exists in the system.
* Admin account is created through database seeding.
* Shopping cart functionality is handled by the frontend.
* All orders are **Cash On Delivery (COD)**.
* Order statuses:

```
Pending
Confirmed
Out For Delivery
Delivered
Cancelled
```

---

## 🗄️ Database Design

### Users

| Field         |
| ------------- |
| id            |
| name          |
| email         |
| password_hash |
| phone_number  |
| role          |
| created_at    |
| updated_at    |
| deleted_at    |

### Refresh Tokens

| Field      |
| ---------- |
| id         |
| user_id    |
| token      |
| expires_at |
| created_at |
| deleted_at |

### Books

| Field      |
| ---------- |
| id         |
| title      |
| author     |
| publisher  |
| category   |
| price      |
| stock      |
| image_path |
| created_at |
| updated_at |
| deleted_at |

### Orders

| Field       |
| ----------- |
| id          |
| user_id     |
| status      |
| address     |
| total_price |
| created_at  |
| updated_at  |
| deleted_at  |

### Order Items

| Field      |
| ---------- |
| id         |
| order_id   |
| book_id    |
| quantity   |
| price      |
| title      |
| author     |
| publisher  |
| image_path |

---

## ✅ Features

### Authentication & User Management

* [x] User Registration
* [x] User Login
* [x] Refresh Token
* [x] Logout
* [x] Authentication Middleware
* [x] Authorization Middleware
* [x] View Profile
* [x] Edit Profile
* [x] Change Password
* [x] Soft Delete Profile

### Admin Features

* [x] Admin Database Seeding
* [x] Add Book
* [x] Update Book
* [x] Soft Delete Book

### Books

* [x] View All Books

  * Search
  * Filtering
  * Sorting
  * Pagination
* [x] View Single Book

### Orders

* [ ] Create Order
* [ ] View My Orders
* [ ] View Single Order
* [ ] Admin View Orders
* [ ] Admin View Single Order
* [ ] Admin Update Order Status
* [ ] Cancel Pending Order

---

## 🔐 Authentication

The API uses:

* Access Tokens (JWT)
* Refresh Tokens
* Role-Based Authorization

Roles:

* `admin`
* `user`

---

## 📖 API Documentation

Swagger documentation is available after running the project:

```
/swagger/index.html
```

---

## 🧪 Testing

Integration tests are implemented for all API endpoints, covering all important test cases.
