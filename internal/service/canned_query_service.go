package service

import (
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

type CannedQueryService struct {
	store *QueryStore
}

func NewCannedQueryService(store *QueryStore) *CannedQueryService {
	return &CannedQueryService{store: store}
}

func (s *CannedQueryService) ListCannedQueries() (*model.CannedQueryList, error) {
	return &model.CannedQueryList{Items: s.store.List()}, nil
}

func (s *CannedQueryService) SaveCannedQuery(req model.SaveCannedQueryRequest) (*model.CannedQuery, error) {
	item, err := s.store.SaveQuery(req)
	if err != nil {
		return nil, err
	}
	if err := s.store.Save(); err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	return item, nil
}

func (s *CannedQueryService) DeleteCannedQuery(id string) error {
	if err := s.store.Delete(id); err != nil {
		return err
	}
	return s.store.Save()
}
