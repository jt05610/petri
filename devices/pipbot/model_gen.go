package PipBot

import (
	"context"
	"fmt"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/device"
	"github.com/jt05610/petri/devices/fraction_collector/pipbot"
	"github.com/jt05610/petri/labeled"
	marlin "github.com/jt05610/petri/marlin/proto/v1"
	"github.com/jt05610/petri/yaml"
	"log"
	"strconv"
	"sync/atomic"
)

var InitialState = &State{
	Pipette: &Fluid{
		Source: &Location{
			Grid: 0,
			Row:  0,
			Col:  0,
		},
		Volume: 0,
	},
	HasTip: false,
}

func MakePlateGrid(xOffset, yOffset float32) *pipbot.Layout {
	ret := &pipbot.Layout{
		Matrices: make([]*pipbot.Matrix, 4),
	}

	ret.Matrices[0] = pipbot.NewMatrix(pipbot.Unknown, "Purp", &pipbot.Position{X: 29 + xOffset, Y: 17 + yOffset, Z: 59.4}, 42.5-29,
		42.5-29, 5, 16, 9.32, pipbot.CylinderFluidLevelFunc(9.32))

	ret.Matrices[1] = pipbot.NewMatrix(pipbot.Unknown, "96", &pipbot.Position{X: 35.7 + xOffset, Y: 92.9 + yOffset, Z: 61},
		9,
		9, 8, 12, 6, pipbot.CylinderFluidLevelFunc(6))

	ret.Matrices[2] = pipbot.NewMatrix(pipbot.Stock, "12", &pipbot.Position{X: 46 + xOffset, Y: 178.5 + xOffset, Z: 61},
		72-46,
		72-46, 3, 4, 21.3, pipbot.CylinderFluidLevelFunc(21.3))

	ret.Matrices[3] = pipbot.NewMatrix(pipbot.Tip, "tips", &pipbot.Position{
		X: 164 + xOffset,
		Y: 106 + yOffset,
		Z: 61,
	}, 8.87, 8.87, 12, 8, 12, nil)

	return ret
}

func AutosamplerGrid(xOffset, yOffset float32) *pipbot.Layout {
	ret := &pipbot.Layout{
		Matrices: make([]*pipbot.Matrix, 5),
	}

	ret.Matrices[0] = pipbot.NewMatrix(pipbot.Unknown, "autosampler", &pipbot.Position{X: 37.5 + xOffset, Y: 55.2 + yOffset, Z: 91.7}, 15.1,
		15.1, 10, 10, 9.75, pipbot.CylinderFluidLevelFunc(9.75))

	ret.Matrices[1] = pipbot.NewMatrix(pipbot.Stock, "diluent", &pipbot.Position{X: 28.8 + xOffset, Y: 230 + yOffset, Z: 87},
		25,
		25, 1, 3, 19.5, pipbot.CylinderFluidLevelFunc(19.5))

	ret.Matrices[2] = pipbot.NewMatrix(pipbot.Stock, "lipid", &pipbot.Position{X: 100.9 + xOffset, Y: 215.6, Z: 64.6},
		20,
		20, 1, 5, 12.7, pipbot.CylinderFluidLevelFunc(12.7))

	ret.Matrices[3] = pipbot.NewMatrix(pipbot.Tip, "tip1", &pipbot.Position{
		X: 200.6 + xOffset,
		Y: 147.9 + yOffset,
		Z: 66,
	}, 8.87, 8.87, 10, 4, 1, nil,
	)
	ret.Matrices[4] = pipbot.NewMatrix(pipbot.Tip, "tip2", &pipbot.Position{
		X: 199.5 + xOffset,
		Y: 25.3 + yOffset,
		Z: 66,
	}, 8.87, 8.87, 12, 4, 1, nil,
	)
	return ret
}

