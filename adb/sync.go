package adb

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func readUInt32(a *AdbConn) uint32 {
	b := make([]byte, 4)
	a.Read(b)
	return binary.LittleEndian.Uint32(b)
}

func parseDent(a *AdbConn) {
	readUInt32(a)           // MODE
	readUInt32(a)           // SIZE
	readUInt32(a)           // MODIFIED TIME
	length := readUInt32(a) // NAME LENGTH

	b := make([]byte, length)
	a.Read(b)

	fmt.Printf("%s\n", b)
}

func Ls(t Transporter, remote string) ([]byte, error) {
	conn, err := t.Dial()
	if err != nil {
		return []byte{}, err
	}
	defer conn.Close()

	t.Transport(conn)
	_, err = conn.WriteCmd("sync:")
	if err != nil {
		return nil, err
	}

	w := bufio.NewWriter(conn)
	w.WriteString("LIST")
	binary.Write(w, binary.LittleEndian, uint32(len(remote)))
	w.WriteString(remote)
	w.Flush()

	b := make([]byte, 4)
	for {
		conn.Read(b)
		id := string(b)
		if id == "DENT" {
			parseDent(conn)
		} else if id == "DONE" {
			break
		}
	}

	return b, nil
}

func PushDevices(devices []*Device, local *os.File, remote string) error {
	d := make([]Transporter, 0, len(devices))
	for _, t := range devices {
		d = append(d, Transporter(t))
	}

	return PushAll(d, local, remote)
}

func PushAll(devices []Transporter, local *os.File, remote string) error {
	info, err := local.Stat()
	if err != nil {
		return err
	}

	d := make([]io.Writer, 0, len(devices))
	for _, t := range devices {
		conn, err := GetPushWriter(Transporter(t), remote, uint32(info.Mode()))
		if err != nil {
			return err
		}
		d = append(d, io.Writer(conn))
	}

	reader := bufio.NewReader(local)
	sections := NewSectionedMultiWriter(d...)
	writer := bufio.NewWriter(sections)
	writer.ReadFrom(reader)
	writer.Flush()
	sections.Close()

	wr := bufio.NewWriter(io.MultiWriter(d...))
	wr.WriteString("DONE")
	binary.Write(wr, binary.LittleEndian, uint32(info.ModTime().Unix()))
	wr.Flush()
	return nil
}

func Push(t Transporter, local *os.File, remote string) error {
	return PushAll([]Transporter{t}, local, remote)
}

func GetPushWriter(t Transporter, remote string, filePerm uint32) (*AdbConn, error) {
	conn, err := t.Dial()
	if err != nil {
		return nil, err
	}

	t.Transport(conn)
	_, err = conn.WriteCmd("sync:")
	if err != nil {
		return nil, err
	}

	w := bufio.NewWriter(conn)
	w.WriteString("SEND")
	binary.Write(w, binary.LittleEndian, uint32(len(remote)+5))
	w.WriteString(remote)
	w.WriteString(",")
	binary.Write(w, binary.LittleEndian, filePerm)
	w.Flush()

	return conn, nil
}

type SectionedMultiWriter struct {
	writer    io.Writer
	buffer    []byte
	bufferIdx int
	section   int
}

func NewSectionedMultiWriter(writers ...io.Writer) *SectionedMultiWriter {
	return &SectionedMultiWriter{writer: io.MultiWriter(writers...), buffer: make([]byte, 65536)}
}

func (w *SectionedMultiWriter) Write(b []byte) (int, error) {
	i := copy(w.buffer[w.bufferIdx:], b)
	w.bufferIdx += i

	atmax := w.bufferIdx == 65536
	if i < len(b) || atmax {
		w.section++
		w.Flush()

		if !atmax {
			return w.Write(b[i:])
		}
	}
	return i, nil
}

func (w *SectionedMultiWriter) Flush() {
	wr := bufio.NewWriter(w.writer)
	wr.WriteString("DATA")
	binary.Write(wr, binary.LittleEndian, uint32(w.bufferIdx))
	wr.Write(w.buffer[:w.bufferIdx])
	wr.Flush()

	w.bufferIdx = 0
}

func (w *SectionedMultiWriter) Close() error {
	if len(w.buffer) != 0 {
		w.Flush()
	}
	return nil
}
