package adb

import (
	"encoding/binary"
	"fmt"
)

func Ls(t Transporter) ([]byte, error) {
	conn, err := t.Dial()
	if err != nil {
		return []byte{}, err
	}
	defer conn.Close()

	t.Transport(conn)
	conn.WriteCmd("sync:")

	c := "/sdcard/DCIM/Camera/"
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(len(c)))

	conn.Write([]byte("LIST"))
	conn.Write(b)
	conn.Write([]byte(c))

	conn.Read(b)
	id := string(b)

	conn.Read(b)
	length := binary.LittleEndian.Uint32(b)

	fmt.Printf("Found %s with length %d\n", id, length)

	b = make([]byte, length)
	conn.Read(b)

	return b, nil
}
