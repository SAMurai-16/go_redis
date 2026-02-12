package commands

import (
	"fmt"

	"samyak.go_redis/store"
)






func HandleTYPE(st *store.Store, parts []string) []byte {
	if len(parts) < 2 {
		return []byte("-ERR wrong number of arguments\r\n")
	}

	key := parts[1]
	t := st.TypeOf(key)

	return []byte(fmt.Sprintf("+%s\r\n", t))
}
