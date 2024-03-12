
### Service consists of:
- client
  - it collects go runtime metrics and sends it to the server
  - it does it concurrently with a certain interval that may be set in config
- server
  - accepts metrics of different types and processes it
  - stores to database or updates if metric has already been added

### How to start server
```bash
go run ./cmd/server
```
Configuration details can be found in `cmd/server/flags.go`

### How to run unit tests
```bash
go test -v ./...
```

### How to run staticlinter
```bash
go run ./cmd/staticlint ./...
```


