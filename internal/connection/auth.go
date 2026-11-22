package connection

import "time"

// AuthReadTimeout limits how long a new connection may wait for its first auth message.
const AuthReadTimeout = 5 * time.Second
