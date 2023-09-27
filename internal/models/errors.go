package models

import "errors"

// ErrPersonNotFound is error occured if no record for given person found in storage.
var ErrPersonNotFound = errors.New("person not found")

// ErrNoChangesMade is error occured if no changes were made after change request.
var ErrNoChangesMade = errors.New("no changes made")

// ErrCouldNotEnrich is error occured if request to API to enrich person could not find info to enrich with.
var ErrCouldNotEnrich = errors.New("could not enrich, try another name")
