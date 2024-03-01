Service consists of:
- client
  - it collects go runtime metrics and sends it to the server
  - it does it concurrently with a certain interval that may be set in config
- server
  - accepts metrics of different types and processes it
  - stores to database or updates if metric has already been added
