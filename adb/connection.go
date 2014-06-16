package adb

import (
	"bufio"
    "errors"
	"fmt"
    "net"
    "strconv"
)

type AdbConn struct {
    conn net.Conn
    r *bufio.Reader
}

func (a *Adb) getConnection() (*AdbConn, error) {
    h := fmt.Sprintf("%s:%d", a.Host, a.Port)
    c, err := net.Dial("tcp", h)
    if err != nil {
        return nil, err
    }
    return &AdbConn{c, bufio.NewReader(c)}, nil
}

func (a *AdbConn) readSize(bcount int) (uint64, error) {
    size := make([]byte, bcount);
    a.r.Read(size);
    return strconv.ParseUint(string(size), 16, 0)
}

func (a *AdbConn) Write(cmd []byte) (int, error) {
    prefix := fmt.Sprintf("%04x", len(cmd))
    w := bufio.NewWriter(a.conn)
    w.WriteString(prefix)
    i, err := w.Write(cmd)

    if err != nil {
        return 0, errors.New(`Could not write to ADB server`)
    }

    w.Flush()

    status := make([]byte, 4);
    _, err = a.Read(status)
    if err != nil || string(status) != `OKAY` {
        return 0, errors.New(`Invalid connection`)
    }

    return i, nil
}

func (a *AdbConn) Read(p []byte) (int, error) {
    if a.r == nil {
    	a.r = bufio.NewReader(a.conn)
    } 

    return a.r.Read(p)
}

func (a *AdbConn) Close() error {
    if a.conn != nil {
        return a.conn.Close()
    }
    return nil
}