func (d *PipBot) LoadMatrix(ctx context.Context, m *pipbot.Matrix) error {
	levelMap, err := d.RedisClient.Load(ctx, m.Name)
	if err != nil {
		return err
	}
	if levelMap == nil {
		return nil
	}
	for k, v := range levelMap {
		m.FluidLevelMap[k] = v
		row, col := m.FromAlphaNumeric(k)
		m.SetLevel(row, col, float32(v))
	}
	for row := 0; row < m.Rows; row++ {
		for col := 0; col < m.Columns; col++ {
			level := m.Cells[row][col].Position.Z
			m.SetLevel(row, col, level)
		}
	}
	return nil
}

func (d *PipBot) loadMatrices(ctx context.Context) error {
	current := d.State()
	for _, m := range current.Layout.Matrices {
		err := d.LoadMatrix(ctx, m)
		if err != nil {
			return err
		}
	}
	d.state.Store(current)
	return nil
}

func (d *PipBot) ChannelTips(matrices []int) chan *pipbot.Position {
	ret := make(chan *pipbot.Position)
	go func() {
		defer close(ret)
		for _, m := range matrices {
			for pos := range d.State().Layout.Matrices[m].Channel() {
				ret <- pos
			}
		}
	}()
	return ret
}

func (d *PipBot) TipsSetup(req *TipRequest) error {
	state := d.State()
	state.TipChannel = d.ChannelTips(req.Grids)
	state = discardTips(req.CurrentIndex, state)
	d.state.Store(state)
	return nil
}

func (d *PipBot) CurrentTip() *TipsResponse {
	state := d.State()
	return &TipsResponse{
		CurrentIndex: state.TipIndex,
	}
}

type Grid string

const (
	Autosampler Grid = "autosampler"
	Plate       Grid = "plate"
)

func NewPipBot(grid Grid, tipGrids []int, firstTip int, srv marlin.MarlinServer) *PipBot {
	d := &PipBot{
		MarlinServer: srv,
		state:        new(atomic.Pointer[State]),
		transferCh:   make(chan *TransferPlan),
		transferring: new(atomic.Bool),
		RedisClient:  NewRedisClient(":"),
	}
	state := InitialState
	state.TipIndex = firstTip
	d.transferring.Store(false)
	if grid == Autosampler {
		state.Layout = AutosamplerGrid(0, 0)
	} else {
		state.Layout = MakePlateGrid(0, 0)
	}

	state.TipChannel = d.ChannelTips(tipGrids)
	state = discardTips(firstTip, state)
	d.state.Store(state)
	err := d.loadMatrices(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return d
}

func (d *PipBot) Close() error {
	for _, m := range d.State().Layout.Matrices {
		err := d.Flush(context.Background(), m.Name, m.FluidLevelMap)
		if err != nil {
			return err
		}
	}
	return nil
}

func discardTips(n int, s *State) *State {
	for i := 0; i < n; i++ {
		<-s.TipChannel
	}
	return s
}

func (d *PipBot) load() *device.Device {
	srv := yaml.Service{}
	df, err := deviceYaml.Open("device.yaml")
	if err != nil {
		log.Fatal(err)
	}
	dev, err := srv.Load(df)
	if err != nil {
		log.Fatal(err)
	}
	ret, err := srv.ToNet(dev, d.Handlers())
	if err != nil {
		log.Fatal(err)
	}
	return ret
}

func (d *PipBot) SetVolume(location *Location, vol float64) {
	current := d.State()
	m := current.Layout.Matrices[location.Grid]
	m.SetVolume(location.Row, location.Col, float32(vol))
	current.Layout.Matrices[location.Grid] = m
	d.state.Store(current)
}
func (d *PipBot) SetVolumes(req VolumeRequest) error {
	current := d.State()
	for grid, positions := range req {
		gridI, err := strconv.Atoi(grid)
		if err != nil {
			return err
		}
		m := current.Layout.Matrices[gridI]
		for pos, vol := range positions {
			loc, err := d.convertGridPos(grid, pos)
			if err != nil {
				return err
			}
			m.SetVolume(loc.Row, loc.Col, float32(vol))
		}
		current.Layout.Matrices[gridI] = m
	}
	d.state.Store(current)
	return nil
}

func (d *PipBot) SetLevel(location *Location, level float32) {
	current := d.State()
	m := current.Layout.Matrices[location.Grid]
	m.SetLevel(location.Row, location.Col, level)
	current.Layout.Matrices[location.Grid] = m
	d.state.Store(current)
}

func (d *PipBot) ChangeVolume(location *Location, vol float32) {
	current := d.State()
	m := current.Layout.Matrices[location.Grid]
	m.ChangeVolume(location.Row, location.Col, vol)
	current.Layout.Matrices[location.Grid] = m
	d.state.Store(current)
}

func (r *StartTransferRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "starttransfer",
	}
}

