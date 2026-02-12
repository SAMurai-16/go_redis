package store

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseStreamID(id string) (int64, int64, error) {
	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid id format")
	}

	ms, err1 := strconv.ParseInt(parts[0], 10, 64)
	seq, err2 := strconv.ParseInt(parts[1], 10, 64)

	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("invalid id format")
	}

	return ms, seq, nil
}

func (s *Store) XAdd(key, id string, fields map[string]string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.data[key]

	// Extract last entry info if exists
	var lastMS int64
	var lastSeq int64

	if exists && entry.Type == StreamType && len(entry.Stream) > 0 {
		lastID := entry.Stream[len(entry.Stream)-1].ID
		lastMS, lastSeq, _ = parseStreamID(lastID)
	}

	var ms, seq int64
	var err error


	//Case 1 : Auto-generated ID
	if id == "*"{
		ms = time.Now().UnixMilli()

		if exists && entry.Type == StreamType && len(entry.Stream) > 0 {
			if ms == lastMS {
				seq = lastSeq + 1
			} else {
				seq = 0
			}
		} else {
			seq = 0
		}

		id = fmt.Sprintf("%d-%d",ms, seq)

	// Case 2: Auto-generate sequence
	} else if strings.HasSuffix(id, "-*") {
		timePart := strings.TrimSuffix(id, "-*")

		ms, err = strconv.ParseInt(timePart, 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid id")
		}


		if !exists || entry.Type != StreamType || len(entry.Stream) == 0 {
			if ms == 0 {
				seq = 1
			} else {
				seq = 0
			}
		} else {
			if ms == lastMS {
				seq = lastSeq + 1
			} else {
				if ms == 0 {
					seq = 1
				} else {
					seq = 0
				}
			}
		}

		id = fmt.Sprintf("%d-%d", ms, seq)

	} else {
		// Case 3: Explicit ID (previous stage)
		ms, seq, err = parseStreamID(id)
		if err != nil {
			return "", fmt.Errorf("invalid id")
		}

		if ms == 0 && seq == 0 {
			return "", fmt.Errorf("ERR The ID specified in XADD must be greater than 0-0")
		}
	}


	// Ordering validation
	if exists && entry.Type == StreamType && len(entry.Stream) > 0 {
		if ms < lastMS || (ms == lastMS && seq <= lastSeq) {
			return "", fmt.Errorf("ERR The ID specified in XADD is equal or smaller than the target stream top item")
		}
	}

	streamEntry := StreamEntry{
		ID:     id,
		Fields: fields,
	}

	if !exists {
		s.data[key] = Entry{
			Type:   StreamType,
			Stream: []StreamEntry{streamEntry},
		}
	} else {
		entry.Stream = append(entry.Stream, streamEntry)
		s.data[key] = entry
	}

	return id, nil
}
