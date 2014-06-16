package adb

import (
	"bufio"
    "errors"
	"fmt"
    "net"
    "strconv"
)

func (a *Adb) getConnection() (net.Conn, error) {
    h := fmt.Sprintf("%s:%d", a.Host, a.Port)
    return net.Dial("tcp", h)
}

func (a *Adb) readSize(reader *bufio.Reader, bcount int) (uint64, error) {
    size := make([]byte, bcount);
    reader.Read(size);
    return strconv.ParseUint(string(size), 16, 0)
}

func (a *Adb) Send(conn net.Conn, cmd string) (*bufio.Reader, error) {
    fmt.Fprintf(conn, "%04x%s", len(cmd), cmd)
    
    reader := bufio.NewReader(conn)
    status := make([]byte, 4);
    _, err := reader.Read(status)
    if err != nil || string(status) != `OKAY` {
        return nil, errors.New(`invalid connection`)
    }

    return reader, nil
}

