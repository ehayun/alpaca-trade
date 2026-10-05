package consts

import "github.com/gogf/gf/v2/errors/gcode"

// CodeUser marks an error whose message is Hebrew text that is safe to show the user.
// Business codes must be > 1000 (GoFrame convention).
var CodeUser = gcode.New(1001, "user error", nil)

// MsgInternal is shown for any error that is not safe to expose.
const MsgInternal = "אירעה שגיאה, נסו שוב מאוחר יותר"
