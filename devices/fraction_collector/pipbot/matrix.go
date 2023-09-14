package pipbot

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
	Content *Mixture
}

// Matrix is an aggregate of Cells. This can be a well plate, pipette tip box,
// tube rack, etc.
type Matrix struct {
	Name    string
	Cells   [][]*Cell
	Home    *Position
	Rows    int
	Columns int
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
	nCol int) *Matrix {
	m := &Matrix{
		Name:    name,
		Cells:   make([][]*Cell, nRow),
		Home:    home,
		Rows:    nRow,
		Columns: nCol,
	}
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
