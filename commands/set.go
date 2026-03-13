package commands

import (
	"strconv"
	"strings"
	"time"

	"samyak.go_redis/store"
)

func HandleSET(st *store.Store, parts []string, propogate bool) []byte {
	if len(parts) < 3 {
		return []byte("-ERR wrong number of arguments for 'set' command\r\n")
	}

	key := parts[1]
	value := parts[2]

	var expireAt time.Time
	if len(parts) >= 5 && strings.ToUpper(parts[3]) == "PX" {
		ms, _ := strconv.Atoi(parts[4])
		expireAt = time.Now().Add(time.Duration(ms) * time.Millisecond)
	}

	st.Set(key, value, expireAt)
	return []byte("+OK\r\n")
}

func ApplyReplicaCommand(st *store.Store, parts []string) {

	cmd := strings.ToUpper(parts[0])

	switch cmd {

	case "SET":
		HandleSET(st, parts, false)

	default:
		// ignore unknown or read-only commands
	}
}
