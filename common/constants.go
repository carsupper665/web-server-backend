// ./common/constants.go
package common

// import (
// 	"os"
// 	"strconv"
// 	"sync"
// 	"time"

// 	"github.com/google/uuid"
// )

const (
	Version = "0.1.0"
	Bulid   = "bata-0.0.5"
)

const (
	RequestIdKey = "F-User-Request-Id"
)

const (
	RoleGuestUser  = 0
	RoleCommonUser = 1
	RoleAdminUser  = 4
	RoleRootUser   = 6
)

const (
	UserStatusEnabled  = 1 // don't use 0, 0 is the default value!
	UserStatusDisabled = 2 // also don't use 0
)
