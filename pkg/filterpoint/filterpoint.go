package filterpoint

import (
	"log"
	"os"
	"pkg/utils"
	"regexp"
	"strings"
	"time"

	"github.com/influxdata/kapacitor/udf/agent"

	"pkg/matchtime"
)

type filterPoint struct {
	timeZone string

	timeMask string

	agent *agent.Agent
}

func newFilterPoint(agent *agent.Agent) *filterPoint {
	return &filterPoint{
		agent: agent,
	}
}

// Return the InfoResponse. Describing the properties of thfp UDF agent.
func (*filterPoint) Info() (*agent.InfoResponse, error) {
	info := &agent.InfoResponse{
		Wants:    agent.EdgeType_BATCH,
		Provides: agent.EdgeType_BATCH,

		Options: map[string]*agent.OptionInfo{
			"timeFilter": {ValueTypes: []agent.ValueType{agent.ValueType_STRING, agent.ValueType_STRING}},
		},
	}

	return info, nil
}

// Initialze the handler based of the provided options.
func (fp *filterPoint) Init(r *agent.InitRequest) (*agent.InitResponse, error) {
	init := &agent.InitResponse{
		Success: true,
		Error:   "",
	}

	for _, opt := range r.Options {
		switch opt.Name {
		case "timeFilter":
			fp.timeMask = strings.TrimSpace(opt.Values[0].Value.(*agent.OptionValue_StringValue).StringValue)
			fp.timeZone = strings.TrimSpace(opt.Values[1].Value.(*agent.OptionValue_StringValue).StringValue)
		}
	}

	if len(fp.timeMask) == 0 {
		init.Success = false
		init.Error = "must supply 'timeFilter'"
	}

	return init, nil
}

// Create a snapshot of the running state of the process.
func (*filterPoint) Snapshot() (*agent.SnapshotResponse, error) {
	return &agent.SnapshotResponse{}, nil
}

// Restore a previous snapshot.
func (*filterPoint) Restore(req *agent.RestoreRequest) (*agent.RestoreResponse, error) {
	return &agent.RestoreResponse{
		Success: true,
	}, nil
}

// Start working with the next batch
func (fp *filterPoint) BeginBatch(begin *agent.BeginBatch) error {
	return nil
}

func (fp *filterPoint) Point(p *agent.Point) error {
	// Convert nanosecond epoch format time to timezone time
	dt := time.Unix(0, p.GetTime())
	timeZone := parseTimeZone(fp.timeZone, p)
	dt = converTimeToTimeZone(&dt, timeZone)

	// Only send back to Kapacitor the data points that match time mask
	if len(fp.timeMask) == 0 || matchtime.MatchTimeWithMask(fp.timeMask, &dt) {
		fp.agent.Responses <- &agent.Response{
			Message: &agent.Response_Point{
				Point: p,
			},
		}
	}

	return nil
}

func parseTimeZone(timezone string, p *agent.Point) string {
	if !strings.HasPrefix(timezone, "{") {
		return timezone
	}

	re, _ := regexp.Compile(`^\{(\S+)\}$`)
	match := re.FindStringSubmatch(timezone)

	if match == nil {
		return timezone
	}

	return utils.StringifyPointByKey(match[1], p)
}

func converTimeToTimeZone(t *time.Time, timeZone string) time.Time {
	// You can obtain the time zone code from this link
	// https://en.wikipedia.org/wiki/List_of_tz_database_time_zones,
	// in which the "TZ database name" (e.g. Pacific/Auckland) is used
	// as the time zone code.

	if len(timeZone) > 0 {
		loc, err := time.LoadLocation(timeZone)
		if err == nil {
			return t.In(loc)
		}
	}

	return *t
}

func (fp *filterPoint) EndBatch(end *agent.EndBatch) error {
	return nil
}

// Stop the handler gracefully.
func (fp *filterPoint) Stop() {
	close(fp.agent.Responses)
}

// Start is the entry point to start UDF
func Start() {
	a := agent.New(os.Stdin, os.Stdout)
	h := newFilterPoint(a)
	a.Handler = h

	log.Println("Starting agent 'filterPoint'")
	a.Start()
	err := a.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
