package pidcat

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/wmbest2/android/adb"
	"regexp"
	"strings"
)

type colorFunc func(format string, a ...interface{}) string

var (
	PID_PARSER = regexp.MustCompile(`\S+\s+(\S+)(?:\s+\S+){5}\s+(?:\S\s)?(\S*)`)

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

func (p *PidCat) matches(pid []byte) bool {
	if len(p.AppFilters) == 0 {
		return true
	}
	_, present := p.pidFilters[string(pid)]
	return present
}

func (p *PidCat) matchesPackage(pack []byte) bool {
	if len(p.AppFilters) == 0 {
		return true
	}
	_, present := p.AppFilters[string(pack)]
	return present
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

func (p *PidCat) Sprint(in []byte) []byte {
	line := ParseLine(in)

	if line == nil {
		return nil
	} else if line.Type == ProcessStart {
		if p.matchesPackage(line.Package) {
			p.pidFilters[string(line.PID)] = true
			return line.Message
		}
		return nil
	} else if line.Type == ProcessStop && p.matches(line.PID) {
		delete(p.pidFilters, string(line.PID))
		return line.Message
	}

	inPid := p.matches(line.PID)
	if inPid && !p.PrettyPrint {
		return []byte(fmt.Sprintf("%s	(%s): %s", line.Tag, line.PID, line.Message))
	} else if !inPid {
		return nil
	}

	tag := string(line.Tag)
	if tag != p.lastTag {
		p.lastTag = tag
		colorize := getColor(tag)
		count := p.TagWidth - len(tag)
		if len(tag) > p.TagWidth {
			tag = tag[:p.TagWidth]
			count = 0
		}
		tag = fmt.Sprintf("%s%s", strings.Repeat(` `, count), tag)
		return []byte(fmt.Sprintln(colorize(tag), " ", TagTypes[string(line.Level)], " ", string(line.Message)))
	} else {
		tag = strings.Repeat(` `, p.TagWidth)
		return []byte(fmt.Sprintln(tag, " ", TagTypes[string(line.Level)], " ", string(line.Message)))
	}

	return nil
}
