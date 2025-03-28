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
- `400 Bad Request`: No exchange rate available within 6 months of th
