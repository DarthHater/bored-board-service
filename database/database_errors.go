package database

import "errors"

// ErrEditPost occurs when a user tries to edit a post after the designated time
var ErrEditPost = errors.New("Posts can only be edited for 10 minutes")
// ErrNoThread occurs when a thread doesn't exist
var ErrNoThread = errors.New("Couldn't find that thread")
// ErrWrongPassword when a user enters a password that doesn't match
var ErrWrongPassword = errors.New("Wrong password")
