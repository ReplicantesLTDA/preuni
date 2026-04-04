module github.com/preuni/svc/auth

go 1.24

require (
	github.com/preuni/pkg v0.0.0
	github.com/go-chi/chi/v5 v5.2.1
	github.com/jackc/pgx/v5 v5.7.2
	go.uber.org/zap v1.27.0
	github.com/stretchr/testify v1.10.0
)

replace github.com/preuni/pkg => ../../pkg
