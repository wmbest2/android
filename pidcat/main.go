package pidcat

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/wmbest2/android/adb"
	"regexp"
	"strings"
)

type colorFunc func(format string, a ...interface{}) string

var (
	PID_PARSER     = regexp.MustCompile(`\S+\s+(\S+)(?:\s+\S+){5}\s+(?:\S\s)?(\S*)`)
	PID_START      = regexp.MustCompile(`^Start proc ([a-zA-Z0-9._:]+) for ([a-z]+ [^:]+): pid=(\d+) uid=(\d+) gids=(.*)$`)
	PID_KILL       = regexp.MustCompile(`^Killing (\d+):([a-zA-Z0-9._:]+)/[^:]+: (.*)$`)
	PID_LEAVE      = regexp.MustCompile(`^No longer want ([a-zA-Z0-9._:]+) \(pid (\d+)\): .*$`)
	PID_DEATH      = regexp.MustCompile(`^Process ([a-zA-Z0-9._:]+) \(pid (\d+)\) has died.?$`)
	LOG_LINE       = regexp.MustCompile(`^([A-Z])/(.+?)\( *(\d+)\): (.*?)$`)
	BUG_LINE       = regexp.MustCompile(`.*nativeGetEnabledTags.*`)
	BACKTRACE_LINE = regexp.MustCompile(`^#(.*?)pc\s(.*?)$`)

	TagTypes     map[string]string
	KnownTags    map[string]colorFunc
	LastUsed     []colorFunc
	HeaderOffset = 1 + 2 + 1 //  space, level, space
)

func init() {
	TagTypes = make(map[string]string)
	TagTypes[`V`] = color.New(color.FgWhite, color.BgBlack).SprintFunc()(` V `)
	TagTypes[`D`] = color.New(color.FgBlack, color.BgBlue).SprintFunc()(` D `)
	TagTypes[`I`] = color.New(color.FgBlack, color.BgGreen).SprintFunc()(` I `)
	TagTypes[`W`] = color.New(color.FgBlack, color.BgYellow).SprintFunc()(` W `)
	TagTypes[`E`] = color.New(color.FgBlack, color.BgRed).SprintFunc()(` E `)
	TagTypes[`F`] = color.New(color.FgBlack, color.BgRed).SprintFunc()(` F `)

	KnownTags = make(map[string]colorFunc)
	KnownTags[`dalvikvm`] = color.WhiteString

	LastUsed = []colorFunc{
		color.RedString, color.GreenString, color.YellowString, color.BlueString,
		color.MagentaString, color.CyanString}
}

type PidCat struct {
	AppFilters  map[string]bool
	pidFilters  map[string]bool
	TagFilters  []string
	PrettyPrint bool
	TagWidth    int
	lastTag     string
}

func NewPidCat(pretty bool, tw int) *PidCat {
	pid := PidCat{PrettyPrint: pretty, TagWidth: tw}
	pid.AppFilters = make(map[string]bool)
	pid.pidFilters = make(map[string]bool)
	return &pid
}

func Clear(t adb.Transporter) {
	adb.ShellSync(t, "logcat", "-c")
}

func (p *PidCat) SetAppFilters(filters ...string) {
	if len(filters) == 1 && filters[0] == "" {
		p.AppFilters = nil
		return
	}
	p.AppFilters = make(map[string]bool)
	for _, f := range filters {
		p.AppFilters[f] = true
	}
}

func (p *PidCat) UpdateAppFilters(t adb.Transporter) error {
	if p.AppFilters == nil {
		return nil
	}

	ps := adb.Shell(t, "ps")

	for line := range ps {
		groups := PID_PARSER.FindSubmatch(line)
		app := string(groups[2])
		_, present := p.AppFilters[app]
		if present {
			p.pidFilters[string(groups[1])] = true
		}
	}

	return nil
}

func (p *PidCat) matches(pid string) bool {
	if len(p.AppFilters) == 0 {
		return true
	}
	_, present := p.pidFilters[pid]
	return present
}

func (p *PidCat) matchesPackage(pack string) bool {
	if len(p.AppFilters) == 0 {
		return true
	}
	_, present := p.AppFilters[pack]
	return present
}

func (p *PidCat) parseDeath(tag string, msg string) (string, string) {
	if tag == "ActivityManager" {
		var matcher *regexp.Regexp
		swap := false
		if PID_KILL.MatchString(msg) {
			matcher = PID_KILL
			swap = true
		} else if PID_LEAVE.MatchString(msg) {
			matcher = PID_LEAVE
		} else if PID_DEATH.MatchString(msg) {
			matcher = PID_DEATH
		}
		if matcher != nil {
			match := matcher.FindStringSubmatch(msg)
			pid := match[2]
			pack := match[1]
			if swap {
				pid = pack
				pack = match[2]
			}
			if p.matchesPackage(pack) && p.matches(pid) {
				return pid, pack
			}
		}
	}

	return "", ""
}

func getColor(tag string) colorFunc {
	v, ok := KnownTags[tag]
	if ok {
		return v
	}
	color := LastUsed[0]
	LastUsed = append(LastUsed[1:], LastUsed[0])
	return color
}

func (p *PidCat) Sprint(line string) string {
	if BUG_LINE.MatchString(line) || !LOG_LINE.MatchString(line) {
		return ""
	}
	var buffer bytes.Buffer

	logline := LOG_LINE.FindStringSubmatch(line)
	if PID_START.MatchString(logline[4]) {
		start := PID_START.FindStringSubmatch(logline[4])
		if p.matchesPackage(start[1]) {
			p.pidFilters[start[3]] = true
			if p.PrettyPrint {
				buffer.WriteString("\n")
				fmt.Sprintf("\nProcess: %s (PID: %s) started\n", start[3], start[1])
			}
		}
	}

	pid, pack := p.parseDeath(logline[2], logline[4])
	if pid != "" {
		delete(p.pidFilters, pid)
		if p.PrettyPrint {
			return fmt.Sprintf("\nProcess: %s (PID: %s) ended\n", pid, pack)
		}
	}

	inPid := p.matches(logline[3])
	if inPid && !p.PrettyPrint {
		return fmt.Sprintln(line)
	} else if !inPid {
		return ""
	}

	tag := strings.TrimSpace(logline[2])
	if tag != p.lastTag {
		p.lastTag = tag
		colorize := getColor(tag)
		count := p.TagWidth - len(tag)
		if len(tag) > p.TagWidth {
			tag = tag[:p.TagWidth]
			count = 0
		}
		tag = fmt.Sprintf("%s%s", strings.Repeat(` `, count), tag)
		return fmt.Sprintln(colorize(tag), " ", TagTypes[logline[1]], " ", logline[4])
	} else {
		tag = strings.Repeat(` `, p.TagWidth)
		return fmt.Sprintln(tag, " ", TagTypes[logline[1]], " ", logline[4])
	}

	return ""
}
