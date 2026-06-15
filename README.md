# SUBIANGO FRAMEWORK

<p align="center">
  <img src="docs/subian-go.png" width="150">
</p>
> Production-Ready Modular Backend REST API
> Built with Go and Echo v5

| Item       | Value                                      |
| ---------- | ------------------------------------------ |
| Repository | https://github.com/arindyproject/subian-go |
| License    | MIT                                        |
| Language   | Go 1.21+                                   |
| Framework  | Echo v5                                    |
| ORM        | GORM                                       |
| Database   | PostgreSQL + Redis                         |

---

# Table of Contents

1. Overview
2. Key Features
3. Tech Stack
4. Project Structure
5. Getting Started
6. Environment Variables
7. Makefile Commands
8. Core Modules
   - Authentication
   - User Management
   - RBAC
9. Code Generator
10. Architecture & Design Patterns
11. RBAC Usage Examples
12. API Documentation (Swagger)
13. Contributing
14. License

---

# 1. Overview

**SubianGO** adalah framework backend REST API modular yang siap digunakan untuk production, dibangun menggunakan **Go** dan **Echo v5**.

Dirancang untuk menghilangkan masalah _blank project syndrome_, SubianGO sudah menyediakan:

- Authentication
- User Management
- RBAC (Role-Based Access Control)
- Module Generator
- Swagger Documentation
- Database Migration & Seeder

Cukup clone repository, konfigurasi environment, lalu fokus membangun fitur bisnis tanpa harus membuat fondasi aplikasi dari nol.

---

# 2. Key Features

## 🔐 Authentication

- JWT Authentication
- Refresh Token
- Login Rate Limiting (Redis)
- Secure Password Hashing

## 👥 User Management

- Full CRUD User
- Profile Settings
- Photo Upload
- Password Reset

## 🛡 RBAC

- Roles
- Permissions
- Direct User Permission
- Superadmin Bypass

## ⚡ Code Generator

Generate module dan submodule secara otomatis melalui CLI tanpa boilerplate.

## 🗄 Database

- GORM Auto Migration
- SQL Migration
- Seeder
- Fresh Migration

## 📚 Documentation

- Swagger/OpenAPI
- Auto Generated API Docs

## 🏗 Architecture

- Clean Architecture
- Separation of Concerns
- Modular Design

---

# 3. Tech Stack

| Component | Technology       |
| --------- | ---------------- |
| Language  | Go 1.21+         |
| Framework | Echo v5          |
| ORM       | GORM             |
| Database  | PostgreSQL       |
| Cache     | Redis            |
| API Docs  | Swagger / Swaggo |

---

# 4. Project Structure

```text
subian-go/
├── cmd/
│   ├── api/
│   ├── migrate/
│   └── seed/
│
├── config/
│
├── internal/
│   ├── apps/
│   ├── modules/
│   │   ├── auth/
│   │   ├── users/
│   │   └── rbac/
│   │
│   └── shared/
│
├── docs/
├── generator.go
├── Makefile
└── go.mod
```

---

# 5. Getting Started

## Prerequisites

- Go 1.21+
- PostgreSQL
- Redis
- GNU Make

## Installation

### Clone Repository

```bash
git clone https://github.com/yourusername/subian-go.git
cd subian-go
```

### Install Dependency

```bash
go mod tidy
```

### Setup Environment

```bash
cp config/.env.dev .env
```

Edit file `.env` sesuai konfigurasi database dan redis.

### Run Migration & Seeder

```bash
make migrate-seed
```

### Start Server

```bash
make run
```

API akan berjalan pada:

```text
http://localhost:1323
```

---

# 6. Environment Variables

| Variable                     | Default               | Description             |
| ---------------------------- | --------------------- | ----------------------- |
| ENV                          | development           | Environment             |
| SERVER_PORT                  | 1323                  | Server Port             |
| BASE_URL                     | http://localhost:1323 | Base URL                |
| DATABASE_URL                 | postgres://...        | PostgreSQL DSN          |
| DB_HOST                      | localhost             | Database Host           |
| DB_PORT                      | 5432                  | Database Port           |
| DB_USER                      | postgres              | Database User           |
| DB_PASSWORD                  | -                     | Database Password       |
| DB_NAME                      | neosim                | Database Name           |
| DB_SSL_MODE                  | disable               | SSL Mode                |
| REDIS_HOST                   | localhost             | Redis Host              |
| REDIS_PORT                   | 6379                  | Redis Port              |
| JWT_SECRET                   | auto-generated        | JWT Secret              |
| JWT_ISSUER                   | neosim                | JWT Issuer              |
| JWT_ACCESS_TOKEN_EXP_MINUTES | 15                    | Access Token Expiry     |
| JWT_REFRESH_TOKEN_EXP_DAYS   | 7                     | Refresh Token Expiry    |
| PASSWORD_MIN_LENGTH          | 6                     | Minimum Password Length |
| IS_REGISTRATION_ACTIVE       | true                  | Public Registration     |
| AUTO_ACTIVE_USER             | true                  | Auto Activate User      |

