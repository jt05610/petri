package pipbot

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync/atomic"
)

type PipBot struct {
	Layout     *Layout
	Current    *Position
	TipChannel <-chan *Position
	client     io.ReadWriteCloser
	busy       atomic.Bool
	rx         chan []byte
	Rate       float64
	TipStart   int
	curTip     int
	cushion    float32
	steps      []*TransParams
	hasTip     bool
}

const (
	TipOffClear   float32 = 85
	TipOnClear    float32 = 142
	CushionVolume float32 = 25
)

func (b *PipBot) SetupDispenser() {
	m := []byte("M302 S1\n")

	_, err := b.client.Write(m)
	if err != nil {
		panic(err)
	}
	m = []byte("M82\n")

	_, err = b.client.Write(m)
	if err != nil {
		panic(err)
	}
	m = []byte("G92 E0\n")
	_, err = b.client.Write(m)
	if err != nil {
		panic(err)
	}
}

const (
	OutFile = "runFile.gcode"
)

// Init gets ready to run a protocol. Note that it automatically selects the last matrix as the tip matrix -- this will
// not be hardcoded in less time pressed versions :)
func (b *PipBot) Init() {
	b.TipChannel = b.Layout.Matrices[3].Channel()
	b.cushion = CushionVolume
	for b.curTip != b.TipStart {
		b.curTip++
		_ = <-b.TipChannel
	}
	b.Home()
	b.SetupDispenser()
	target := b.Current
	target.Z = TipOffClear
	m := []byte("G92 E-30\n")
	_, err := b.client.Write(m)
	if err != nil {
		panic(err)
	}
	b.Dispense()
	b.ResetCush()
	m = []byte("G92 E0\n")
	_, err = b.client.Write(m)
	if err != nil {
		panic(err)
	}
	b.Do(target)
}

func (b *PipBot) Bytes() []byte {
	return <-b.rx
}

// getTip gets the next tip position and increments the counter
func (b *PipBot) getTip() *Position {
	b.curTip++
	return <-b.TipChannel
}

func (b *PipBot) Transfer(src *Cell, dest *Cell, vol float32, eject bool) {
	// get increment tip id and pickup the tip
	if !b.hasTip {
		t := b.getTip()
		b.hasTip = true
		b.Do(t)
		t.Z = TipOnClear
		b.Do(t)
	}

	// go to source and insert into fluid
	t := src.Position
	tmp := t.Z
	b.Do(t)

	// draw fluid
	b.Pickup(vol)

	// remove from container
	t.Z = TipOnClear
	b.Do(t)
	t.Z = tmp
	// go to dest and insert into fluid
	t = dest.Position
	b.Do(t)

	// dispense fluid
	b.Dispense()
	// remove from container
	t.Z = TipOnClear
	b.Do(t)

	b.ResetCush()

	if eject {
		b.Eject()
		b.hasTip = false
	}

	m := []byte("M400\n")
	_, err := b.client.Write(m)
	if err != nil {
		panic(err)
	}
}

func NewPipBot(port string, baud int, firstTip int) *PipBot {
	var err error
	ret := &PipBot{
		rx:       make(chan []byte),
		Layout:   MakeGrid(31.5-29, 40-17),
		TipStart: firstTip,
		curTip:   0,
		hasTip:   false,
	}

	ret.client, err = os.Create(OutFile)

	if err != nil {
		panic(err)
	}
	return ret
}

func (b *PipBot) Close() {
	_ = b.client.Close()
}

func (b *PipBot) Do(target *Position) {
	b.GoTo(target)
}

func (b *PipBot) Pickup(volume float32) {
	travel := volume / 10
	target := Position{Z: 85}
	target.Z = target.Z + 1
	m := []byte(fmt.Sprintf("G1 F500 E-%v\n", travel))
	_, err := b.client.Write(m)
	if err != nil {
		panic(err)
	}
}

func (b *PipBot) Dispense() {
	m := []byte("G1 F500 E0\n")
	_, err := b.client.Write(m)
	if err != nil {
		panic(err)
	}
}

func (b *PipBot) ResetCush() {
	m := []byte("G1 F500 E0\n")
	_, err := b.client.Write(m)
	if err != nil {
		panic(err)
	}
}

func (b *PipBot) Home() {
	m := []byte("G28\n")
	_, err := b.client.Write(m)
	if err != nil {
		panic(err)
	}

	b.Current = &Position{
		X: 0,
		Y: 0,
		Z: 0,
	}
}

type TransParams struct {
	SrcRow int
	SrcCol int
	DstRow int
	DstCol int
	eject  bool
}

func (b *PipBot) Run() {
	for i, s := range b.steps {
		fmt.Println(fmt.Sprintf("Step %v/%v", i, len(b.steps)))
		b.Transfer(b.Layout.Matrices[2].Cells[s.SrcRow][s.SrcCol], b.Layout.Matrices[1].Cells[s.DstRow][s.DstCol], 100, s.eject)
	}
}

func (b *PipBot) Plan(file string) {
	b.steps = make([]*TransParams, 0)
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	rdr := csv.NewReader(f)
	rows, err := rdr.ReadAll()
	if err != nil {
		panic(err)
	}

	for i, row := range rows {
		for j, cell := range row[:12] {
			eject := false
			if j < 11 {
				if row[j+1] != cell {
					eject = true
				}
			}
			if j == 11 {
				if i != 7 {
					if rows[i+1][0] != cell {
						eject = true
					}
				}
			}
			switch cell {
			case "Blue":
				t := &TransParams{
					SrcRow: 0,
					SrcCol: 0,
					DstRow: i,
					DstCol: j,
					eject:  eject,
				}
				b.steps = append(b.steps, t)
			case "Red":
				t := &TransParams{
					SrcRow: 0,
					SrcCol: 1,
					DstRow: i,
					DstCol: j,
					eject:  eject,
				}
				b.steps = append(b.steps, t)
			case "Orange":
				t := &TransParams{
					SrcRow: 0,
					SrcCol: 2,
					DstRow: i,
					DstCol: j,
					eject:  eject,
				}
				b.steps = append(b.steps, t)
			}
		}
	}
}

func (b *PipBot) GoTo(p *Position) {
	target := b.Current
	target.X = p.X
	target.Y = p.Y
	_, err := b.client.Write(target.XY(b.Rate))
	if err != nil {
		panic(err)
	}
	target.Z = p.Z
	_, err = b.client.Write(target.Low(b.Rate))
	if err != nil {
		panic(err)
	}
	b.Current = target
}

func (b *PipBot) Eject() {
	target := &Position{
		X: 10,
		Y: b.Current.Y,
		Z: 154,
	}
	b.Do(target)
	target.Z = 85
	b.Do(target)
}

func (b *PipBot) Listen(ctx context.Context) bool {
	b.rx = make(chan []byte)
	scan := bufio.NewScanner(b.client)
	cont := true
	go func() {
		defer close(b.rx)
		for scan.Scan() {
			select {
			case <-ctx.Done():
				cont = false
			default:
				bytes := scan.Bytes()
				b.rx <- bytes
				fmt.Println(string(bytes))
			}
		}
	}()
	return cont
}
