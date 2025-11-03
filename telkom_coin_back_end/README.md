# Telkom Coin Backend

A blockchain-based payment application backend built with Go, featuring a custom cryptocurrency called "Telkom Coin". This application provides PayPal-like functionality with top-up, withdrawal, transfer, and transaction history features.

## Features

- **User Authentication**: Registration, login with JWT tokens
- **Wallet Management**: Automatic wallet generation with encrypted private keys
- **PIN Security**: 6-digit PIN for transaction authorization
- **Top-up**: Convert IDR to Telkom Coin via multiple payment methods
- **Transfer**: Send Telkom Coin to other users by wallet address or username
- **Withdrawal**: Convert Telkom Coin back to IDR
- **Transaction History**: Complete transaction tracking with blockchain hashing
- **KYC Verification**: Know Your Customer verification system
- **Balance Management**: Real-time balance tracking with locked balance support

## Tech Stack

- **Language**: Go 1.24.5
- **Framework**: Gin (HTTP web framework)
- **Database**: MySQL with GORM ORM
- **Authentication**: JWT tokens
- **Encryption**: AES-GCM for private keys, bcrypt for passwords
- **Blockchain**: SHA256 transaction hashing

## Project Structure

```
telkom_coin_back_end/
├── app/                    # Application setup and routing
├── cmd/                    # Main application entry point
├── config/                 # Configuration and database setup
├── internal/
│   ├── blockchain/         # Blockchain service for transaction hashing
│   ├── dto/               # Data Transfer Objects
│   │   ├── request/       # Request DTOs
│   │   └── response/      # Response DTOs
│   ├── handler/           # HTTP handlers
│   ├── middleware/        # HTTP middleware (JWT auth)
│   ├── models/           # Database models
│   ├── repository/       # Data access layer
│   └── services/         # Business logic layer
├── pkg/
│   ├── crypto/           # Cryptographic utilities
│   ├── helpers/          # Helper functions
│   └── validator/        # Custom validators
└── scripts/              # Database migration and seeding scripts
```

## Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd telkom_coin_back_end
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Setup environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials and other configurations
   ```

4. **Setup database**
   ```bash
   # Create MySQL database
   mysql -u root -p -e "CREATE DATABASE telkom_coin_db;"
   
   # Run migrations
   go run scripts/migrate.go
   
   # Seed initial data (optional)
   go run scripts/seed.go
   ```

5. **Run the application**
   ```bash
   go run cmd/main.go
   ```

## Environment Variables

Copy `.env.example` to `.env` and configure:

```env
# Server
PORT=8080

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASS=password
DB_NAME=telkom_coin_db

# JWT
JWT_SECRET=your-super-secret-jwt-key
```

## API Endpoints

### Authentication
- `POST /register` - User registration
- `POST /login` - User login

### User Management
- `GET /api/profile` - Get user profile
- `GET /api/profile/detail` - Get detailed profile with balance
- `PUT /api/profile` - Update profile
- `POST /api/change-password` - Change password
- `POST /api/change-pin` - Change PIN

### Balance
- `GET /api/balance` - Get current balance

### Top-up
- `POST /api/topup` - Request top-up
- `GET /api/topup/history` - Get top-up history
- `GET /api/topup/payment-methods` - Get available payment methods

### Transfer
- `POST /api/transfer` - Transfer by wallet address
- `POST /api/transfer/by-username` - Transfer by username
- `GET /api/transfer/history` - Get transfer history
- `POST /api/transfer/validate-recipient` - Validate recipient

### Withdrawal
- `POST /api/withdraw` - Request withdrawal
- `GET /api/withdraw/history` - Get withdrawal history
- `GET /api/withdraw/methods` - Get available withdrawal methods

### Transactions
- `GET /api/transactions` - Get transaction history
- `GET /api/transactions/:hash` - Get transaction by hash

## Database Schema

### Users Table
- ID, Username, Email, Phone
- Password Hash, PIN Hash
- Wallet Address, Encrypted Private Key
- KYC Status, Account Status

### Balances Table
- User ID, Wallet Address
- Available Balance, Locked Balance

### Transactions Table
- Transaction Hash, From/To Addresses
- Amount, Type, Status
- Block Number, Gas Used
- Metadata (JSON)

## Security Features

- **JWT Authentication**: Secure API access
- **Password Hashing**: bcrypt for password security
- **Private Key Encryption**: AES-GCM encryption
- **PIN Protection**: SHA256 hashed PINs for transactions
- **Balance Locking**: Prevents double-spending during pending transactions

## Development

### Running Tests
```bash
go test ./...
```

### Code Structure
- **Repository Pattern**: Data access abstraction
- **Service Layer**: Business logic separation
- **Handler Layer**: HTTP request handling
- **Middleware**: Cross-cutting concerns (auth, logging)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.
