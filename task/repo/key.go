package repo

import (
	"bytes"
	"encoding/binary"
	"time"
)

type key struct {
	completedTime time.Time
	id            uint64
}

func (k key) Marshal() []byte {
	buf := bytes.Buffer{}
	buf.Write(timeToBytes(k.completedTime))
	buf.Write(intToBytes(k.id))
	return buf.Bytes()
}

func unmarshalKeyBytes(keyBytes []byte) key {
	completedTimeBytes, idBytes := keyBytes[:8], keyBytes[8:] // both completed time and id are marshalled into 8 byte sequences
	return key{
		completedTime: bytesToTime(completedTimeBytes),
		id:            bytesToInt(idBytes),
	}
}

func timeToBytes(t time.Time) []byte {
	var epochNanos uint64
	if !t.IsZero() {
		epochNanos = uint64(t.UnixNano())
	}
	return intToBytes(epochNanos)
}

func bytesToTime(bytes []byte) time.Time {
	epochNanos := bytesToInt(bytes)
	if epochNanos == 0 {
		return time.Time{}
	}
	return time.Unix(0, int64(epochNanos))
}

func intToBytes(i uint64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, i)
	return bytes
}

func bytesToInt(bytes []byte) uint64 {
	return binary.BigEndian.Uint64(bytes)
}
