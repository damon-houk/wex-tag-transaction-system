# WEX TAG Transaction Processing System

This application provides a system for storing purchase transactions and retrieving them with currency conversion functionality.

## Requirements

- Go 1.18 or higher

## Running the Application

To run the application locally:

\`\`\`bash
make run
\`\`\`

## API Endpoints

### Store a Purchase Transaction

\`\`\`
POST /transactions
\`\`\`

### Retrieve a Transaction with Currency Conversion

\`\`\`
GET /transactions/:id?currency=EUR
\`\`\`

## Development

\`\`\`bash
# Run tests
make test

# Build the application
make build
\`\`\`

## License

[MIT License](LICENSE)
