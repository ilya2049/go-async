package worksgroup

import (
	"context"
	"errors"

	"golang.org/x/sync/errgroup"
)

var (
	ErrInterrupt = errors.New("works interrupted")
)

func Run[WR any](
	ctx context.Context,
	parallelWorks int,
	works []func(context.Context) (WR, error),
	addWorkResult func(workResult WR),
) error {
	if len(works) == 0 {
		return nil
	}

	if parallelWorks <= 0 {
		return nil
	}

	errGroup, ctx := errgroup.WithContext(ctx)
	const serviceProcessesCount = 2
	errGroup.SetLimit(parallelWorks + serviceProcessesCount)

	workResultsChan := make(chan WR)

	errGroup.Go(func() error {
		for i := 0; i < len(works); i++ {
			select {
			case workResult := <-workResultsChan:
				addWorkResult(workResult)
			case <-ctx.Done():
				return nil
			}
		}

		return nil
	})

	errGroup.Go(func() error {
		for _, work := range works {
			work := work

			select {
			case <-ctx.Done():
				return nil
			default:
				errGroup.Go(func() error {
					workResult, err := work(ctx)
					if err != nil {
						return err
					}

					select {
					case <-ctx.Done():
						return nil
					case workResultsChan <- workResult:
					}

					return nil
				})
			}

		}

		return nil
	})

	if err := errGroup.Wait(); err != nil && !errors.Is(err, ErrInterrupt) {
		return err
	}

	return nil
}
