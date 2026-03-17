package cartesia

// CursorPage is a page of results from a cursor-paginated API endpoint.
type CursorPage[T any] struct {
	Data    []T     `json:"data"`
	HasMore bool    `json:"has_more"`
	Next    *string `json:"next,omitempty"`
}

// ListParams contains common cursor-based pagination parameters.
type ListParams struct {
	Limit         *int    `url:"limit,omitempty"`
	StartingAfter *string `url:"starting_after,omitempty"`
}

// PageIterator provides automatic pagination over cursor-paginated endpoints.
type PageIterator[T any] struct {
	current *CursorPage[T]
	fetch   func(params ListParams) (*CursorPage[T], error)
	params  ListParams
	err     error
	started bool
}

// NewPageIterator creates an iterator that fetches pages using the provided function.
func NewPageIterator[T any](fetch func(ListParams) (*CursorPage[T], error), params ListParams) *PageIterator[T] {
	return &PageIterator[T]{
		fetch:  fetch,
		params: params,
	}
}

// Next advances to the next page. Returns false when no more pages exist or on error.
func (it *PageIterator[T]) Next() bool {
	if it.err != nil {
		return false
	}

	if !it.started {
		it.started = true
		page, err := it.fetch(it.params)
		if err != nil {
			it.err = err
			return false
		}
		it.current = page
		return len(page.Data) > 0
	}

	if !it.current.HasMore || it.current.Next == nil {
		return false
	}

	it.params.StartingAfter = it.current.Next
	page, err := it.fetch(it.params)
	if err != nil {
		it.err = err
		return false
	}
	it.current = page
	return len(page.Data) > 0
}

// Current returns the items from the current page.
func (it *PageIterator[T]) Current() []T {
	if it.current == nil {
		return nil
	}
	return it.current.Data
}

// Err returns any error encountered during iteration.
func (it *PageIterator[T]) Err() error {
	return it.err
}

// Int is a helper that returns a pointer to an int value.
func Int(v int) *int { return &v }

// String is a helper that returns a pointer to a string value.
func String(v string) *string { return &v }

// Bool is a helper that returns a pointer to a bool value.
func Bool(v bool) *bool { return &v }

// Float64 is a helper that returns a pointer to a float64 value.
func Float64(v float64) *float64 { return &v }
