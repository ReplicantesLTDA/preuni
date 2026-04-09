module github.com/preuni/svc/simulation

go 1.24

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/preuni/pkg v0.0.0
)

require (
	github.com/stretchr/testify v1.10.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
)

replace github.com/preuni/pkg => ../../pkg
