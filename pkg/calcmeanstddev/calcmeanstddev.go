package calcmeanstddev

import (
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/influxdata/kapacitor/udf/agent"

	"pkg/matchtime"
)

type calcMeanStddev struct {
	timeFilter string
	timeZone   string
	field      string
	entries    []float64

	timeMask string
	now      time.Time

	agent *agent.Agent
}

func newCalcMeanStddev(agent *agent.Agent) *calcMeanStddev {
	return &calcMeanStddev{
		agent: agent,
	}
}

// Return the InfoResponse. Describing the properties of thfp UDF agent.
func (*calcMeanStddev) Info() (*agent.InfoResponse, error) {
	info := &agent.InfoResponse{
		Wants:    agent.EdgeType_BATCH,
		Provides: agent.EdgeType_BATCH,

		Options: map[string]*agent.OptionInfo{
			"timeFilter": {ValueTypes: []agent.ValueType{agent.ValueType_STRING, agent.ValueType_STRING}},
			"field":      {ValueTypes: []agent.ValueType{agent.ValueType_STRING}},
		},
	}

	return info, nil
}

// Initialze the handler based of the provided options.
func (sm *calcMeanStddev) Init(r *agent.InitRequest) (*agent.InitResponse, error) {
	init := &agent.InitResponse{
		Success: true,
		Error:   "",
	}

	for _, opt := range r.Options {
		switch opt.Name {
		case "timeFilter":
			sm.timeFilter = strings.TrimSpace(opt.Values[0].Value.(*agent.OptionValue_StringValue).StringValue)
			sm.timeZone = strings.TrimSpace(opt.Values[1].Value.(*agent.OptionValue_StringValue).StringValue)
		case "field":
			sm.field = strings.TrimSpace(opt.Values[0].Value.(*agent.OptionValue_StringValue).StringValue)
		}
	}

	if len(sm.field) == 0 {
		init.Success = false
		init.Error = "must supply 'field'"
	}

	return init, nil
}

// Create a snapshot of the running state of the process.
func (*calcMeanStddev) Snapshot() (*agent.SnapshotResponse, error) {
	return &agent.SnapshotResponse{}, nil
}

// Restore a previous snapshot.
func (*calcMeanStddev) Restore(req *agent.RestoreRequest) (*agent.RestoreResponse, error) {
	return &agent.RestoreResponse{
		Success: true,
	}, nil
}

// Start working with the next batch
func (sm *calcMeanStddev) BeginBatch(begin *agent.BeginBatch) error {
	// Housekeeping for each time serise
	sm.entries = nil
	now := time.Now()
	sm.now = converTimeToTimezone(&now, sm.timeZone)
	sm.timeMask = sm.generateTimeMask()

	return nil
}

func (sm *calcMeanStddev) Point(p *agent.Point) error {
	// Convert nanosecond epoch format time to timezone time
	dt := time.Unix(0, p.GetTime())
	dt = converTimeToTimezone(&dt, sm.timeZone)

	// Only process data points that match time mask
	if len(sm.timeMask) == 0 || matchtime.MatchTimeWithMask(sm.timeMask, &dt) {
		val, ok := p.FieldsDouble[sm.field]
		if !ok {
			i := p.FieldsInt[sm.field]
			val = float64(i)
		}
		sm.entries = append(sm.entries, val)
	}

	return nil
}

func converTimeToTimezone(t *time.Time, timeZone string) time.Time {
	if len(timeZone) > 0 {
		loc, err := time.LoadLocation(timeZone)
		if err == nil {
			return t.In(loc)
		}
	}

	return *t
}

func (sm *calcMeanStddev) generateTimeMask() string {
	// Replace the now field in the time filter with the related
	// value of the time Now.
	// For example, given the time now is 2019-08-27T20:30:00Z
	// and the time filter is "W>=1 & W<=5 & h==now & m==now & Y==now",
	// the result after processed is "W>=1 & W <=5 & h==20 & m==30 & Y==2019"

	var sb strings.Builder
	curr := ""

	for i := 0; i < len(sm.timeFilter); i++ {
		ch := sm.timeFilter[i]
		if ch == 'h' || ch == 'm' || ch == 'W' || ch == 'D' || ch == 'M' || ch == 'Y' || ch == 's' {
			curr = string(ch)
			sb.WriteByte(ch)
		} else if ch == 'n' {
			v := matchtime.GetTimeField(curr, &sm.now)
			sb.WriteString(strconv.Itoa(v))
			i += 2
		} else if !unicode.IsSpace(rune(ch)) {
			sb.WriteByte(ch)
		}
	}

	return sb.String()
}

func (sm *calcMeanStddev) EndBatch(end *agent.EndBatch) error {
	// Send the new data point back to Kapacitor
	p := &agent.Point{
		FieldsDouble: make(map[string]float64),
	}

	if len(sm.entries) > 0 {
		m, sd := calculateMeanStddev(sm.entries)

		p.FieldsDouble["mean"] = m
		p.FieldsDouble["stddev"] = sd
		p.Time = end.GetTmax()
		p.Name = end.GetName()
		p.Group = end.GetGroup()
		p.Tags = end.GetTags()

		sm.agent.Responses <- &agent.Response{
			Message: &agent.Response_Point{
				Point: p,
			},
		}
	}

	return nil
}

// Stop the handler gracefully.
func (sm *calcMeanStddev) Stop() {
	close(sm.agent.Responses)
}

func calculateMeanStddev(data []float64) (float64, float64) {
	var t float64

	for _, v := range data {
		t += v
	}

	m := t / float64(len(data))

	var va float64
	for _, v := range data {
		va += math.Pow(v-m, 2)
	}

	return m, math.Sqrt(va / float64(len(data)))
}

// Start is the entry point to start UDF
func Start() {
	a := agent.New(os.Stdin, os.Stdout)
	h := newCalcMeanStddev(a)
	a.Handler = h

	log.Println("Starting agent 'calcmeanstddev'")
	a.Start()
	err := a.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
