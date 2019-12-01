package action

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/npenkov/gcqlsh/internal/output"

	"github.com/npenkov/gcqlsh/internal/db"

	"github.com/gocql/gocql"
)

type tracer struct {
	cks *db.CQLKeyspaceSession
	tw  *traceWriter
	buf *bytes.Buffer
	cf  func()
}

type traceRow struct {
	timestamp string
	source    string
	elapsed   int
	activity  string
}

func NewTracer(cks *db.CQLKeyspaceSession) *tracer {
	if cks.TracingEnabled {
		b := &bytes.Buffer{}
		s, cf, err := cks.CloneSession()
		if err != nil {
			output.PrintError(fmt.Sprintf("Cannot create trace session: %v", err))
			return &tracer{cks: cks, tw: nil, buf: nil, cf: nil}
		}
		t := NewTraceWriter(s, b)
		return &tracer{cks: cks, tw: t, buf: b, cf: cf}
	}
	return &tracer{cks: cks, tw: nil, buf: nil, cf: nil}
}

func (t *tracer) Query(stmt string, values ...interface{}) *gocql.Query {
	if t.cks.TracingEnabled {
		return t.cks.Session.Query(stmt, values...).Trace(t.tw)
	}
	return t.cks.Session.Query(stmt, values...)
}

func (t *tracer) Close() {
	if t.cks.TracingEnabled {
		fmt.Print(t.buf.String())
		t.cf()
	}
}

type traceWriter struct {
	session *gocql.Session
	w       io.Writer
	mu      sync.Mutex
}

func NewTraceWriter(session *gocql.Session, w io.Writer) *traceWriter {
	return &traceWriter{session: session, w: w}
}

func (t *traceWriter) Trace(traceId []byte) {
	var (
		coordinator string
		duration    int
	)
	iter := t.session.Query(`SELECT coordinator, duration
			FROM system_traces.sessions
			WHERE session_id = ?`, traceId).Iter()

	iter.Scan(&coordinator, &duration)
	if err := iter.Close(); err != nil {
		t.mu.Lock()
		fmt.Fprintln(t.w, "Error:", err)
		t.mu.Unlock()
		return
	}

	var (
		timestamp time.Time
		activity  string
		source    string
		elapsed   int
	)

	t.mu.Lock()
	defer t.mu.Unlock()

	iter = t.session.Query(`SELECT event_id, activity, source, source_elapsed
			FROM system_traces.events
			WHERE session_id = ?`, traceId).Iter()

	colWidths := []int{23, 12, 12, 60}
	tRow := []*traceRow{}

	for iter.Scan(&timestamp, &activity, &source, &elapsed) {
		tr := &traceRow{
			timestamp: timestamp.Format("2006/01/02 15:04:05.999999"),
			source:    source,
			elapsed:   elapsed,
			activity:  activity,
		}
		tRow = append(tRow, tr)
		if colWidths[3] < len(tr.activity) {
			colWidths[3] = len(tr.activity)
		}
	}
	if len(tRow) <= 0 {
		return
	}

	var addSpaceColor = 0
	if output.Colorful() {
		addSpaceColor = 9
	}

	fmt.Fprintf(t.w, "Tracing session %016x (coordinator: %s, duration: %v):\n",
		traceId, coordinator, time.Duration(duration)*time.Microsecond)

	fmt.Fprintf(t.w, fmt.Sprintf("| %%%ds ", colWidths[0]+addSpaceColor), output.Magenta("timestamp"))
	fmt.Fprintf(t.w, fmt.Sprintf("| %%%ds ", colWidths[1]+addSpaceColor), output.Magenta("source"))
	fmt.Fprintf(t.w, fmt.Sprintf("| %%%ds ", colWidths[2]+addSpaceColor), output.Magenta("elapsed"))
	fmt.Fprintf(t.w, fmt.Sprintf("| %%%ds \n", colWidths[3]+addSpaceColor), output.Magenta("action"))

	fmt.Fprintf(t.w, fmt.Sprintf("+%%%ds", colWidths[0]+2), strings.Repeat("-", colWidths[0]+2))
	fmt.Fprintf(t.w, fmt.Sprintf("+%%%ds", colWidths[1]+2), strings.Repeat("-", colWidths[1]+2))
	fmt.Fprintf(t.w, fmt.Sprintf("+%%%ds", colWidths[2]+2), strings.Repeat("-", colWidths[2]+2))
	fmt.Fprintf(t.w, fmt.Sprintf("+%%%ds\n", colWidths[3]+2), strings.Repeat("-", colWidths[3]+2))

	for _, r := range tRow {
		// fmt.Fprintf(t.w, "%s: %s (source: %s, elapsed: %d)\n",
		// timestamp.Format("2006/01/02 15:04:05.999999"), activity, source, elapsed)
		fmt.Fprintf(t.w, fmt.Sprintf("| %%%ds ", colWidths[0]+addSpaceColor), output.Yellow(r.timestamp))
		fmt.Fprintf(t.w, fmt.Sprintf("| %%%ds ", colWidths[1]+addSpaceColor), output.Yellow(r.source))
		fmt.Fprintf(t.w, fmt.Sprintf("| %%%ds ", colWidths[2]+addSpaceColor), output.Yellow(r.elapsed))
		fmt.Fprintf(t.w, fmt.Sprintf("| %%%ds \n", colWidths[3]+addSpaceColor), output.Yellow(r.activity))
	}

	if err := iter.Close(); err != nil {
		fmt.Fprintln(t.w, "Error:", err)
	}
}
