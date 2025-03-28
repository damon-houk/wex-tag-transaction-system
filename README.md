# WEX TAG Transaction Processing System

This application provides a system for storing purchase transactions and retrieving them with currency conversion functionality using the Treasury Reporting Rates of Exchange API.

## Requirements

- Go 1.18 or higher

## Running the Application

To run the application locally:

```bash
make run
```

The server will start on port 8080 by default.

## API Documentation

### 1. Store a Purchase Transaction

Store a new purchase transaction with description, date, and amount.

**Endpoint:** `POST /transactions`

**Request Body:**
```json
{
  "description": "Office supplies",
  "date": "2023-04-15",
  "amount": 125.45
}
```

**Request Constraints:**
- `description`: Must not exceed 50 characters
- `date`: Must be a valid date in YYYY-MM-DD format and not in the future
- `amount`: Must be a positive number (will be rounded to the nearest cent)

**Success Response (201 Created):**
```json
{
  "id": "7f6c7d78-9b5e-4b6a-8d7c-5d8e6f7a8b9c"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid input data (with descriptive error message)
- `500 Internal Server Error`: Server-side error

### 2. Retrieve a Transaction

Retrieve a transaction by its ID.

**Endpoint:** `GET /transactions/{id}`

**Success Response (200 OK):**
```json
{
  "id": "7f6c7d78-9b5e-4b6a-8d7c-5d8e6f7a8b9c",
  "description": "Office supplies",
  "date": "2023-04-15",
  "amount": 125.45
}
```

**Error Responses:**
- `404 Not Found`: Transaction not found
- `500 Internal Server Error`: Server-side error

### 3. Retrieve a Transaction with Currency Conversion

Retrieve a transaction converted to a specified currency.

**Endpoint:** `GET /transactions/{id}/convert?currency={currency_code}`

**Query Parameters:**
- `currency`: The three-letter currency code to convert to (e.g., EUR, GBP, CAD)

**Success Response (200 OK):**
```json
{
  "id": "7f6c7d78-9b5e-4b6a-8d7c-5d8e6f7a8b9c",
  "description": "Office supplies",
  "date": "2023-04-15",
  "original_amount": 125.45,
  "currency": "EUR",
  "exchange_rate": 0.93,
  "converted_amount": 116.67,
  "rate_date": "2023-04-05"
}
```

**Error Responses:**
- `400 Bad Request`: Missing or invalid currency parameter
- `404 Not Found`: Transaction not found
- `400 Bad Request`: No exchange rate available within 6 months of the transaction date
- `503 Service Unavailable`: Treasury API unavailable
- `500 Internal Server Error`: Server-side error

## Currency Conversion Rules

When converting between currencies, the following rules apply:

1. The system uses the Treasury Reporting Rates of Exchange API to get conversion rates.
2. The conversion rate used will be the most recent rate available on or before the transaction date.
3. The rate must be from within 6 months prior to the transaction date.
4. If no suitable rate is available, an error is returned.
5. The converted amount is rounded to two decimal places (nearest cent).

## API Examples

### Create a New Transaction

```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Office supplies",
    "date": "2023-04-15",
    "amount": 125.45
  }'
```

### Retrieve a Transaction

```bash
# Replace TRANSACTION_ID with the ID returned from the create endpoint
curl http://localhost:8080/transactions/TRANSACTION_ID
```

### Convert a Transaction to EUR

```bash
# Replace TRANSACTION_ID with the ID of the transaction to convert
curl http://localhost:8080/transactions/TRANSACTION_ID/convert?currency=EUR
```

### Common Currency Codes

- EUR: Euro
- GBP: British Pound
- CAD: Canadian Dollar
- JPY: Japanese Yen
- AUD: Australian Dollar
- CHF: Swiss Franc
- CNY: Chinese Yuan
## Development

```bash
# Run tests
make test

# Build the application
make build

# Run linter
make lint

# Clean build artifacts
make clean
```

## Architecture

The application follows a hexagonal architecture (ports and adapters) pattern with the following layers:

- **Domain Layer**: Core business entities and repository interfaces
- **Application Layer**: Business logic and services
- **Infrastructure Layer**: Database implementation, API clients, and HTTP handlers

This architecture ensures separation of concerns and facilitates testing.

## License

[MIT License](LICENSE)
