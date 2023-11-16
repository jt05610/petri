package PipBot

import (
	"context"
	"github.com/jt05610/petri/devices/fraction_collector/pipbot"
	marlin "github.com/jt05610/petri/marlin/proto/v1"
	"go.uber.org/zap"
	"strconv"
)

type Location struct {
	Grid int `json:"grid"`
	Row  int `json:"row"`
	Col  int `json:"col"`
}

func (l *Location) Equals(other *Location) bool {
	return l.Grid == other.Grid && l.Row == other.Row && l.Col == other.Col
}

func (l *Location) Position(layout *pipbot.Layout) *pipbot.Position {
	return layout.Matrices[l.Grid].Cells[l.Row][l.Col].Position
}

type TransferPlan struct {
	Technique      PipetteTechnique
	Source         *Location
	Dest           *Location
	ReplaceTip     bool
	PreRinses      int
	FillVolume     float32
	ExcessVolume   float32
	AspirationRate float32
	DispenseVolume float32
	DispenseRate   float32
	DwellTime      int
}

var (
	TipOffClearZ = float32(96)
	TipOnClearZ  = float32(150)
	EjectX       = float32(10)
	EjectZ       = float32(165)
	MoveSpeed    = float32(4000)
	Slope        = float32(1)
	Intercept    = float32(0)
	TipDepth     = float32(2)
)

func (p *TransferPlan) needsNewTip(s *State) bool {
	if !s.HasTip || p.ReplaceTip {
		return true
	}
	if p.Source != nil && s.Pipette.Source != nil {

		if !p.Source.Equals(s.Pipette.Source) {
			if p.Source.Equals(s.Pipette.LastDest) {
				return false
			}
			return true
		}

		return false
	}
	return false
}

func (p *TransferPlan) clear(s *State) *marlin.MoveRequest {
	var z *float32
	if s.HasTip {
		z = &TipOnClearZ
	} else {
		z = &TipOffClearZ
	}
	return &marlin.MoveRequest{
		Z:     z,
		Speed: &MoveSpeed,
	}
}

func (p *TransferPlan) ejectTip(s *State) (*State, []*marlin.MoveRequest) {
	return &State{
			Pipette:       nil,
			HasTip:        false,
			TipChannel:    s.TipChannel,
			Layout:        s.Layout,
			TipIndex:      s.TipIndex,
			needsPreRinse: s.needsPreRinse,
			extruderPos:   s.extruderPos,
		},
		[]*marlin.MoveRequest{
			p.clear(s),
			{
				X:     &EjectX,
				Speed: &MoveSpeed,
			},
			{
				Z:     &EjectZ,
				Speed: &MoveSpeed,
			},
			p.clear(s),
		}
}

func (p *TransferPlan) XY(pos *pipbot.Position) *marlin.MoveRequest {
	return &marlin.MoveRequest{
		X:     &pos.X,
		Y:     &pos.Y,
		Speed: &MoveSpeed,
	}
}

func (p *TransferPlan) Z(pos *pipbot.Position) *marlin.MoveRequest {
	return &marlin.MoveRequest{
		Z:     &pos.Z,
		Speed: &MoveSpeed,
	}
}

func (p *TransferPlan) getTip(s *State) (*State, []*marlin.MoveRequest) {
	tipLoc := <-s.TipChannel
	newState := &State{
		Pipette:       nil,
		HasTip:        true,
		TipChannel:    s.TipChannel,
		TipIndex:      s.TipIndex + 1,
		Layout:        s.Layout,
		needsPreRinse: p.PreRinses > 0,
		extruderPos:   s.extruderPos,
	}
	raisePos := &pipbot.Position{
		Z: tipLoc.Z + 5,
	}
	lowPos := &pipbot.Position{
		Z: tipLoc.Z - 1,
	}
	return newState, []*marlin.MoveRequest{
		p.clear(s),
		p.XY(tipLoc),
		p.Z(tipLoc),
		p.Z(raisePos),
		p.Z(lowPos),
		p.Z(raisePos),
		p.Z(lowPos),
		p.clear(newState),
	}
}

