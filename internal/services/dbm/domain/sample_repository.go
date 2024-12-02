package domain

import "context"

type SampleRepository interface {
	StoreSamples(ctx context.Context, samples []*QuerySample) error
}
