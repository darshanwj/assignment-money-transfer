# assignment-money-transfer
a RESTful API for money transfers between accounts.

## Installation

Clone the repo and run
```
$ go run main.go
```

## Usage

Server runs on `http://localhost:8090`

### Create some accounts

```
$ curl -X POST http://localhost:8090/accounts -d '{"id":1,"customer_id":1,"balance":{"amount":500,"currency":"USD"}}'

$ curl -X POST http://localhost:8090/accounts -d '{"id":2,"customer_id":2,"balance":{"amount":300,"currency":"USD"}}'
```

### Do a money transfer
```
$ curl -X POST http://localhost:8090/transfer -d '{"sender_account_id":1,"receiver_account_id":2,"amount":{"amount":100,"currency":"USD"}}'
```

### List all accounts
```
$ curl http://localhost:8090/accounts
```

## Tests

### Run functional tests
```
$ go test
```

## Notes
- This API uses [https://github.com/Rhymond/go-money](https://github.com/Rhymond/go-money) for safe operation and storage of money values. Value of 100 is the smallest unit and represents $1  
- Money transfers involving currency conversion is not supported
- Database operations that have to be atomic are identified with comments