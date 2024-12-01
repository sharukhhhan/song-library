package service

import "errors"

var (
	ErrSongAlreadyExists = errors.New("song already exists")
	ErrSongNotFound      = errors.New("song not found")
)
