package fracCollector

import (
	"github.com/jt05610/petri/devices/fraction_collector/pipbot"
	marlin "github.com/jt05610/petri/marlin/proto/v1"
)

type FractionCollector struct {
	marlin.MarlinServer
	*pipbot.Layout
}

type CollectRequest struct {
	Position   string  `json:"position"`
	Grid       string  `json:"grid"`
	WasteVol   float64 `json:"wastevol"`
	CollectVol float64 `json:"collectvol"`
}

type CollectResponse struct {
	Position   string  `json:"position"`
	Grid       string  `json:"grid"`
	WasteVol   float64 `json:"wastevol"`
	CollectVol float64 `json:"collectvol"`
}

type CollectedRequest struct {
	Position string `json:"position"`
	Grid     string `json:"grid"`
}

type CollectedResponse struct {
	Position string `json:"position"`
	Grid     string `json:"grid"`
}
