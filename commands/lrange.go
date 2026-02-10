package commands

import (
	"fmt"
	"strconv"

	"samyak.go_redis/store"
)

func HandleLRANGE(st *store.Store, parts []string) []byte {
	if len(parts) < 4 {
		return []byte("-ERR wrong number of arguments\r\n")
	}

	key := parts[1]

	start, err1 := strconv.Atoi(parts[2])
	stop, err2 := strconv.Atoi(parts[3])

	if err1 != nil || err2 != nil {
		return []byte("-ERR invalid index\r\n")
	}
	values := st.LRange(key, start, stop)

	resp := fmt.Sprintf("*%d\r\n", len(values))
	for _, v := range values {
		resp += fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
	}

	return []byte(resp)
}
