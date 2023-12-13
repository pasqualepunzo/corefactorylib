package lib

import (
	"os"
	"strconv"
)

var jsonLog, _ = strconv.ParseBool(os.Getenv("jsonLog"))
