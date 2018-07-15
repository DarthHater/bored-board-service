package database

import "errors"

var ErrEditPost = errors.New("Posts can only be edited for 10 minutes")