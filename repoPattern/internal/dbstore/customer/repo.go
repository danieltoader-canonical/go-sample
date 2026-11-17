package customer

import (
	"context"
	"errors"
	"log/slog"

	"repoPattern/pkg/postgres"
)

type Repo struct {
	client  postgres.Client
	options *postgres.Options
}

func NewRepo(c postgres.Client) *Repo {
	return &Repo{
		client:  c,
		options: c.GetOptions(),
	}
}

func (r *Repo) GetCustomer(ctx context.Context, customerId int) Customer {
	ctx, cancel := context.WithTimeout(ctx, r.options.ReadTimeoutMedium)
	defer cancel()

	var customer Customer
	err := r.client.QueryRow(ctx,
		`SELECT customer_id, is_anonymous, customer_name FROM customer WHERE customer_id = $1`,
		customerId,
	).Scan(&customer.CustomerId, &customer.IsAnonymous, &customer.CustomerName)
	if err != nil {
		if !errors.Is(err, postgres.ErrNoRows) {
			slog.ErrorContext(ctx, "unable to get customer from PostgreSQL", "customerId", customerId, "error", err)
		}
		return Customer{}
	}
	return customer
}

func (r *Repo) UpdateCustomerName(ctx context.Context, customerId int, name string) (*Customer, error) {
	ctx, cancel := context.WithTimeout(ctx, r.options.WriteTimeoutMedium)
	defer cancel()

	var customer Customer
	err := r.client.QueryRow(ctx,
		`INSERT INTO customer (customer_id, customer_name) VALUES ($1, $2)
		ON CONFLICT (customer_id) DO UPDATE SET customer_name = EXCLUDED.customer_name
		RETURNING customer_id, is_anonymous, customer_name`,
		customerId, name,
	).Scan(&customer.CustomerId, &customer.IsAnonymous, &customer.CustomerName)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *Repo) UpdateCustomerAnonymousStatus(ctx context.Context, customerId int, isAnonymous bool) error {
	ctx, cancel := context.WithTimeout(ctx, r.options.WriteTimeoutMedium)
	defer cancel()

	var id int
	err := r.client.QueryRow(ctx,
		`INSERT INTO customer (customer_id, is_anonymous) VALUES ($1, $2)
		ON CONFLICT (customer_id) DO UPDATE SET is_anonymous = EXCLUDED.is_anonymous
		RETURNING customer_id`,
		customerId, isAnonymous,
	).Scan(&id)
	return err
}
