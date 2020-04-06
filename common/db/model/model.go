package model

const (
	GOOD = 0
	DNS_BLOCKED = 1
	TCP_BLOCKED = 1 << 1
	TCP_RESET = 1 << 2
	WRONG_PAGE = 1 << 3
	BLANK_PAGE = 1 << 4
)

type URLStatus struct {
	URL string
	Status int
}

