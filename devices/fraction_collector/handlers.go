package fracCollector

import (
	"context"
	"errors"
	"github.com/jt05610/petri/devices/fraction_collector/pipbot"
	marlin "github.com/jt05610/petri/marlin/proto/v1"
	"strconv"
	"time"
)

var (
	ClearHeight    = float32(40)
	DispenseHeight = float32(0)
	Speed          = float32(1000)
	flowRate       = float32(2) // mL/min
	wasteX         = float32(0)
	wasteY         = float32(0)
)

func (d *FractionCollector) rowIndexFromLetter(grid int, letter string) int {
	diff := int(letter[0]) - int('A')
	return d.Matrices[grid].Rows - diff - 1
}

func (d *FractionCollector) goTo(p *pipbot.Cell) {
	_, err := d.Move(context.Background(), &marlin.MoveRequest{
		Z:     &ClearHeight,
		Speed: &Speed,
	})
	if err != nil {
		panic(err)
	}
	_, err = d.Move(context.Background(), &marlin.MoveRequest{
		X:     &p.X,
		Y:     &p.Y,
		Speed: &Speed,
	})
	if err != nil {
		panic(err)
	}
	_, err = d.Move(context.Background(), &marlin.MoveRequest{
		Z:     &DispenseHeight,
		Speed: &Speed,
	})
}

func (d *FractionCollector) park(vol float32) {
	wasteTime := (vol / flowRate) * 60000
	time.Sleep(time.Duration(wasteTime) * time.Millisecond)
}

func (d *FractionCollector) goToWaste() {
	_, err := d.Move(context.Background(), &marlin.MoveRequest{
		Z:     &ClearHeight,
		Speed: &Speed,
	})
	if err != nil {
		panic(err)
	}
	_, err = d.Move(context.Background(), &marlin.MoveRequest{
		X:     &wasteX,
		Y:     &wasteY,
		Speed: &Speed,
	})
	if err != nil {
		panic(err)
	}
	_, err = d.Move(context.Background(), &marlin.MoveRequest{
		Z:     &DispenseHeight,
		Speed: &Speed,
	})
}

func (d *FractionCollector) Collect(ctx context.Context, req *CollectRequest) (*CollectResponse, error) {
	matI, err := strconv.Atoi(req.Grid)
	if err != nil {
		return nil, err
	}
	row := d.rowIndexFromLetter(matI, req.Position)
	col, err := strconv.Atoi(req.Position[1:])
	if err != nil {
		return nil, err
	}
	target := d.Layout.Matrices[matI].Cells[row][col-1]
	_, err = d.FanOn(context.Background(), &marlin.FanOnRequest{})
	if err != nil {
		return nil, err
	}
	d.goTo(target)
	if err != nil {
		return nil, err
	}
	d.park(float32(req.WasteVol))
	_, err = d.FanOff(context.Background(), &marlin.FanOffRequest{})
	if err != nil {
		return nil, err
	}
	d.park(float32(req.WasteVol))
	_, err = d.FanOn(context.Background(), &marlin.FanOnRequest{})
	if err != nil {
		return nil, err
	}
	d.goToWaste()
	_, err = d.FanOff(context.Background(), &marlin.FanOffRequest{})
	if err != nil {
		return nil, err
	}
	return &CollectResponse{}, errors.New("not implemented")
}

func (d *FractionCollector) Collected(ctx context.Context, req *CollectedRequest) (*CollectedResponse, error) {
	return &CollectedResponse{}, errors.New("not implemented")
}
