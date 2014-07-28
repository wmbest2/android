package pidcat

import (
	"bytes"
	"fmt"
	"regexp"
)

type LineType int

const (
	ProcessStart LineType = iota
	ProcessStop
	Log
	Backtrace
)

var (
	PID_START      = regexp.MustCompile(`^Start proc ([a-zA-Z0-9._:]+) for ([a-z]+ [^:]+): pid=(\d+) uid=(\d+) gids=(.*)$`)
	PID_KILL       = regexp.MustCompile(`^Killing (\d+):([a-zA-Z0-9._:]+)/[^:]+: (.*)$`)
	PID_LEAVE      = regexp.MustCompile(`^No longer want ([a-zA-Z0-9._:]+) \(pid (\d+)\): .*$`)
	PID_DEATH      = regexp.MustCompile(`^Process ([a-zA-Z0-9._:]+) \(pid (\d+)\) has died.?$`)
	LOG_LINE       = regexp.MustCompile(`^([A-Z])/(.+?)\( *(\d+)\): (.*?)$`)
	BUG_LINE       = regexp.MustCompile(`.*nativeGetEnabledTags.*`)
	BACKTRACE_LINE = regexp.MustCompile(`^#(.*?)pc\s(.*?)$`)
)

type Line struct {
	Type    LineType
	PID     []byte
	Package []byte
	Level   []byte
	Tag     []byte
	Message []byte
}

func parseDeath(tag, msg []byte) ([]byte, []byte) {
	if bytes.Compare(tag, []byte("ActivityManager")) == 0 {
		var matcher *regexp.Regexp
		swap := false
		if PID_KILL.Match(msg) {
			matcher = PID_KILL
			swap = true
		} else if PID_LEAVE.Match(msg) {
			matcher = PID_LEAVE
		} else if PID_DEATH.Match(msg) {
			matcher = PID_DEATH
		}
		if matcher != nil {
			match := matcher.FindSubmatch(msg)
			pid := match[2]
			pack := match[1]
			if swap {
				pid = pack
				pack = match[2]
			}
			return pid, pack
		}
	}

	return nil, nil
}

func ParseLine(line []byte) *Line {
	if BUG_LINE.Match(line) || !LOG_LINE.Match(line) {
		return nil
	}
	var out Line

	logline := LOG_LINE.FindSubmatch(line)
	if PID_START.Match(logline[4]) {
		start := PID_START.FindSubmatch(logline[4])
		out.PID = start[3]
		out.Package = start[1]
		out.Type = ProcessStart
		out.Message = []byte(fmt.Sprintf("\nProcess: %s (PID: %s) started\n", start[3], start[1]))
		return &out
	}

	pid, pack := parseDeath(logline[2], logline[4])
	if pid != nil {
		out.PID = pid
		out.Package = pack
		out.Type = ProcessStop
		out.Message = []byte(fmt.Sprintf("\nProcess: %s (PID: %s) ended\n", pack, pid))
		return &out
	}

	tag := bytes.TrimSpace(logline[2])
	out.Type = Log
	out.PID = logline[3]
	out.Message = logline[4]
	out.Tag = tag
	out.Level = logline[1]
	return &out
}
