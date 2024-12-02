package domain

import "context"

type DataBaseReader interface {
	TakeSnapshot(ctx context.Context) ([]*DataBaseSnapshot, error)
}
