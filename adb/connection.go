package adb

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Transport int

const (
	Any Transport = iota
	Emulator
	Usb
)

type Transporter interface {
	Dial() (*AdbConn, error)
	Transport(conn *AdbConn) error
}

type Dialer struct {
	Host string
	Port int
}

type AdbConn struct {
	conn net.Conn
	r    *bufio.Reader
}

func (a *Dialer) Dial() (*AdbConn, error) {
	h := fmt.Sprintf("%s:%d", a.Host, a.Port)
	c, err := net.Dial("tcp", h)
	if err != nil {
		if c != nil {
			fmt.Printf("Closing connection to %s\n", c)
			c.Close()
		}
		return nil, err
	}
	return &AdbConn{c, bufio.NewReader(c)}, nil
}

func (a *AdbConn) TransportAny() error {
	cmd := fmt.Sprintf("host:transport-any")
	_, err := a.WriteCmd(cmd)
	return err
}

func (a *AdbConn) TransportEmulator() error {
	cmd := fmt.Sprintf("host:transport-local")
	_, err := a.WriteCmd(cmd)
	return err
}

func (a *AdbConn) TransportUsb() error {
	cmd := fmt.Sprintf("host:transport-usb")
	_, err := a.WriteCmd(cmd)
	return err
}

func (a *AdbConn) TransportSerial(ser string) error {
	cmd := fmt.Sprintf("host:transport:%s", ser)
	_, err := a.WriteCmd(cmd)
	return err
}

func (a *AdbConn) Shell(args ...string) error {
	cmd := fmt.Sprintf("shell:%s", strings.Join(args, " "))
	_, err := a.WriteCmd(cmd)
	return err
}

func (a *AdbConn) Log(args ...string) error {
	cmd := fmt.Sprintf("log:%s", strings.Join(args, " "))
	_, err := a.WriteCmd(cmd)
	return err
}

func (a *AdbConn) readSize(bcount int) (uint64, error) {
	size := make([]byte, bcount)
	a.r.Read(size)
	return strconv.ParseUint(string(size), 16, 0)
}

func (a *AdbConn) WriteCmd(cmd string) (int, error) {
	prefix := fmt.Sprintf("%04x", len(cmd))
	w := bufio.NewWriter(a)
	w.WriteString(prefix)
	i, err := w.WriteString(cmd)

	if err != nil {
		return 0, errors.New(`Could not write to ADB server`)
	}

	w.Flush()

	return i, a.VerifyOk()
}

func (a *AdbConn) readUint32() uint32 {
	size := make([]byte, 4)
	a.r.Read(size)
	return binary.LittleEndian.Uint32(size)
}

func (a *AdbConn) ReadCode() (string, error) {
	status := make([]byte, 4)
	_, err := a.Read(status)
	if err != nil {
		return "UNKN", err
	}
	return string(status), nil
}

func (a *AdbConn) VerifyOk() error {
	code, err := a.ReadCode()
	if err != nil || code != `OKAY` {
		return errors.New(`Invalid connection CODE: ` + code)
	}
	return nil
}

func (a *AdbConn) Write(b []byte) (int, error) {
	if a == nil || a.conn == nil {
		return 0, errors.New(`Could not write to ADB server`)
	}
	return a.conn.Write(b)
}

func (a *AdbConn) Read(p []byte) (int, error) {
	if a == nil {
		return 0, errors.New(`Could not read from ADB server`)
	} else if a.r == nil {
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
