package pipbot

const Port = "COM5"

// const Port = "/dev/cu.usbserial-1110"

const Baud = 115200

func MakeGrid(xOffset, yOffset float32) *Layout {
	ret := &Layout{
		Matrices: make([]*Matrix, 4),
	}

	ret.Matrices[0] = NewMatrix(Unknown, "Purp", &Position{X: 29 + xOffset, Y: 17 + yOffset, Z: 40}, 42.5-29,
		42.5-29, 5, 16, 12, nil)

	ret.Matrices[1] = NewMatrix(Unknown, "96", &Position{X: 35.5 + xOffset, Y: 86.5 + yOffset, Z: 74.5},
		9,
		9, 8, 12, 12, nil)

	ret.Matrices[2] = NewMatrix(Stock, "12", &Position{X: 46 + xOffset, Y: 178.5 + xOffset, Z: 75},
		72-46,
		72-46, 3, 4, 12, nil)

	ret.Matrices[3] = NewMatrix(Tip, "tips", &Position{
		X: 165 + xOffset,
		Y: 103.5 + yOffset,
		Z: 73.5,
	}, 173.8-165, 173.8-165, 12, 8, 12, nil)

	return ret
}