func (p *TransferPlan) changeTip(s *State) (*State, []*marlin.MoveRequest) {
	ret := make([]*marlin.MoveRequest, 0)
	if s.HasTip {
		newState, moves := p.ejectTip(s)
		ret = append(ret, moves...)
		s = newState
	}
	newState, moves := p.getTip(s)
	ret = append(ret, moves...)
	newState.needsPreRinse = true
	return newState, ret
}

func (p *TransferPlan) initialize(s *State) (*State, []*marlin.MoveRequest) {
	ret := make([]*marlin.MoveRequest, 0)
	if p.needsNewTip(s) {
		newState, moves := p.changeTip(s)
		ret = append(ret, moves...)
		s = newState
	}
	return s, ret
}

func (p *TransferPlan) moveToSource(s *State) (*State, []*marlin.MoveRequest) {
	mat := s.Layout.Matrices[p.Source.Grid]
	z := float32(mat.GetFluidLevel(p.Source.Row, p.Source.Col)) - TipDepth
	ret := []*marlin.MoveRequest{
		p.clear(s),
		p.XY(p.Source.Position(s.Layout)),
	}
	ret = append(ret, p.Z(&pipbot.Position{Z: z}))
	return s, ret
}

func volToMM(vol float32, intercept, slope float32) *float32 {
	res := (vol - intercept) / slope
	return &res
}

func (p *TransferPlan) pickup(s *State, vol float32) (*State, *marlin.MoveRequest) {
	curVol := float32(0)
	if s.Pipette != nil {
		curVol = s.Pipette.Volume
	}
	dist := volToMM(vol, Intercept, Slope)
	*dist *= -1
	extruderPos := s.extruderPos + *dist
	s.Layout.Matrices[p.Source.Grid].ChangeVolume(p.Source.Row, p.Source.Col, -vol)
	return &State{
			Pipette:       &Fluid{Source: p.Source, Volume: curVol + vol, LastDest: p.Dest},
			HasTip:        true,
			TipChannel:    s.TipChannel,
			Layout:        s.Layout,
			needsPreRinse: s.needsPreRinse,
			TipIndex:      s.TipIndex,
			extruderPos:   extruderPos,
		}, &marlin.MoveRequest{
			E:     dist,
			Speed: volToMM(p.AspirationRate, Intercept, Slope),
		}
}

func (p *TransferPlan) deliver(s *State, vol, offset float32) (*State, *marlin.MoveRequest) {
	curVol := float32(0)
	if s.Pipette != nil {
		curVol = s.Pipette.Volume
	}
	newVol := curVol - vol
	if newVol < 0 {
		panic("not enough fluid")
	}
	newFluid := &Fluid{Source: s.Pipette.Source, Volume: newVol, LastDest: p.Dest}
	dist := volToMM(vol, Intercept, Slope)
	ePos := s.extruderPos + *dist

	s.Layout.Matrices[p.Dest.Grid].ChangeVolume(p.Dest.Row, p.Dest.Col, vol)
	return &State{
			Pipette:       newFluid,
			HasTip:        true,
			TipChannel:    s.TipChannel,
			Layout:        s.Layout,
			needsPreRinse: s.needsPreRinse,
			TipIndex:      s.TipIndex,
			extruderPos:   ePos,
		}, &marlin.MoveRequest{
			E:     &ePos,
			Speed: volToMM(p.AspirationRate, Intercept, Slope),
		}
}

func (p *TransferPlan) rinse(s *State, vol float32) (*State, []*marlin.MoveRequest) {
	ret := make([]*marlin.MoveRequest, 2)
	s, ret[0] = p.pickup(s, vol)
	s, ret[1] = p.deliver(s, vol, 0)
	return s, ret
}

func (p *TransferPlan) aspirate(s *State) (*State, []*marlin.MoveRequest) {
	ret := make([]*marlin.MoveRequest, 0)
	doPreRinse := false
	if s.needsPreRinse {
		doPreRinse = true
	}
	if doPreRinse {
		for i := 0; i < p.PreRinses; i++ {
			newState, moves := p.rinse(s, p.FillVolume)
			ret = append(ret, moves...)
			s = newState
		}
		s.needsPreRinse = false
	}
	newState, move := p.pickup(s, p.FillVolume)
	ret = append(ret, move)
	return newState, ret
}

