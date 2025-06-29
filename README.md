# Ticketing System API

A comprehensive ticketing system built with Go, Gin, GORM, and MySQL. This system provides user management, event management, ticket sales, and administrative reporting with role-based access control.

## Features

### 🔐 Authentication & Authorization

- JWT-based authentication
- Role-based access control (Admin/User)
- Secure password hashing with bcrypt
- Auto-seeded admin account

### 👥 User Management

- User registration and login
- Profile management
- Admin user management
- Account activation/deactivation

### 🎫 Event Management

- CRUD operations for events
- Event categorization and filtering
- Capacity management
- Event status tracking (Active, Ongoing, Completed, Cancelled)
- Advanced search and filtering

### 🎟️ Ticket Management

- Ticket purchasing with validation
- Ticket cancellation (with business rules)
- Ticket status management
- User ticket history
- Inventory management

### 📊 Reporting & Analytics

- Sales summary reports
- Event-specific reports
- Revenue tracking
- User statistics

### 🔍 Advanced Features

- **Pagination**: Efficient data pagination for all list endpoints
- **Search**: Full-text search across multiple fields
- **Filtering**: Advanced filtering by multiple criteria
- **Validation**: Comprehensive input validation
- **Error Handling**: Standardized error responses
- **API Documentation**: Auto-generated Swagger documentation

## Architecture

This project follows **Clean Architecture** principles with clear separation of concerns:

```
├── entity/          # Business entities and data models
├── repository/      # Data access layer
├── service/         # Business logic layer
├── controller/      # HTTP handlers and routing
├── middleware/      # Authentication, CORS, validation
├── config/          # Configuration and database setup
└── main.go          # Application entry point
```

## Tech Stack

- **Framework**: Gin (HTTP web framework)
- **ORM**: GORM (Object-relational mapping)
- **Database**: MySQL
- **Authentication**: JWT (JSON Web Tokens)
- **Documentation**: Swagger/OpenAPI
- **Validation**: go-playground/validator
- **Configuration**: Environment variables

## Quick Start

### Prerequisites

- Go 1.21 or higher
- MySQL 5.7 or higher
- Git

### Installation

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd ticketing-system
   ```

2. **Install dependencies**

   ```bash
   go mod tidy
   ```

3. **Set up environment variables**
   Create a `.env` file in the root directory:

   ```env
   DB_HOST=localhost
   DB_PORT=3306
   DB_USERNAME=root
   DB_PASSWORD=your_password
   DB_NAME=ticketing_system

   JWT_SECRET=your-super-secret-jwt-key-here-change-in-production
   JWT_EXPIRE_HOURS=24

   GIN_MODE=debug
   PORT=8080

   ADMIN_EMAIL=admin@ticketing.com
   ADMIN_PASSWORD=admin123
   ```

4. **Create MySQL database**

   ```sql
   CREATE DATABASE ticketing_system;
   ```

5. **Run the application**
   ```bash
   go run main.go
   ```

The server will start on `http://localhost:8080`

### Default Admin Account

- **Email**: admin@ticketing.com
- **Password**: admin123

## API Documentation

Once the server is running, you can access:

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/health

## API Endpoints

### Authentication

- `POST /api/v1/register` - Register new user
- `POST /api/v1/login` - User login

### User Management

- `GET /api/v1/profile` - Get user profile
- `PUT /api/v1/profile` - Update user profile
- `GET /api/v1/users` - Get all users (Admin)
- `DELETE /api/v1/users/{id}` - Delete user (Admin)

### Event Management

- `GET /api/v1/events` - Get all events (with filtering)
- `GET /api/v1/events/{id}` - Get event by ID
- `GET /api/v1/events/active` - Get active events
- `GET /api/v1/events/upcoming` - Get upcoming events
- `POST /api/v1/events` - Create event (Admin)
- `PUT /api/v1/events/{id}` - Update event (Admin)
- `DELETE /api/v1/events/{id}` - Delete event (Admin)

### Ticket Management

- `POST /api/v1/tickets` - Buy tickets
- `GET /api/v1/tickets` - Get all tickets (Admin)
- `GET /api/v1/tickets/my` - Get user's tickets
- `GET /api/v1/tickets/{id}` - Get ticket by ID
- `PATCH /api/v1/tickets/{id}` - Update ticket status (Admin)
- `PATCH /api/v1/tickets/{id}/cancel` - Cancel ticket

### Reports

- `GET /api/v1/reports/summary` - Get summary report (Admin)
- `GET /api/v1/reports/event/{id}` - Get event report (Admin)

## Request/Response Examples

### User Registration

```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```

### Buy Tickets

```bash
curl -X POST http://localhost:8080/api/v1/tickets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "event_id": "event-uuid-here",
    "quantity": 2
  }'
```

### Get Events with Filtering

```bash
curl "http://localhost:8080/api/v1/events?page=1&limit=10&category=concert&location=jakarta&min_price=100000"
```

## Business Rules

### Event Management

- Event names must be unique
- Events cannot be modified once they're not in "active" status
- Events with sold tickets cannot be deleted
- Event dates cannot be in the past

### Ticket Management

- Users can only purchase tickets for active events
- Ticket purchases are blocked 1 hour before event start
- Users can cancel tickets up to 2 hours before event start
- Ticket cancellation returns tickets to event availability
- Users can only view/cancel their own tickets (except admins)

### Validation Rules

- Email must be valid and unique
- Passwords must be at least 6 characters
- Event capacity must be positive
- Ticket prices must be non-negative
- Event dates must be in the future

## Database Schema

The system uses the following main entities:

- **Users**: Authentication and profile information
- **Events**: Event details, capacity, and pricing
- **Tickets**: Ticket purchases and status tracking

All entities include soft delete functionality and audit timestamps.

## Security Features

- **JWT Authentication**: Secure token-based authentication
- **Password Hashing**: bcrypt with cost factor 12
- **CORS Support**: Configurable cross-origin resource sharing
- **Input Validation**: Comprehensive request validation
- **SQL Injection Protection**: GORM ORM prevents SQL injection
- **Role-based Access**: Admin and user role separation

## Development

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
go build -o ticketing-system main.go
```

### Environment Modes

- **debug**: Development mode with detailed logging
- **release**: Production mode with optimized performance

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions, please open an issue in the repository or contact the development team.

---

**Built with ❤️ using Go, Gin, and GORM**
