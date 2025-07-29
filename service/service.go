package service

import "fm/store"

type service struct {
	s3Client store.S3Client
	store store.Store
}

func New(store store.Store, s3 store.S3Client)*service{
	return &service{
		store:store,
		s3Client: s3,
	}
}