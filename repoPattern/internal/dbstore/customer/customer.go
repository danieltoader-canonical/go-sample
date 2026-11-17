package customer

import "context"

type Customer struct {
	CustomerId   int    `json:"customerId"`
	IsAnonymous  bool   `json:"IsAnonymous"`
	CustomerName string `json:"customerName"`
}

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=customer.go -destination=repo_mock.go -package=customer

type Repository interface {
	GetCustomer(context.Context, int) Customer
	UpdateCustomerName(context.Context, int, string) (*Customer, error)
	UpdateCustomerAnonymousStatus(context.Context, int, bool) error
}
