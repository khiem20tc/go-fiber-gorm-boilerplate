package validator

type PaginationParams struct {
	Page     int `json:"page" validate:"page"`
	PageSize int `json:"pageSize" validate:"page_size"`
}

type SortParams struct {
	SortField string `json:"sortField"`
	SortOrder string `json:"sortOrder" validate:"sortOrder"`
}

type CommonParams struct {
	PaginationParams
	SortParams
}

type WithEmail struct {
	Email string `json:"email" validate:"required,email,min=10,max=50"`
}
