package fracCollector

import (
	"context"
	"github.com/jt05610/petri/devices/fraction_collector/pipbot"
	marlin "github.com/jt05610/petri/marlin/proto/v1"
	"strconv"
	"time"
)

var (
	ClearHeight    = float32(55)
	DispenseHeight = float32(25)
	Speed          = float32(3000)

	wasteX = float32(0)
)

func (d *FractionCollector) rowIndexFromLetter(grid int, letter string) int {
	if letter[1] == '0' {
		return 0
	}
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
	wasteTime := (vol / d.flowRate) * 60000
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
		Speed: &Speed,
	})
	if err != nil {
		panic(err)
	}
	_, err = d.Move(context.Background(), &marlin.MoveRequest{
		Z:     &DispenseHeight,
		Speed: &Speed,
	})
	d.wasting = true
}

func (d *FractionCollector) Collect(ctx context.Context, req *CollectRequest) (*CollectResponse, error) {
	d.flowRate = req.FlowRate
	if d.wasting {
		_, err := d.FanOff(context.Background(), &marlin.FanOffRequest{})
		if err != nil {
			panic(err)
		}
		return &CollectResponse{}, nil
	}
	d.park(float32(req.WasteVol))
	_, err := d.FanOff(context.Background(), &marlin.FanOffRequest{})
	if err != nil {
		return nil, err
	}
	d.park(float32(req.CollectVol))
	_, err = d.FanOn(context.Background(), &marlin.FanOnRequest{})
	if err != nil {
		return nil, err
	}
	return &CollectResponse{}, nil
}

func (d *FractionCollector) Collected(ctx context.Context, req *CollectedRequest) (*CollectedResponse, error) {
	return &CollectedResponse{}, nil
}

func (d *FractionCollector) MoveTo(ctx context.Context, req *MoveToRequest) (*MoveToResponse, error) {
	matI, err := strconv.Atoi(req.Grid)
	if err != nil {
		return nil, err
	}
	row := d.rowIndexFromLetter(matI, req.Position)
	col, err := strconv.Atoi(req.Position[1:])
	if row == 0 && col == 0 {
		d.goToWaste()
		_, err = d.FanOff(context.Background(), &marlin.FanOffRequest{})
		if err != nil {
			return nil, err
		}
		return &MoveToResponse{}, nil
	}
	d.wasting = false
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

	return &MoveToResponse{}, nil
}
