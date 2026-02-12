package commands

import (
	"fmt"
	"strconv"

	"samyak.go_redis/store"
)

func HandleLPOP(st *store.Store, parts []string) []byte {
	if len(parts) < 2 {
		return []byte("-ERR wrong number of arguments\r\n")
	}

	key := parts[1]

	// Case 1: no count → old behavior
	if len(parts) == 2 {
		values := st.LPop(key, 1)
		if len(values) == 0 {
			return []byte("$-1\r\n")
		}

		v := values[0]
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
	}

	// Case 2: with count
	count, err := strconv.Atoi(parts[2])
	if err != nil {
		return []byte("-ERR invalid count\r\n")
	}

	values := st.LPop(key, count)

	// Encode RESP array
	resp := fmt.Sprintf("*%d\r\n", len(values))
	for _, v := range values {
		resp += fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
	}

	return []byte(resp)
}
