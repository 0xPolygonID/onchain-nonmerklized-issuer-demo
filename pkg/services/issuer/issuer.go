package issuer

import (
	"context"
)

type IssuerService struct {
	issuers []string
}

func NewIssuerService(
	issuers []string,
) *IssuerService {
	return &IssuerService{
		issuers: issuers,
	}
}

func (is *IssuerService) GetIssuersList(_ context.Context) []string {
	return is.issuers
}
