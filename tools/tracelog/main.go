package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	linePat = regexp.MustCompile(`\w+\.go:\d+: +(\d+)\((\d+):(\d+)\) +# +(\d+) +(.+?) +@0x([0-9a-z]+)(?: +(.+))?`)
	copyPat = regexp.MustCompile(`^\+\+\+ +(\d+)\((\d+):(\d+)\) +copy`)
	endPat  = regexp.MustCompile(`^=== +End`)
	hndlPat = regexp.MustCompile(`^=== +Handle`)
	kvPat   = regexp.MustCompile(`^(\w+)=(.+)`)
)

type machID [3]int

var zeroMachID machID

type record struct {
	mid   machID
	count int
	ip    uint64
	act   string
	rest  []byte
}

type sessions map[machID]session

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

func (ss sessions) append(rec record) {
	sess := ss[rec.mid]
	sess.mid = rec.mid
	sess.recs = append(sess.recs, rec)
	ss[rec.mid] = sess
}

func (ss sessions) extend(mid machID, s string) {
	sess := ss[mid]
	sess.extra = append(sess.extra, s)
	ss[mid] = sess
}

func (ss sessions) link(pid, cid machID) {
	sess := ss[cid]
	sess.pid = pid
	ss[cid] = sess
}

func (ss sessions) walk(mid machID, f func(session)) session {
	sess, ok := ss[mid]
	if !ok {
		return sess
	}

	var q []session
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

func parseSessions(r io.Reader) (sessions, error) {
	var tail machID
	sessions := make(sessions)
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Bytes()

		if match := linePat.FindSubmatch(line); match != nil {
			var rec record
			rec.mid[0], _ = strconv.Atoi(string(match[1]))
			rec.mid[1], _ = strconv.Atoi(string(match[2]))
			rec.mid[2], _ = strconv.Atoi(string(match[3]))
			rec.count, _ = strconv.Atoi(string(match[4]))
			rec.act = strings.TrimRight(string(match[5]), " \r\n")
			rec.ip, _ = strconv.ParseUint(string(match[6]), 16, 64)
			rec.rest = match[7]
			sessions.append(rec)

			if match := copyPat.FindStringSubmatch(rec.act); match != nil {
				var pid machID
				pid[0], _ = strconv.Atoi(match[1])
				pid[1], _ = strconv.Atoi(match[2])
				pid[2], _ = strconv.Atoi(match[3])
				sessions.link(pid, rec.mid)
			}

			if endPat.MatchString(rec.act) {
				sess := sessions[rec.mid]
				if match := kvPat.FindSubmatch(rec.rest); match != nil {
					switch string(match[1]) {
					case "err":
						sess.err = string(match[2])
					case "values":
						sess.values = string(match[2])
					default:
						log.Printf("UNKNON End key/val: %q = %q\n", match[1], match[2])
					}
				}
				sessions[rec.mid] = sess
				tail = rec.mid
			}

			if hndlPat.MatchString(rec.act) {
				tail = zeroMachID
			}

		} else if tail != zeroMachID {
			sessions.extend(tail, strings.TrimRight(string(line), " \r\n"))
		}

	}
	return sessions, sc.Err()
}

func main() {
	sessions, err := parseSessions(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	for _, sess := range sessions {
		if sess.err != "" {
			continue
		}

		fmt.Printf("%v %v\n", sess.mid, sess.values)
		sessions.walk(sess.mid, func(sess session) {
			for _, rec := range sess.recs {
				fmt.Printf("	%v\n", rec)
			}
		})
		for _, line := range sess.extra {
			fmt.Printf("    %s\n", line)
		}
		fmt.Println()

	}

}
