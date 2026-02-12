package commands

import (
	"fmt"

	"samyak.go_redis/store"
)

func HandleXADD(st *store.Store, parts []string) []byte {
	// XADD key id field value [field value ...]
	if len(parts) < 5 || len(parts)%2 != 1 {
		return []byte("-ERR wrong number of arguments\r\n")
	}


	key := parts[1]
	id := parts[2]

	fields := make(map[string]string)

	for i := 3; i < len(parts); i += 2 {
		field := parts[i]
		value := parts[i+1]
		fields[field] = value
	}

	returnID,err:= st.XAdd(key, id, fields)
	if err!= nil{
		return []byte(fmt.Sprintf("-%s\r\n", err.Error()))
	}

	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(returnID), returnID))
}
