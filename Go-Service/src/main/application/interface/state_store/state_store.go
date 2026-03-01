package state_store

import "context"

type StateStore interface {
	GenerateState(ctx context.Context) (string, error)
	ValidateState(ctx context.Context, state string) error
}
