package ports

import "context"

// UnitOfWorkFactory begins a request-scoped unit of work.
type UnitOfWorkFactory func(ctx context.Context) (UnitOfWork, error)
