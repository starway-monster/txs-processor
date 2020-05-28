package processor

import "errors"

var CommitError = errors.New("could not process block")

var ConnectionError = errors.New("could not connect")

var BlockHeightError = errors.New("received block at invalid height")
