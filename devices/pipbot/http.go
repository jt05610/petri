package PipBot

import (
	"context"
	"encoding/json"
	"github.com/jt05610/petri/devices/fraction_collector/pipbot"
	"io"
	"net/http"
	"sort"
	"strings"
)

type SetSingleVolumeRequest struct {
	Grid     int     `json:"grid"`
	Position string  `json:"position"`
	Volume   float32 `json:"volume"`
}

type SetMultipleVolumeRequest []*SetSingleVolumeRequest

type Methods map[string]http.Handler

func (h Methods) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer func(r io.ReadCloser) {
		_, _ = io.Copy(io.Discard, r)
		_ = r.Close()
	}(r.Body)

	if handler, ok := h[r.Method]; ok {
		if handler == nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		} else {
			handler.ServeHTTP(w, r)
		}
		return
	}

	w.Header().Add("Allow", h.allowedMethods())
	if r.Method != http.MethodOptions {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (h Methods) allowedMethods() string {
	a := make([]string, 0, len(h))
	for k := range h {
		a = append(a, k)
	}
	sort.Strings(a)

	return strings.Join(a, ", ")
}

func (d *PipBot) VolumeHandler() http.Handler {
	return Methods{
		http.MethodGet:  http.HandlerFunc(d.GetVolumeHandler),
		http.MethodPost: http.HandlerFunc(d.PostVolumeHandler),
	}
}

func (d *PipBot) GetVolumeHandler(w http.ResponseWriter, _ *http.Request) {
	ret := d.Volumes()
	err := EncodeJSON(w, ret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type VolumeRequest map[string]pipbot.FluidLevelMap

type VolumeResponse map[string]pipbot.FluidLevelMap

func DecodeJSON(r io.Reader, v interface{}) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func EncodeJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	return enc.Encode(v)
}

func (d *PipBot) PostVolumeHandler(w http.ResponseWriter, r *http.Request) {
	var req VolumeRequest
	if err := DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err := d.SetVolumes(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	d.GetVolumeHandler(w, r)
}

func (d *PipBot) TipHandler() http.Handler {
	return Methods{
		http.MethodGet:  http.HandlerFunc(d.GetTipHandler),
		http.MethodPost: http.HandlerFunc(d.PostTipHandler),
	}
}

func (d *PipBot) GetTipHandler(w http.ResponseWriter, r *http.Request) {

}

type TipRequest struct {
	Grids        []int `json:"grids"`
	CurrentIndex int   `json:"current_index"`
}

type TipsResponse struct {
	CurrentIndex int `json:"current_index"`
}

func (d *PipBot) PostTipHandler(w http.ResponseWriter, r *http.Request) {
}

func (d *PipBot) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/volume", d.VolumeHandler())
	mux.Handle("/tip", d.TipHandler())
	mux.Handle("/transfer", d.TransferHandler())
	mux.Handle("/batch", d.BatchHandler())
	return mux
}

func (d *PipBot) TransferHandler() http.Handler {
	return Methods{
		http.MethodPost: http.HandlerFunc(d.PostTransferHandler),
		http.MethodGet:  http.HandlerFunc(d.GetTransferHandler),
	}
}

func (d *PipBot) PostTransferHandler(w http.ResponseWriter, r *http.Request) {
	var req StartTransferRequest
	if err := DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err := d.StartTransfer(context.Background(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	d.GetTransferHandler(w, r)
}

type IsTransferring struct {
	IsTransferring bool `json:"is_transferring"`
}

func (d *PipBot) GetTransferHandler(w http.ResponseWriter, r *http.Request) {
	isTransferring := d.transferring.Load()
	err := EncodeJSON(w, &IsTransferring{IsTransferring: isTransferring})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (d *PipBot) BatchHandler() http.Handler {
	return Methods{
		http.MethodPost: http.HandlerFunc(d.PostBatchHandler),
		http.MethodGet:  http.HandlerFunc(d.GetBatchHandler),
	}
}

func (d *PipBot) PostBatchHandler(w http.ResponseWriter, r *http.Request) {
	var req []StartTransferRequest
	if err := DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	d.GetBatchHandler(w, r)
}

func (d *PipBot) GetBatchHandler(w http.ResponseWriter, r *http.Request) {
	err := EncodeJSON(w, d.batch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