---

# 7. Makefile Commands

## Build & Run

```bash
make run
make build
make clean
```

## Database & Migration

```bash
make migrate-dev
make migrate-prod
make migrate-sql
make migrate-sql-prod
make migrate-fresh-dev
make migrate-fresh-prod
make seed
make seed-prod
make migrate-seed
make migrate-fresh-seed
make create-migration
```

## Generator & Documentation

```bash
make gen-module name=artikel
make gen-module name=artikel add=kategori

make swagger-install
make swagger-gen
make swagger-fmt
make run-api
```

## Testing & Security

```bash
make test
make test-auth
make test-users
make test-rbac

make gen-jwt-dev
make gen-jwt-prod
```

---

# 8. Core Modules

## 8.1 Authentication

Base Route:

```text
/api/v1/auth
```

| Method | Endpoint         |
| ------ | ---------------- |
| POST   | /login           |
| POST   | /register        |
| POST   | /refresh         |
| POST   | /forgot-password |
| POST   | /reset-password  |
| POST   | /logout          |
| POST   | /logout-all      |

---

## 8.2 User Management

Base Route:

```text
/api/v1/users
```

| Method | Endpoint             |
| ------ | -------------------- |
| GET    | /                    |
| GET    | /:id                 |
| GET    | /username/:username  |
| POST   | /                    |
| PUT    | /:id                 |
| DELETE | /:id                 |
| GET    | /deleted             |
| PUT    | /:id/change-password |
| POST   | /:id/reset-password  |
| GET    | /:id/settings        |
| PUT    | /:id/settings        |
| PUT    | /:id/photo           |
| DELETE | /:id/photo           |

### Authorization

```http
Authorization: Bearer <access_token>
```

---

## 8.3 RBAC

Base Route:

```text
/api/v1/rbac
```

### Permissions

| Method | Endpoint         |
| ------ | ---------------- |
| GET    | /permissions     |
| POST   | /permissions     |
| GET    | /permissions/:id |
| PUT    | /permissions/:id |
| DELETE | /permissions/:id |

### Roles

| Method | Endpoint   |
| ------ | ---------- |
| GET    | /roles     |
| POST   | /roles     |
| GET    | /roles/:id |
| PUT    | /roles/:id |
| DELETE | /roles/:id |

### Role Permissions

```text
POST   /roles/:id/permissions
PUT    /roles/:id/permissions
DELETE /roles/:id/permissions
```

### User Roles

```text
GET    /users/:user_id/roles
POST   /users/:user_id/roles
PUT    /users/:user_id/roles
DELETE /users/:user_id/roles
```

### User Permissions

```text
GET    /users/:user_id/permissions
POST   /users/:user_id/permissions
```

> Superadmin otomatis melewati seluruh pemeriksaan permission.

---

# 9. Code Generator

Generate module utama:

```bash
make gen-module name=artikel
```

Generate sub-module:

```bash
make gen-module name=artikel add=kategori
```

Folder yang dihasilkan:

```text
contracts/
dto/
handlers/
services/
repositories/
migrations/
```

---

# 10. Architecture & Design Patterns

SubianGO menerapkan **Clean Architecture**.

Struktur layer:

```text
contracts
dto
handlers
services
repositories
migrations
middlewares
```

Prinsip:

- Dependency Inversion
- Separation of Concerns
- Modular Design
- Testability

---

# 11. RBAC Usage Examples

## Route Protection

```go
protected.PUT("/:id", h.UpdateUserHandler,
    rbacMiddlewares.RequireSelfOrPermission(
        rbacRepo,
        rbacModels.PermUsersUpdate,
    ),
)
```

```go
protected.DELETE("/:id", h.DeleteUserHandler,
    rbacMiddlewares.RequirePermission(
        rbacRepo,
        rbacModels.PermUsersDelete,
    ),
)
```

## Available Middleware

- RequirePermission()
- RequireAnyPermission()
- RequireRole()
- RequireSuperadmin()
- RequireSelf()
- RequireSelfOrPermission()
- RequireSelfOrRole()

---

# 12. API Documentation (Swagger)

Swagger UI:

```text
http://localhost:1323/swagger/index.html
```

Generate Documentation:

```bash
make swagger-gen
```

Format Comments:

```bash
make swagger-fmt
```

---

# 13. Contributing

1. Fork Repository
2. Create Feature Branch

```bash
git checkout -b feature/AmazingFeature
```

3. Commit Changes

```bash
git commit -m "Add some AmazingFeature"
```

4. Push Branch

```bash
git push origin feature/AmazingFeature
```

5. Open Pull Request

---

# 14. License

Distributed under the MIT License.

See:

```text
LICENSE
```

for more information.

---

<div align="center">

Made with AI and Coffee

**[Aji A.A]**

GitHub: https://github.com/arindyproject

</div>