func (p *TransferPlan) moveToDest(s *State) (*State, []*marlin.MoveRequest) {
	mat := s.Layout.Matrices[p.Dest.Grid]
	z := float32(mat.GetFluidLevel(p.Source.Row, p.Source.Col))
	return s, []*marlin.MoveRequest{
		p.clear(s),
		p.XY(p.Dest.Position(s.Layout)),
		p.Z(&pipbot.Position{Z: z}),
	}
}

func (p *TransferPlan) dispense(s *State) (*State, []*marlin.MoveRequest) {
	ret := make([]*marlin.MoveRequest, 0)
	newState, move := p.deliver(s, p.DispenseVolume, 0)
	ret = append(ret, move)
	return newState, ret
}

func (p *TransferPlan) finish(s *State) (*State, []*marlin.MoveRequest) {
	ret := []*marlin.MoveRequest{
		p.clear(s),
	}
	return s, ret
}

type movePlan func(*State) (*State, []*marlin.MoveRequest)

func (p *TransferPlan) Moves(s *State) (*State, []*marlin.MoveRequest, error) {
	ret := make([]*marlin.MoveRequest, 0)
	handlers := []movePlan{
		p.initialize,
		p.moveToSource,
		p.aspirate,
		p.moveToDest,
		p.dispense,
		p.finish,
	}
	for _, h := range handlers {
		newState, moves := h(s)
		ret = append(ret, moves...)
		s = newState
	}
	return s, ret, nil
}

func (d *PipBot) rowIndexFromLetter(grid int, letter string) int {
	if letter[1] == '0' {
		return 0
	}
	diff := int(letter[0]) - int('A')
	current := d.State()
	return current.Matrices[grid].Rows - diff - 1
}

func (d *PipBot) convertGridPos(grid string, pos string) (*Location, error) {
	matI, err := strconv.Atoi(grid)
	if err != nil {
		return nil, err
	}
	row := d.rowIndexFromLetter(matI, pos)
	col, err := strconv.Atoi(pos[1:])
	if row == 0 && col == 0 {
		return &Location{
			Grid: 0,
			Row:  0,
			Col:  0,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &Location{
		Grid: matI,
		Row:  row,
		Col:  col - 1,
	}, nil
}

func (d *PipBot) makePlan(req *StartTransferRequest) (*TransferPlan, error) {
	source, err := d.convertGridPos(req.SourceGrid, req.SourcePos)
	if err != nil {
		return nil, err
	}
	dest, err := d.convertGridPos(req.DestGrid, req.DestPos)
	if err != nil {
		return nil, err
	}
	return &TransferPlan{
		Technique:      PipetteTechnique(int(req.Technique)),
		Source:         source,
		Dest:           dest,
		ReplaceTip:     req.Multi <= 0,
		PreRinses:      2,
		FillVolume:     float32(req.Volume),
		DispenseVolume: float32(req.Volume),
		AspirationRate: float32(req.FlowRate),
		DispenseRate:   float32(req.FlowRate),
		DwellTime:      0,
	}, nil
}

func (d *PipBot) do(ctx context.Context, mr []*marlin.MoveRequest) error {
	nMoves := len(mr)
	for i, r := range mr {
		d.logger.Info("Move", zap.Int("move", i), zap.Int("total", nMoves), zap.Any("request", r))
		select {
		case <-ctx.Done():
			return nil
		default:
			d.logger.Info("Sending move", zap.Any("request", r))
			_, err := d.Move(ctx, r)
			if err != nil {
				d.logger.Error("Error sending move", zap.Error(err))
				return err
			}
		}
	}
	return nil
}

func (d *PipBot) StartTransfer(ctx context.Context, req *StartTransferRequest) (*StartTransferResponse, error) {
	plan, err := d.makePlan(req)
	if err != nil {
		return nil, err
	}
	d.transferring.Store(true)
	initialState := d.State()
	go func() {
		newState, moves, err := plan.Moves(initialState)
		if err != nil {
			panic(err)
		}
		err = d.do(ctx, moves)
		if err != nil {
			panic(err)
		}
		d.state.Store(newState)
		d.transferring.Store(false)
	}()
	return &StartTransferResponse{}, nil
}

func (d *PipBot) FinishTransfer(ctx context.Context, _ *FinishTransferRequest) (*FinishTransferResponse, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if !d.transferring.Load() {
				return nil, nil
			}
		}
	}
}