func (r *StartTransferRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "starttransfer" {
		return fmt.Errorf("expected event name starttransfer, got %s", event.Name)
	}
	if event.Data["sourceGrid"] != nil {
		ds := event.Data["sourceGrid"].(string)

		r.SourceGrid = ds
	}

	if event.Data["destGrid"] != nil {
		ds := event.Data["destGrid"].(string)

		r.DestGrid = ds
	}

	if event.Data["sourcePos"] != nil {
		ds := event.Data["sourcePos"].(string)

		r.SourcePos = ds
	}

	if event.Data["destPos"] != nil {
		ds := event.Data["destPos"].(string)

		r.DestPos = ds
	}

	if event.Data["multi"] != nil {
		ds := event.Data["multi"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.Multi = d
	}

	if event.Data["flowRate"] != nil {
		ds := event.Data["flowRate"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.FlowRate = d
	}

	if event.Data["volume"] != nil {
		ds := event.Data["volume"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.Volume = d
	}

	if event.Data["technique"] != nil {
		ds := event.Data["technique"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.Technique = d
	}

	return nil
}

func (r *StartTransferResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "starttransfer",
		Fields: []*labeled.Field{
			{
				Name: "sourceGrid",
				Type: "string",
			},

			{
				Name: "destGrid",
				Type: "string",
			},

			{
				Name: "sourcePos",
				Type: "string",
			},

			{
				Name: "destPos",
				Type: "string",
			},

			{
				Name: "multi",
				Type: "number",
			},

			{
				Name: "flowRate",
				Type: "number",
			},

			{
				Name: "volume",
				Type: "number",
			},

			{
				Name: "technique",
				Type: "number",
			},
		},
		Data: map[string]interface{}{
			"SourceGrid": r.SourceGrid,
			"DestGrid":   r.DestGrid,
			"SourcePos":  r.SourcePos,
			"DestPos":    r.DestPos,
			"Multi":      r.Multi,
			"FlowRate":   r.FlowRate,
			"Volume":     r.Volume,
			"Technique":  r.Technique,
		},
	}

	return ret
}

func (r *StartTransferResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "starttransfer" {
		return fmt.Errorf("expected event name starttransfer, got %s", event.Name)
	}
	return nil
}

func (r *FinishTransferRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "finishtransfer",
	}
}

func (r *FinishTransferRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "finishtransfer" {
		return fmt.Errorf("expected event name finishtransfer, got %s", event.Name)
	}
	if event.Data["message"] != nil {
		ds := event.Data["message"].(string)

		r.Message = ds
	}

	return nil
}

func (r *FinishTransferResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "finishtransfer",
		Fields: []*labeled.Field{
			{
				Name: "message",
				Type: "string",
			},
		},
		Data: map[string]interface{}{
			"Message": r.Message,
		},
	}

	return ret
}

func (r *FinishTransferResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "finishtransfer" {
		return fmt.Errorf("expected event name finishtransfer, got %s", event.Name)
	}
	return nil
}

func (d *PipBot) Handlers() control.Handlers {
	return control.Handlers{

		"starttransfer": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(StartTransferRequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.StartTransfer(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},

		"finishtransfer": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(FinishTransferRequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.FinishTransfer(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},
	}
}
