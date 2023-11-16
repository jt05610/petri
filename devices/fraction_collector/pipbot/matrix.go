package pipbot

import (
	"fmt"
	"math"
	"strconv"
)

type CellType uint8

const (
	Tip CellType = iota
	Stock
	Standard
	Unknown
)

// Mixture describes what is in a cell.
// Mixtures can be stocks, mixtures of mixtures, dilutions, etc.
type Mixture struct {
	Contents map[Cell]float32
}

// Cell is the fundamental discrete addressable unit in the system.
// A cell can be a pipette tip position, an individual well of a plate, etc.
type Cell struct {
	Kind CellType
	*Position
	Volume float64
}

type FluidLevelMap map[string]float64

type FluidLevelDeltaFunc func(dV float64) float64

func CylinderFluidLevelFunc(diameterMM float64) FluidLevelDeltaFunc {
	area := math.Pi * math.Pow(diameterMM/2, 2)
	return func(dV float64) float64 {
		return dV / area
	}
}

func (m FluidLevelMap) dispense(dest string, dP float64) {
	m[dest] += dP
}

func (m FluidLevelMap) aspirate(src string, dP float64) {
	m[src] -= dP
}

func (m FluidLevelMap) changeVolume(dest string, dV float64, f FluidLevelDeltaFunc) {
	m[dest] += f(dV)
}

func (m FluidLevelMap) setLevel(dest string, p float64) {
	m[dest] = p
}

func (m FluidLevelMap) setVolume(dest string, base, vol float64, f FluidLevelDeltaFunc) {
	m[dest] = base + f(vol)
}

func (m FluidLevelMap) zero(basePosition float64) {
	for k := range m {
		m[k] = basePosition
	}
}

// Matrix is an aggregate of Cells. This can be a well plate, pipette tip box,
// tube rack, etc.
type Matrix struct {
	Name           string
	Cells          [][]*Cell
	Home           *Position
	FluidLevelFunc FluidLevelDeltaFunc
	FluidLevelMap  FluidLevelMap
	Diameter       float64
	Rows           int
	Columns        int
}

func (m *Matrix) Channel() <-chan *Position {
	res := make(chan *Position, m.Rows*m.Columns)
	go func() {
		for row := 0; row < m.Rows; row++ {
			for col := 0; col < m.Columns; col++ {
				res <- m.Cells[row][col].Position
			}
		}
	}()
	return res
}

// Layout describes how individual Matrix units are arranged on the build plate.
type Layout struct {
	Matrices []*Matrix
}

func NewMatrix(kind CellType, name string, home *Position, rowSpace, colSpace float32, nRow,
	nCol int, wellDiameter float64, deltaFunc FluidLevelDeltaFunc) *Matrix {
	m := &Matrix{
		Name:           name,
		Cells:          make([][]*Cell, nRow),
		Home:           home,
		Rows:           nRow,
		Columns:        nCol,
		Diameter:       wellDiameter,
		FluidLevelFunc: deltaFunc,
		FluidLevelMap:  make(FluidLevelMap, nRow*nCol),
	}
	m.FluidLevelMap.zero(float64(home.Z))
	for row := 0; row < nRow; row++ {
		m.Cells[row] = make([]*Cell, nCol)
		for col := 0; col < nCol; col++ {
			m.Cells[row][col] = &Cell{
				Kind: kind,
				Position: &Position{
					X: home.X + (float32(col) * colSpace),
					Y: home.Y + (float32(row) * rowSpace),
					Z: home.Z,
				},
			}
		}
	}
	return m
}

func (m *Matrix) asAlphaNumeric(row, col int) string {
	return fmt.Sprintf("%s%d", string(rune('A'+row)), col+1)
}

func (m *Matrix) FromAlphaNumeric(an string) (row, col int) {
	row = int(an[0] - 'A')
	colI, err := strconv.Atoi(an[1:])
	if err != nil {
		panic(err)
	}
	col = colI - 1
	return
}

func (m *Matrix) Identifier(row, col int) string {
	return fmt.Sprintf("%s_%s", m.Name, m.asAlphaNumeric(row, col))
}

func (m *Matrix) ChangeVolume(row, col int, vol float32) {
	current := m.Cells[row][col].Volume
	m.Cells[row][col].Volume = current + float64(vol)
	m.FluidLevelMap.changeVolume(m.asAlphaNumeric(row, col), float64(vol), m.FluidLevelFunc)
}

func (m *Matrix) DeltaP(dV float64) float64 {
	return m.FluidLevelFunc(dV)
}

func (m *Matrix) SetVolume(row, col int, vol float32) {
	m.Cells[row][col].Volume = float64(vol)
	base := m.Cells[row][col].Position.Z
	m.FluidLevelMap.setVolume(m.asAlphaNumeric(row, col), float64(base), float64(vol), m.FluidLevelFunc)
}

func (m *Matrix) SetLevel(row, col int, p float32) {
	m.FluidLevelMap.setLevel(m.asAlphaNumeric(row, col), float64(p))
}

func (m *Matrix) GetFluidLevel(row, col int) float64 {
	return m.FluidLevelMap[m.asAlphaNumeric(row, col)]
}
