module github.com/preuni/svc/monolith

go 1.24

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/jackc/pgx/v5 v5.7.2
	github.com/preuni/pkg v0.0.0
	github.com/preuni/svc/auth v0.0.0
	github.com/preuni/svc/user v0.0.0
)

require (
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)

replace github.com/preuni/pkg => ../../pkg

replace github.com/preuni/svc/auth => ../auth

replace github.com/preuni/svc/user => ../user
