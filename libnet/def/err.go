package def

import (
	"errors"
)

var  (
	SessionCannotFoundErr = errors.New("session cannot found,but still called")
)