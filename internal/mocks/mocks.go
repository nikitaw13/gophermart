package mocks

// All mocks for the project are generated from here:
//   go generate ./internal/mocks
//go:generate mockgen -source=../service/auth.go -destination=auth.go -package=mocks
//go:generate mockgen -source=../service/storage.go -destination=database.go -package=mocks
//go:generate mockgen -source=../handler/service.go -destination=handler.go -package=mocks
//go:generate mockgen -source=../accrual/storage.go -destination=accrual.go -package=mocks
