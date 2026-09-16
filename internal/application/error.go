package application

import "errors"

var (
	MissingFile    = errors.New("missing file to upload")
	InvalidFile    = errors.New("invalid file")
	FailedReadFile = errors.New("fail read to upload")
)
