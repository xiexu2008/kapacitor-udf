package interpolate

import (
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/influxdata/kapacitor/udf/agent"

	"pkg/utils"
)

type interpolateHandler struct {
	toField     string
	inputString string

	agent *agent.Agent
}

func newInterpolateHandler(agent *agent.Agent) *interpolateHandler {
	return &interpolateHandler{
		agent: agent,
	}
}

// Return the InfoResponse. Describing the properties of thip UDF agent.
func (*interpolateHandler) Info() (*agent.InfoResponse, error) {
	info := &agent.InfoResponse{
		Wants:    agent.EdgeType_BATCH,
		Provides: agent.EdgeType_BATCH,

		Options: map[string]*agent.OptionInfo{
			"string":  {ValueTypes: []agent.ValueType{agent.ValueType_STRING}},
			"toField": {ValueTypes: []agent.ValueType{agent.ValueType_STRING}},
		},
	}

	return info, nil
}

// Initialze the handler based of the provided options.
func (ip *interpolateHandler) Init(r *agent.InitRequest) (*agent.InitResponse, error) {
	init := &agent.InitResponse{
		Success: true,
		Error:   "",
	}

	for _, opt := range r.Options {
		switch opt.Name {
		case "string":
			ip.inputString = opt.Values[0].Value.(*agent.OptionValue_StringValue).StringValue
		case "toField":
			ip.toField = strings.TrimSpace(opt.Values[0].Value.(*agent.OptionValue_StringValue).StringValue)
		}
	}

	if len(ip.inputString) == 0 || len(ip.toField) == 0 {
		init.Success = false
		init.Error = "must supply 'toField' and 'string'"
	}

	return init, nil
}

// Create a snapshot of the running state of the process.
func (*interpolateHandler) Snapshot() (*agent.SnapshotResponse, error) {
	return &agent.SnapshotResponse{}, nil
}

// Restore a previous snapshot.
func (*interpolateHandler) Restore(req *agent.RestoreRequest) (*agent.RestoreResponse, error) {
	return &agent.RestoreResponse{
		Success: true,
	}, nil
}

// Start working with the next batch
func (*interpolateHandler) BeginBatch(begin *agent.BeginBatch) error {
	return nil
}

func (ip *interpolateHandler) Point(p *agent.Point) error {
	// Interpolate the string and save it to the 'FieldsString'
	strInterplolated, _ := interplolateString(ip.inputString, p)

	if p.FieldsString == nil {
		p.FieldsString = make(map[string]string)
	}

	p.FieldsString[ip.toField] = strInterplolated

	// Send the new data point back to Kapacitor
	ip.agent.Responses <- &agent.Response{
		Message: &agent.Response_Point{
			Point: p,
		},
	}

	return nil
}

func interplolateString(str string, p *agent.Point) (string, error) {
	// To interpolate string like "Lower {lowerThresh} upper {upperThresh} within {withinSec}s"
	// with the fields or tags defined in Kapacitor pointer

	var sb strings.Builder
	var keyName strings.Builder
	isKeyName := false
	for _, ch := range str {
		if ch == '{' {
			isKeyName = true
			keyName.Reset()
		} else if ch == '}' {
			key := keyName.String()
			val := utils.StringifyPointByKey(key, p)
			sb.WriteString(val)

			isKeyName = false
			keyName.Reset()
		} else {
			if isKeyName {
				if !unicode.IsSpace(ch) {
					keyName.WriteRune(ch)
				}
			} else {
				sb.WriteRune(ch)
			}
		}
	}

	return sb.String(), nil
}

func (ip *interpolateHandler) EndBatch(end *agent.EndBatch) error {
	return nil
}

// Stop the handler gracefully.
func (ip *interpolateHandler) Stop() {
	close(ip.agent.Responses)
}

// Start is the entry point of starting UDF interpolate
func Start() {
	a := agent.New(os.Stdin, os.Stdout)
	h := newInterpolateHandler(a)
	a.Handler = h

	log.Println("Starting agent 'interpolate'")
	a.Start()
	err := a.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
