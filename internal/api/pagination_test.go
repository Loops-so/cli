package api

import (
	"errors"
	"testing"
)

func TestPaginate(t *testing.T) {
	t.Run("single page", func(t *testing.T) {
		calls := 0
		fetch := func(cursor string) ([]string, *Pagination, error) {
			calls++
			return []string{"a", "b"}, &Pagination{NextCursor: ""}, nil
		}

		items, err := Paginate(fetch)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 2 {
			t.Errorf("expected 2 items, got %d", len(items))
		}
		if calls != 1 {
			t.Errorf("expected 1 fetch call, got %d", calls)
		}
	})

	t.Run("multiple pages", func(t *testing.T) {
		pages := []struct {
			items  []string
			cursor string
		}{
			{[]string{"a", "b"}, "cursor1"},
			{[]string{"c", "d"}, "cursor2"},
			{[]string{"e"}, ""},
		}
		call := 0
		fetch := func(cursor string) ([]string, *Pagination, error) {
			p := pages[call]
			call++
			return p.items, &Pagination{NextCursor: p.cursor}, nil
		}

		items, err := Paginate(fetch)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 5 {
			t.Errorf("expected 5 items, got %d", len(items))
		}
		if call != 3 {
			t.Errorf("expected 3 fetch calls, got %d", call)
		}
	})

	t.Run("error on first fetch", func(t *testing.T) {
		fetchErr := errors.New("api error")
		fetch := func(cursor string) ([]string, *Pagination, error) {
			return nil, nil, fetchErr
		}

		_, err := Paginate(fetch)
		if !errors.Is(err, fetchErr) {
			t.Errorf("expected fetch error, got %v", err)
		}
	})

	t.Run("error mid-pagination", func(t *testing.T) {
		fetchErr := errors.New("api error")
		call := 0
		fetch := func(cursor string) ([]string, *Pagination, error) {
			call++
			if call == 2 {
				return nil, nil, fetchErr
			}
			return []string{"a"}, &Pagination{NextCursor: "cursor1"}, nil
		}

		_, err := Paginate(fetch)
		if !errors.Is(err, fetchErr) {
			t.Errorf("expected fetch error, got %v", err)
		}
	})
}
