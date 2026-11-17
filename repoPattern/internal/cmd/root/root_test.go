package root

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"repoPattern/internal/dbstore/customer"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRootCommand(t *testing.T) {
	failure := errors.New("operation failed")
	for _, tt := range []struct {
		name       string
		args       []string
		update     bool
		newName    string
		initErr    error
		updateErr  error
		missing    bool
		writeError bool
		wantErr    string
	}{
		{name: "read"},
		{name: "rename", args: []string{"--name", "Alice"}, update: true, newName: "Alice"},
		{name: "empty name", args: []string{"--name="}, update: true},
		{name: "initialization error", initErr: failure, wantErr: failure.Error()},
		{name: "update error", args: []string{"--name", "Alice"}, update: true, newName: "Alice", updateErr: failure, wantErr: failure.Error()},
		{name: "missing customer", missing: true, wantErr: "customer 1 could not be loaded"},
		{name: "output error", writeError: true, wantErr: failure.Error()},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := customer.NewMockRepository(gomock.NewController(t))
			record := customer.Customer{CustomerId: 1, CustomerName: tt.newName}
			if tt.initErr == nil {
				var updateCall *gomock.Call
				if tt.update {
					updateCall = repo.EXPECT().UpdateCustomerName(gomock.Any(), 1, tt.newName).Return(&record, tt.updateErr)
				}
				if tt.updateErr == nil {
					result := record
					if tt.missing {
						result = customer.Customer{}
					}
					call := repo.EXPECT().GetCustomer(gomock.Any(), 1).Return(result)
					if updateCall != nil {
						call.After(updateCall)
					}
				}
			}
			cleaned := false
			command := NewRootCommand(func(ctx context.Context) (customer.Repository, func(), error) {
				_, ok := ctx.Deadline()
				require.True(t, ok)
				return repo, func() { cleaned = true }, tt.initErr
			})
			var out bytes.Buffer
			command.SetOut(&out)
			if tt.writeError {
				command.SetOut(errorWriter{failure})
			}
			command.SetArgs(append([]string{}, tt.args...))
			err := command.Execute()
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.JSONEq(t, `{"customerId":1,"IsAnonymous":false,"customerName":"`+tt.newName+`"}`, out.String())
			}
			require.Equal(t, tt.initErr == nil, cleaned)
		})
	}
}

func TestRootCommandWithoutDependencies(t *testing.T) {
	for _, tt := range []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "help", args: []string{"--help"}},
		{name: "zero ID", args: []string{"--customer-id", "0"}, wantErr: "customer-id must be positive"},
		{name: "negative ID", args: []string{"--customer-id=-1"}, wantErr: "customer-id must be positive"},
		{name: "invalid ID", args: []string{"--customer-id", "abc"}, wantErr: "invalid argument"},
		{name: "positional argument", args: []string{"unexpected"}, wantErr: "unknown command"},
		{name: "unknown flag", args: []string{"--unknown"}, wantErr: "unknown flag"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			command := NewRootCommand(func(context.Context) (customer.Repository, func(), error) {
				t.Fatal("dependencies must not be initialized")
				return nil, nil, nil
			})
			command.SetOut(&bytes.Buffer{})
			command.SetArgs(tt.args)
			err := command.Execute()
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

type errorWriter struct{ err error }

func (w errorWriter) Write([]byte) (int, error) { return 0, w.err }
