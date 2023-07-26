package petri_test

import (
	"github.com/jt05610/petri"
	"testing"
)

func runServiceTest(srv petri.Service, t *testing.T) {
	t.Run("CreateNet", func(t *testing.T) {
		n, err := srv.CreateNet("test")
		if err != nil {
			t.Error(err)
		}
		if n.Name != "test" {
			t.Error("Name not set")
		}
		if n.ID == "" {
			t.Error("ID not set")
		}
	})
	t.Run("DeleteNet", func(t *testing.T) {
		n, err := srv.CreateNet("test")
		if err != nil {
			t.Error(err)
		}
		n, err = srv.DeleteNet(n.ID)
		if err != nil {
			t.Error(err)
		}
		if n.Name != "test" {
			t.Error("Name not set")
		}
		if n.ID == "" {
			t.Error("ID not set")
		}
	})
	t.Run("Nets", func(t *testing.T) {

	})
	t.Run("Net", func(t *testing.T) {

	})
}
