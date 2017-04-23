package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	linePat = regexp.MustCompile(`\w+\.go:\d+: +(\d+)\((\d+):(\d+)\) +# +(\d+) +(.+) +@0x([0-9a-z]+)(?: +(.+))?`)

	actPat = regexp.MustCompile(`(` +
		`^\+\+\+ +(\d+)\((\d+):(\d+)\) +copy` +
		`)|(` +
		`^=== +End` +
		`)|(` +
		`^=== +Handle` +
		`)`)

	kvPat = regexp.MustCompile(`^(\w+)=(.+)`)
)

type machID [3]int

var zeroMachID machID

type record struct {
	mid   machID
	count int
	ip    uint64
	act   string
	rest  string
}

type recordKind int

const (
	unknownLine = recordKind(iota)
	genericLine
	copyLine
	endLine
	hndlLine
)

func (ss sessions) parseRecord(line []byte) (rec record, kind recordKind) {
	match := linePat.FindSubmatch(line)
	if match == nil {
		kind = unknownLine
		return
	}

	kind = genericLine
	rec.mid[0], _ = strconv.Atoi(string(match[1]))
	rec.mid[1], _ = strconv.Atoi(string(match[2]))
	rec.mid[2], _ = strconv.Atoi(string(match[3]))
	rec.count, _ = strconv.Atoi(string(match[4]))
	rec.act = strings.TrimRight(string(match[5]), " \r\n")
	rec.ip, _ = strconv.ParseUint(string(match[6]), 16, 64)
	rec.rest = string(match[7])

	sess := ss.session(rec.mid)
	sess.mid = rec.mid
	sess.recs = append(sess.recs, rec)

	switch amatch := actPat.FindStringSubmatch(rec.act); {
	case amatch == nil:
	case amatch[1] != "": // copy
		kind = copyLine
		sess.pid[0], _ = strconv.Atoi(amatch[2])
		sess.pid[1], _ = strconv.Atoi(amatch[3])
		sess.pid[2], _ = strconv.Atoi(amatch[4])

	case amatch[5] != "": // end
		kind = endLine
		if match := kvPat.FindStringSubmatch(rec.rest); match != nil {
			sess := ss.session(rec.mid)
			switch string(match[1]) {
			case "err":
				sess.err = string(match[2])
			case "values":
				sess.values = string(match[2])
			default:
				log.Printf("UNKNOWN End key/val: %q = %q\n", match[1], match[2])
			}
		}

	case amatch[6] != "": // handle
		kind = hndlLine
	}

	return
}

type sessions map[machID]*session

type session struct {
	mid, pid machID
	recs     []record
	err      string
	values   string
	extra    []string
}

func (mid machID) String() string {
	return fmt.Sprintf("%d(%d:%d)", mid[0], mid[1], mid[2])
}

func (rec record) String() string {
	return fmt.Sprintf(
		"% 10v #% 4d @%#04x % -30s %q",
		rec.mid,
		rec.count,
		rec.ip,
		rec.act,
		rec.rest,
	)
}

func (ss sessions) session(mid machID) *session {
	sess := ss[mid]
	if sess == nil {
		sess = &session{mid: mid}
		ss[mid] = sess
	}
	return sess
}

func (ss sessions) extend(mid machID, s string) {
	sess := ss.session(mid)
	sess.extra = append(sess.extra, s)
}

func (ss sessions) walk(mid machID, f func(*session)) *session {
	sess := ss[mid]
	if sess == nil {
		return nil
	}

	var q []*session
	for sess.pid != zeroMachID {
		q = append(q, sess)
		sess = ss[sess.pid]
	}

	for {
		f(sess)
		if i := len(q) - 1; i >= 0 {
			sess, q = q[i], q[:i]
		} else {
			break
		}
	}
	return sess
}

func (ss sessions) parentIDs(sess *session) []machID {
	n := 0
	for id := sess.pid; id != zeroMachID; id = ss[id].pid {
		n++
	}
	ids := make([]machID, n)
	for i, id := len(ids), sess.pid; id != zeroMachID; id = ss[id].pid {
		i--
		ids[i] = id
	}
	return ids
}

func (ss sessions) fullID(sess *session) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%d(", sess.mid[0])
	for _, id := range ss.parentIDs(sess) {
		fmt.Fprintf(&buf, "%d:", id[2])
	}
	fmt.Fprintf(&buf, "%d)", sess.mid[2])
	return buf.String()
}

func (ss sessions) sessionLog(sess *session, logf func(string, ...interface{})) {
	ss.walk(sess.mid, func(sess *session) {
		for _, rec := range sess.recs {
			logf("%v", rec)
		}
	})
	for _, line := range sess.extra {
		logf("%s", line)
	}
}

func parseSessions(r io.Reader) (sessions, error) {
	var tail machID
	sessions := make(sessions)
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Bytes()
		switch rec, kind := sessions.parseRecord(line); kind {
		case unknownLine:
			if tail != zeroMachID {
				extra := strings.TrimRight(string(line), " \r\n")
				sessions.extend(tail, extra)
			}
		case endLine:
			tail = rec.mid
		case hndlLine:
			tail = zeroMachID
		}
	}
	return sessions, sc.Err()
}

func main() {
	var terse bool

	flag.BoolVar(&terse, "terse", false, "don't print full session logs")
	flag.Parse()

	sessions, err := parseSessions(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	mids := make([]machID, 0, len(sessions))
	for mid, sess := range sessions {
		mids = append(mids, mid)
	}
	sort.Slice(mids, func(i, j int) bool {
		return mids[i][0] < mids[j][0] ||
			mids[i][1] < mids[j][1] ||
			mids[i][2] < mids[j][2]
	})
	for _, mid := range mids {
		sess := sessions[mid]
		if sess.err != "" {
			fmt.Printf("%s\terr=%v\n", sessions.fullID(sess), sess.err)
		} else {
			fmt.Printf("%s\tvalues=%v\n", sessions.fullID(sess), sess.values)
		}
		if !terse {
			sessions.sessionLog(sess, func(format string, args ...interface{}) {
				fmt.Printf("	"+format+"\n", args...)
			})
			fmt.Println()
		}
	}
}
