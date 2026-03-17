package api

type Pagination struct {
	TotalResults    int    `json:"totalResults"`
	ReturnedResults int    `json:"returnedResults"`
	PerPage         int    `json:"perPage"`
	TotalPages      int    `json:"totalPages"`
	NextCursor      string `json:"nextCursor"`
	NextPage        int    `json:"nextPage"`
}

type PaginationParams struct {
	PerPage string
	Cursor  string
}

func Paginate[T any](fetch func(cursor string) ([]T, *Pagination, error)) ([]T, error) {
	var all []T
	cursor := ""
	for {
		items, page, err := fetch(cursor)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
		if page.NextCursor == "" {
			return all, nil
		}
		cursor = page.NextCursor
	}
}
