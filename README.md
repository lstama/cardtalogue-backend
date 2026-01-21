# CRUD API with Authentication and Role-Based Access Control

A comprehensive CRUD API built with Go, featuring authentication, role-based authorization, database migrations, and clean architecture.

## Features

- 🔐 JWT-based Authentication & Authorization
- 👥 Role-based Access Control (Admin & Editor)
- 🏗️ Clean Architecture (Handler -> Service -> Repository pattern)
- 🗄️ Database migrations with up/down support
- 📝 CRUD operations for Users and Items
- ✅ Input validation
- 🚀 RESTful API design
- 📊 Pagination support
- 🐳 Docker support for database
- 🧪 Comprehensive error handling
- 📝 Structured logging

## User Roles & Permissions

### Admin
- ✅ Create, read, update, delete users
- ✅ Create, read, update, delete items
- ✅ Full access to all endpoints

### Editor  
- ✅ Create, read, update items
- ✅ Can disable items (status: disabled)
- ❌ Cannot delete items (only admins can)
- ❌ Cannot manage users

### Public
- ✅ View items list (no authentication required)

## Default Admin Account

The system automatically creates a default admin account:
- **Email**: admin@example.com
- **Password**: admin123
- **Role**: admin

**⚠️ IMPORTANT**: Change the default password after first login!

## API Endpoints

### Authentication

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@example.com",
  "password": "admin123"
}
```

**Response:**
```json
{
  "data": {
    "user": {
      "id": 1,
      "name": "Administrator",
      "email": "admin@example.com",
      "role": "admin",
      "age": 30
    },
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

#### Refresh Token
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

#### Get Profile
```http
GET /api/v1/profile
Authorization: Bearer <access_token>
```

#### Change Password
```http
PUT /api/v1/profile/password
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "current_password": "admin123",
  "new_password": "newpassword123"
}
```

### User Management (Admin Only)

#### Create User
```http
POST /api/v1/users
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "John Editor",
  "email": "john@example.com",
  "password": "password123",
  "role": "editor",
  "age": 30
}
```

#### Get All Users
```http
GET /api/v1/users?limit=20&offset=0
Authorization: Bearer <admin_token>
```

#### Get User by ID
```http
GET /api/v1/users/{id}
Authorization: Bearer <admin_token>
```

#### Update User
```http
PUT /api/v1/users/{id}
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "John Updated",
  "role": "admin"
}
```

#### Delete User
```http
DELETE /api/v1/users/{id}
Authorization: Bearer <admin_token>
```

### Items Management

#### Get All Items (Public - No Auth Required)
```http
GET /api/v1/items?limit=20&offset=0
```

#### Get Item by ID (Authenticated Users)
```http
GET /api/v1/items/{id}
Authorization: Bearer <access_token>
```

#### Create Item (Editor/Admin)
```http
POST /api/v1/items
Authorization: Bearer <editor_or_admin_token>
Content-Type: application/json

{
  "name": "Sample Item",
  "description": "This is a sample item",
  "status": "enabled"
}
```

#### Update Item (Editor/Admin)
```http
PUT /api/v1/items/{id}
Authorization: Bearer <editor_or_admin_token>
Content-Type: application/json

{
  "name": "Updated Item",
  "description": "Updated description",
  "status": "disabled"
}
```

#### Delete Item (Admin Only)
```http
DELETE /api/v1/items/{id}
Authorization: Bearer <admin_token>
```

## Quick Start

### 1. Clone and Setup
```bash
git clone <repository>
cd cardtalogue-backend
cp .env.example .env  # Edit with your configuration
```

### 2. Start Database
```bash
make docker-up
```

### 3. Install Dependencies
```bash
make dev-setup
```

### 4. Run Migrations
```bash
make migrate-up
```

### 5. Start Application
```bash
make run
```

The API will be available at `http://localhost:8080`

## Authentication Flow

1. **Login** with email/password to get access and refresh tokens
2. **Include Bearer token** in Authorization header for protected endpoints
3. **Refresh token** when access token expires
4. **Change password** as needed through the profile endpoint

## Example Usage Flow

```bash
# 1. Login as admin
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}'

# 2. Create an editor user (save the token from step 1)
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"John Editor","email":"john@example.com","password":"password123","role":"editor","age":25}'

# 3. Create an item (can be done by editor or admin)
curl -X POST http://localhost:8080/api/v1/items \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Item","description":"A test item","status":"enabled"}'

# 4. View items (public endpoint - no auth needed)
curl http://localhost:8080/api/v1/items

# 5. Editor can disable an item
curl -X PUT http://localhost:8080/api/v1/items/1 \
  -H "Authorization: Bearer <editor_token>" \
  -H "Content-Type: application/json" \
  -d '{"status":"disabled"}'

# 6. Only admin can delete items
curl -X DELETE http://localhost:8080/api/v1/items/1 \
  -H "Authorization: Bearer <admin_token>"
```

## Database Schema

### Users Table
- `id` - Primary key
- `name` - User's full name
- `email` - Unique email address
- `password` - Hashed password
- `role` - admin or editor
- `age` - User's age
- `created_at`, `updated_at`, `deleted_at` - Timestamps

### Items Table  
- `id` - Primary key
- `name` - Item name
- `description` - Item description
- `status` - enabled or disabled
- `created_at`, `updated_at`, `deleted_at` - Timestamps

## Security Features

- ✅ Password hashing with bcrypt
- ✅ JWT tokens with expiration
- ✅ Role-based access control
- ✅ Input validation
- ✅ SQL injection protection via GORM
- ✅ Soft deletes

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

TODO:

Assign default pack and language, for now ignore them (or maybe pack only?)