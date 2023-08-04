package fixtures

import (
	"github.com/jt05610/petri"
	"testing"
)

type Errors struct {
	Getter  error
	Lister  error
	Adder   error
	Remover error
	Updater error
}

type ServiceTestCase struct {
	Name string
	Srv  petri.Service
	petri.Input
	Expect petri.Object
	petri.Update
	petri.Filter
	Errors Errors
}

func RunServiceTest(t *testing.T, tc *ServiceTestCase) {
	items := make([]petri.Object, 2)
	// Add 2 copies of the input
	t.Run(tc.Name+".Adder", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			items[i] = RunAdderTest(t, tc)
		}
	})

	// check that the items were added
	t.Run(tc.Name+".Lister", func(t *testing.T) {
		RunListerTest(t, tc, items)
	})
	// check that the items can be retrieved
	t.Run(tc.Name+".Getter", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			RunGetterTest(t, tc, items[i].Identifier())
		}
	})
	// check that the items can be updated
	t.Run(tc.Name+".Updater", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			RunUpdaterTest(t, tc, items[i].Identifier())
		}
	})
	// remove first item
	t.Run(tc.Name+".Remover", func(t *testing.T) {
		RunRemoverTest(t, tc, items[0].Identifier())
	})
	// make sure we cant get the first item after it was removed
	t.Run(tc.Name+".Getter", func(t *testing.T) {
		RunGetterTest(t, tc, items[0].Identifier())
	})
	// make sure it doesn't appear in the list
	t.Run(tc.Name+".Lister", func(t *testing.T) {
		RunListerTest(t, tc, items[1:])
	})
}

func RunAdderTest(t *testing.T, tc *ServiceTestCase) petri.Object {
	actual, err := tc.Srv.Add(tc.Input)
	if err.Error() != tc.Errors.Adder.Error() {
		t.Fatalf("Adder test failed: %v", err)
	}
	if actual.Kind() != tc.Expect.Kind() {
		t.Fatalf("Adder kind test failed: expected %v, got %v", tc.Expect.Kind(), actual.Kind())
	}
	if actual.Identifier() == "" {
		t.Fatalf("Adder ID test failed: expected non-empty ID, got %v", actual.Identifier())
	}
	return actual
}

func RunListerTest(t *testing.T, tc *ServiceTestCase, expect []petri.Object) {
	actual, err := tc.Srv.List(tc.Filter)
	if err.Error() != tc.Errors.Lister.Error() {
		t.Fatalf("Lister test failed: %v", err)
	}
	if len(actual) != 2 {
		t.Fatalf("Lister length test failed: expected 2, got %v", len(actual))
	}
	for i := 0; i < 2; i++ {
		if actual[i].Identifier() != expect[i].Identifier() {
			t.Fatalf("Lister ID test failed: expected %v, got %v", expect[i].Identifier(), actual[i].Identifier())
		}
		if actual[i].Kind() != expect[i].Kind() {
			t.Fatalf("Lister kind test failed: expected %v, got %v", expect[i].Kind(), actual[i].Kind())
		}
	}
}

func RunGetterTest(t *testing.T, tc *ServiceTestCase, id string) {
	actual, err := tc.Srv.Get(id)
	if err.Error() != tc.Errors.Getter.Error() {
		t.Fatalf("Getter test failed: %v", err)
	}
	if actual.Identifier() != id {
		t.Fatalf("Getter ID test failed: expected %v, got %v", id, actual.Identifier())
	}
	if actual.Kind() != tc.Expect.Kind() {
		t.Fatalf("Getter kind test failed: expected %v, got %v", tc.Expect.Kind(), actual.Kind())
	}
}

func RunUpdaterTest(t *testing.T, tc *ServiceTestCase, id string) {
	actual, err := tc.Srv.Update(id, tc.Update)
	if err.Error() != tc.Errors.Updater.Error() {
		t.Fatalf("Updater test failed: %v", err)
	}
	if actual.Identifier() != id {
		t.Fatalf("Updater ID test failed: expected %v, got %v", id, actual.Identifier())
	}
	if actual.Kind() != tc.Expect.Kind() {
		t.Fatalf("Updater kind test failed: expected %v, got %v", tc.Expect.Kind(), actual.Kind())
	}
}

func RunRemoverTest(t *testing.T, tc *ServiceTestCase, id string) {
	actual, err := tc.Srv.Remove(id)
	if err.Error() != tc.Errors.Remover.Error() {
		t.Fatalf("Remover test failed: %v", err)
	}
	if actual.Identifier() != id {
		t.Fatalf("Remover ID test failed: expected %v, got %v", id, actual.Identifier())
	}
	if actual.Kind() != tc.Expect.Kind() {
		t.Fatalf("Remover kind test failed: expected %v, got %v", tc.Expect.Kind(), actual.Kind())
	}
}
