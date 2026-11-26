package domain

import "errors"

var (
	ErrMissingHeader     = errors.New("missing header")
	ErrDuplicateHeader   = errors.New("duplicate header")
	ErrChunkBeforeHeader = errors.New("chunk before header")
	ErrDuplicateFinalize = errors.New("duplicate finalize")
	ErrInvalidChunkSeq   = errors.New("invalid chunk seq")
	ErrMissingChunks     = errors.New("missing chunks")
	ErrUnkownSnapId      = errors.New("unkown snapshot id")
	ErrSessionExpired    = errors.New("session expired")
)
