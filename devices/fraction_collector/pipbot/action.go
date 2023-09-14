package pipbot

import (
	"fmt"
	"time"
)

type Position struct {
	X float32
	Y float32
	Z float32
}

func (p *Position) XY(rate ...float64) []byte {
	var fr float64
	if len(rate) > 1 {
		fr = rate[0]
	} else {
		fr = 1000
	}
	return []byte(fmt.Sprintf("G1 F%v X%v Y%v Z%v\n", fr, p.X, p.Y, p.Z))
}

func (p *Position) Low(rate ...float64) []byte {
	var fr float64
	if len(rate) > 1 {
		fr = rate[0]
	} else {
		fr = 1000
	}
	return []byte(fmt.Sprintf("G1 F%v Z%v\n", fr, p.Z))
}

type Transfer struct {
	Tip       *Position
	TipChange bool
	Src       *Position
	Dest      *Position
	Volume    float32
	slope     float32
	intercept float32
}

func (t *Transfer) getTip() {

}

func (t *Transfer) drawFluid() {

}

func (t *Transfer) dispenseFluid() {

}

func (t *Transfer) ejectTip() {

}

func (t *Transfer) Bytes() [][]byte {
	return [][]byte{}
}

func (t *Transfer) Finish() {

}

type Heat struct {
	Duration    time.Duration
	Temperature float32
}

func (t *Transfer) Heat() {

}

func (h *Heat) Bytes() [][]byte {
	return [][]byte{}

}

type Shake struct {
	Duration    time.Duration
	Temperature float32
}

func (s *Shake) Bytes() [][]byte {
	return [][]byte{}
}

type Action interface {
	Bytes() [][]byte
}

func Do(a Action, waitForFinish bool) <-chan []byte {
	res := make(chan []byte)
	go func(a Action) {
		defer close(res)
		for _, b := range a.Bytes() {
			res <- b
		}
	}(a)
	return res
}
