package entities

type Transaction struct {
}

type GetTransactionListReq struct {
	Paginator
	UID uint `json:"-"`
}
