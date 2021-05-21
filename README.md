# Fetch Rewards Examples

## Author - Maxwell Farver

----

## Running the application

### Docker

1) Download and setup Docker on your machine
2) Clone the repository `git clone https://github.com/max-farver/fetch-rewards-example.git`
3) Run `docker compose up` from the root of the cloned repository

### Executable files

- Navigate to [Releases](https://github.com/max-farver/fetch-rewards-example/releases) and download the file that corresponds to your machine.
    - **NOTE:** I recommend the Docker approach to launching the application if Docker is already installed as the executable could be blocked by a security system (this can be overridden though).
    
## Endpoints

### Health check
`http:/0.0.0.0:8000/health`
- Responds with "Healthy" if the application is running correctly.

### Transactions
`http:/0.0.0.0:8000/transactions`
- Accepts POST requests to create new Transactions.

*Request*
```json
{
  "payer": "DANNON",
  "points": 1000,
  "timestamp": "2020-11-02T14:00:00Z"
}
```

*Response*
```json
{
  "payer": "DANNON",
  "points": 1000,
  "timestamp": "2020-11-02T14:00:00Z"
}
```

### Spend
`http:/0.0.0.0:8000/spend`
- Accepts POST requests to spend points.

*Request*
```json
{ 
  "points": 1000
}
```

*Response*
```json
[
    {
        "payer": "Dannon",
        "points": -780
    },
    {
        "payer": "Miller",
        "points": -220
    }
]
```

### Balance
`http:/0.0.0.0:8000/balance`
- Accepts GET requests to retrieve the current balance of all Payers.

*Response*
```json
{
  "Dannon": 220,
  "Miller": 1000
}
```

## Tech

- **Language:** Go
- **Database:** Sqlite
- **Distribution:**
  - Docker
  - Goreleaser