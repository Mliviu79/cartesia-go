package cartesia_test

import (
	"encoding/json"
	"fmt"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
	"github.com/google/go-querystring/query"
)

func TestCursorPage_JSONUnmarshal(t *testing.T) {
	raw := `{"data":["a","b","c"],"has_more":true,"next":"cursor123"}`
	var page cartesia.CursorPage[string]
	if err := json.Unmarshal([]byte(raw), &page); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(page.Data) != 3 {
		t.Errorf("expected 3 items, got %d", len(page.Data))
	}
	if page.Data[0] != "a" || page.Data[1] != "b" || page.Data[2] != "c" {
		t.Errorf("unexpected data: %v", page.Data)
	}
	if !page.HasMore {
		t.Error("expected HasMore=true")
	}
	if page.Next == nil || *page.Next != "cursor123" {
		t.Error("expected Next='cursor123'")
	}
}

func TestCursorPage_JSONUnmarshal_NoMore(t *testing.T) {
	raw := `{"data":[1,2],"has_more":false}`
	var page cartesia.CursorPage[int]
	if err := json.Unmarshal([]byte(raw), &page); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(page.Data) != 2 {
		t.Errorf("expected 2 items, got %d", len(page.Data))
	}
	if page.HasMore {
		t.Error("expected HasMore=false")
	}
	if page.Next != nil {
		t.Error("expected Next=nil")
	}
}

func TestListParams_URLEncoding(t *testing.T) {
	tests := []struct {
		name     string
		params   cartesia.ListParams
		expected map[string]string
		absent   []string
	}{
		{
			name:   "empty params",
			params: cartesia.ListParams{},
			absent: []string{"limit", "starting_after"},
		},
		{
			name:     "limit only",
			params:   cartesia.ListParams{Limit: cartesia.Int(10)},
			expected: map[string]string{"limit": "10"},
			absent:   []string{"starting_after"},
		},
		{
			name:     "both params",
			params:   cartesia.ListParams{Limit: cartesia.Int(25), StartingAfter: cartesia.String("abc")},
			expected: map[string]string{"limit": "25", "starting_after": "abc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, err := query.Values(tt.params)
			if err != nil {
				t.Fatalf("query.Values error: %v", err)
			}
			for k, v := range tt.expected {
				if got := vals.Get(k); got != v {
					t.Errorf("expected %s=%q, got %q", k, v, got)
				}
			}
			for _, k := range tt.absent {
				if vals.Has(k) {
					t.Errorf("expected %s to be absent", k)
				}
			}
		})
	}
}

func TestPageIterator_MultiplePages(t *testing.T) {
	pages := []cartesia.CursorPage[string]{
		{Data: []string{"a", "b"}, HasMore: true, Next: cartesia.String("cursor1")},
		{Data: []string{"c", "d"}, HasMore: true, Next: cartesia.String("cursor2")},
		{Data: []string{"e"}, HasMore: false, Next: nil},
	}
	callIdx := 0

	fetch := func(params cartesia.ListParams) (*cartesia.CursorPage[string], error) {
		if callIdx >= len(pages) {
			t.Fatal("too many fetch calls")
		}
		page := pages[callIdx]
		callIdx++
		return &page, nil
	}

	iter := cartesia.NewPageIterator(fetch, cartesia.ListParams{Limit: cartesia.Int(2)})

	var allItems []string
	for iter.Next() {
		allItems = append(allItems, iter.Current()...)
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(allItems) != 5 {
		t.Errorf("expected 5 items, got %d: %v", len(allItems), allItems)
	}
	expected := []string{"a", "b", "c", "d", "e"}
	for i, v := range expected {
		if allItems[i] != v {
			t.Errorf("item[%d]: expected %q, got %q", i, v, allItems[i])
		}
	}
	if callIdx != 3 {
		t.Errorf("expected 3 fetch calls, got %d", callIdx)
	}
}

func TestPageIterator_StopsWhenHasMoreFalse(t *testing.T) {
	callCount := 0
	fetch := func(params cartesia.ListParams) (*cartesia.CursorPage[int], error) {
		callCount++
		return &cartesia.CursorPage[int]{
			Data:    []int{1, 2, 3},
			HasMore: false,
			Next:    nil,
		}, nil
	}

	iter := cartesia.NewPageIterator(fetch, cartesia.ListParams{})
	pageCount := 0
	for iter.Next() {
		pageCount++
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pageCount != 1 {
		t.Errorf("expected 1 page, got %d", pageCount)
	}
	if callCount != 1 {
		t.Errorf("expected 1 fetch call, got %d", callCount)
	}
}

func TestPageIterator_PropagatesFetchError(t *testing.T) {
	expectedErr := fmt.Errorf("network failure")
	fetch := func(params cartesia.ListParams) (*cartesia.CursorPage[string], error) {
		return nil, expectedErr
	}

	iter := cartesia.NewPageIterator(fetch, cartesia.ListParams{})
	if iter.Next() {
		t.Error("expected Next()=false on error")
	}
	if iter.Err() != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, iter.Err())
	}
}

func TestPageIterator_ErrorOnSecondPage(t *testing.T) {
	expectedErr := fmt.Errorf("server error")
	callIdx := 0
	fetch := func(params cartesia.ListParams) (*cartesia.CursorPage[string], error) {
		callIdx++
		if callIdx == 1 {
			return &cartesia.CursorPage[string]{
				Data:    []string{"x"},
				HasMore: true,
				Next:    cartesia.String("cur"),
			}, nil
		}
		return nil, expectedErr
	}

	iter := cartesia.NewPageIterator(fetch, cartesia.ListParams{})
	count := 0
	for iter.Next() {
		count++
	}
	if count != 1 {
		t.Errorf("expected 1 successful page, got %d", count)
	}
	if iter.Err() != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, iter.Err())
	}
}

func TestPageIterator_EmptyFirstPage(t *testing.T) {
	fetch := func(params cartesia.ListParams) (*cartesia.CursorPage[string], error) {
		return &cartesia.CursorPage[string]{
			Data:    []string{},
			HasMore: false,
		}, nil
	}

	iter := cartesia.NewPageIterator(fetch, cartesia.ListParams{})
	if iter.Next() {
		t.Error("expected Next()=false for empty page")
	}
	if iter.Err() != nil {
		t.Errorf("unexpected error: %v", iter.Err())
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("Int", func(t *testing.T) {
		p := cartesia.Int(42)
		if p == nil || *p != 42 {
			t.Errorf("expected *42, got %v", p)
		}
	})

	t.Run("String", func(t *testing.T) {
		p := cartesia.String("hello")
		if p == nil || *p != "hello" {
			t.Errorf("expected *'hello', got %v", p)
		}
	})

	t.Run("Bool", func(t *testing.T) {
		p := cartesia.Bool(true)
		if p == nil || !*p {
			t.Errorf("expected *true, got %v", p)
		}
		p2 := cartesia.Bool(false)
		if p2 == nil || *p2 {
			t.Errorf("expected *false, got %v", p2)
		}
	})

	t.Run("Float64", func(t *testing.T) {
		p := cartesia.Float64(3.14)
		if p == nil || *p != 3.14 {
			t.Errorf("expected *3.14, got %v", p)
		}
	})
}
