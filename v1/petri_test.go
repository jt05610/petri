package petri

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
	petri2 "github.com/jt05610/petri/proto/v1"
	learner "github.com/jt05610/petri/v1/testLearner"
	model "github.com/jt05610/petri/v1/testModel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"sync"
	"testing"
	"time"
)

type learnerService struct {
	learner.UnimplementedLearnerServiceServer
	store *model.Store
}

func resultToData(r *model.Result) []*model.Datum {
	prodMeasurements := map[string][]*model.Measurement{}
	for _, m := range r.Measurements {
		if _, ok := prodMeasurements[*m.Product.Id]; !ok {
			prodMeasurements[*m.Product.Id] = []*model.Measurement{}
		}
		prodMeasurements[*m.Product.Id] = append(prodMeasurements[*m.Product.Id], m)
	}
	data := make([]*model.Datum, len(prodMeasurements))
	i := 0
	for _, measurements := range prodMeasurements {
		data[i] = &model.Datum{
			Product:      measurements[0].Product,
			Measurements: measurements,
		}
		i++
	}
	return data
}

func (l *learnerService) Complete(_ context.Context, request *learner.CompleteRequest) (*learner.CompleteResponse, error) {
	resData := resultToData(request.Learning)
	l.store.Data = append(l.store.Data, resData...)
	return &learner.CompleteResponse{
		Learned: l.store,
		Updated: &model.Store{Data: resData},
	}, nil
}

func (l *learnerService) Learn(_ context.Context, request *learner.LearnRequest) (*learner.LearnResponse, error) {
	l.store = request.Learned
	return &learner.LearnResponse{
		Learning: request.Result,
	}, nil
}

var _ learner.LearnerServiceServer = (*learnerService)(nil)

func fakeResult(p *model.Product) *model.Result {
	return &model.Result{
		Measurements: fakeMeasurements(p),
	}
}

func fakeSample() *model.Sample {
	fields := "abc"
	name := fmt.Sprintf("testSample_%s", petri.ID())
	values := make([]*model.Quantity, 3)
	for i := 0; i < 3; i++ {
		values[i] = fakeQuantity(&fields)
	}

	return &model.Sample{
		Fields: &fields,
		Values: values,
		Name:   &name,
	}
}

func fakeProduct() *model.Product {
	m := "volume"
	id := petri.ID()
	loc := "A1"
	return &model.Product{
		Amount:   fakeQuantity(&m),
		Sample:   fakeSample(),
		Id:       &id,
		Location: &loc,
	}
}

func fakeMethod() *string {
	methods := []string{
		"size",
		"pdi",
		"zeta",
	}
	randIn := rand.Int31n(int32(len(methods)))
	return &methods[randIn]
}

func randomFloat() *float32 {
	val := rand.Float32()
	return &val
}

func methodUnit(method *string) *string {
	methodMap := map[string]string{
		"size": "nm",
		"pdi":  "",
		"zeta": "mV",
	}
	if val, ok := methodMap[*method]; ok {
		if val == "" {
			return nil
		}
		return &val
	}
	return nil
}

func fakeQuantity(method *string) *model.Quantity {
	switch *method {
	case "size":
		return &model.Quantity{
			Value: randomFloat(),
			Unit:  methodUnit(method),
		}
	case "pdi":
		return &model.Quantity{
			Value: randomFloat(),
			Unit:  nil,
		}
	case "zeta":
		return &model.Quantity{
			Value: randomFloat(),
			Unit:  methodUnit(method),
		}
	case "volume":
		unit := "mL"
		return &model.Quantity{
			Value: randomFloat(),
			Unit:  &unit,
		}
	default:
		return &model.Quantity{
			Value: randomFloat(),
		}
	}
}

func fakeMeasurements(p *model.Product) []*model.Measurement {
	n := rand.Int31n(10)
	ret := make([]*model.Measurement, n)
	for i := 0; i < int(n); i++ {
		method := fakeMethod()
		ret[i] = &model.Measurement{
			Method:  fakeMethod(),
			Product: p,
			Value:   fakeQuantity(method),
		}
	}
	return ret
}

func fakeData() []*model.Datum {
	n := rand.Int31n(10)
	ret := make([]*model.Datum, n)
	for i := 0; i < int(n); i++ {
		prod := fakeProduct()
		ret[i] = &model.Datum{
			Product:      prod,
			Measurements: fakeMeasurements(prod),
		}
	}
	return ret
}

func fakeLearned() *model.Store {
	return &model.Store{
		Data: fakeData(),
	}
}

func fakeLearnRequest() *learner.LearnRequest {
	return &learner.LearnRequest{
		Result:  fakeResult(fakeProduct()),
		Learned: fakeLearned(),
	}
}

func TestTokenConverter_Marshal(t *testing.T) {
	tc := NewTokenConverter[*learner.LearnRequest]()
	if len(tc.Fields()) == 0 {
		t.Fatal("failed to get fields")
	}
	req := fakeLearnRequest()
	initialBytes, err := proto.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	tm, err := tc.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(tm) != len(tc.Fields()) {
		t.Fatalf("expected %d fields, got %d", len(tc.Fields()), len(tm))
	}
	converted, err := tc.Unmarshal(tm)
	if err != nil {
		t.Fatal(err)
	}
	convertedBytes, err := proto.Marshal(converted)
	if err != nil {
		t.Fatal(err)
	}
	if len(initialBytes) != len(convertedBytes) {
		t.Fatalf("expected %d bytes, got %d", len(initialBytes), len(convertedBytes))
	}
	for i, b := range initialBytes {
		if convertedBytes[i] != b {
			t.Fatalf("expected byte %d to be %d, got %d", i, b, convertedBytes[i])
		}
	}
}

func TestHandler_Handle(t *testing.T) {
	srv := learnerService{}
	h := NewHandler(srv.Learn)
	ctx := context.Background()
	in := fakeLearnRequest()
	expected := &learner.LearnResponse{
		Learning: in.Result,
	}
	tm, err := h.InCvt.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	out, err := h.Handle(ctx, tm)
	if err != nil {
		t.Fatal(err)
	}
	converted, err := h.OutCvt.Unmarshal(out)
	if err != nil {
		t.Fatal(err)
	}
	bb1, err := proto.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	bb2, err := proto.Marshal(converted)
	if err != nil {
		t.Fatal(err)
	}
	if len(bb1) != len(bb2) {
		t.Fatalf("expected %d bytes, got %d", len(bb1), len(bb2))
	}
	for i, b := range bb1 {
		if bb2[i] != b {
			t.Fatalf("expected byte %d to be %d, got %d", i, b, bb2[i])
		}
	}

	complete := &learner.CompleteRequest{
		Learning: converted.Learning,
	}
	cHandle := NewHandler(srv.Complete)
	newData := resultToData(converted.Learning)
	data := in.Learned.Data
	allData := make([]*model.Datum, len(data)+len(newData))
	copy(allData, data)
	copy(allData[len(data):], newData)
	cExpected := &learner.CompleteResponse{
		Learned: &model.Store{
			Data: allData,
		},
		Updated: &model.Store{
			Data: newData,
		},
	}
	cTm, err := cHandle.InCvt.Marshal(complete)
	if err != nil {
		t.Fatal(err)
	}
	cOut, err := cHandle.Handle(ctx, cTm)
	if err != nil {
		t.Fatal(err)
	}
	cConverted, err := cHandle.OutCvt.Unmarshal(cOut)
	if err != nil {
		t.Fatal(err)
	}
	cbb1, err := proto.Marshal(cExpected)
	if err != nil {
		t.Fatal(err)
	}
	cbb2, err := proto.Marshal(cConverted)
	if err != nil {
		t.Fatal(err)
	}
	if len(cbb1) != len(cbb2) {
		t.Fatalf("expected %d bytes, got %d", len(cbb1), len(cbb2))
	}
}

func TestRegisterTransitionHandler(t *testing.T) {
	n := LoadNet("../examples/dbtl/petri/learner.yaml")
	srv := learnerService{}
	h := NewHandler(srv.Learn)
	err := RegisterTransitionHandler(n, "learn", h)
	if err != nil {
		t.Fatal(err)
	}
	h2 := NewHandler(srv.Complete)
	err = RegisterTransitionHandler(n, "complete", h2)
	if err != nil {
		t.Fatal(err)
	}
	req := fakeLearnRequest()
	tm, err := h.InCvt.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	m := petri.Marking{
		"result":   []petri.Token{tm["result"]},
		"learned":  []petri.Token{tm["learned"]},
		"learning": []petri.Token{},
		"updated":  []petri.Token{},
	}
	updated, err := n.Process(context.Background(), m)
	if err != nil {
		t.Fatal(err)
	}
	newData := resultToData(req.Result)
	data := req.Learned.Data
	allData := make([]*model.Datum, len(data)+len(newData))
	copy(allData, data)
	copy(allData[len(data):], newData)
	cExpected := &learner.LearnResponse{
		Learning: req.Result,
	}

	tm, err = h.OutCvt.Marshal(cExpected)
	if err != nil {
		t.Fatal(err)

	}

	expected := petri.Marking{
		"result":   []petri.Token{},
		"learned":  []petri.Token{},
		"learning": []petri.Token{tm["learning"]},
		"updated":  []petri.Token{},
	}

	if !updated.Equals(expected) {
		t.Fatal("expected updated to equal expected")
	}
}

type learnerClient struct {
	learner.LearnerServiceClient
	RPCClient
}

func connectClient(conn grpc.ClientConnInterface) (*learnerClient, error) {
	return &learnerClient{
		LearnerServiceClient: learner.NewLearnerServiceClient(conn),
		RPCClient:            NewClient(conn),
	}, nil
}

func TestServer(t *testing.T) {
	n := LoadNet("../examples/dbtl/petri/learner.yaml")
	srv := &learnerService{}
	opt := &ServerOptions[learner.LearnerServiceServer]{
		Host:   "localhost:0",
		Server: srv,
		Net:    n,
		Reg:    learner.RegisterLearnerServiceServer,
		Handlers: HandlerMap{
			"learn":    NewHandler(srv.Learn),
			"complete": NewHandler(srv.Complete),
		},
	}
	service, err := NewServer(opt)
	if err != nil {
		t.Fatal(err)
	}
	addr := service.Addr().String()
	fmt.Printf("server listening on %s\n", addr)
	var wg sync.WaitGroup
	wg.Add(1)
	ctx, can := context.WithCancel(context.Background())
	defer can()
	go func() {
		defer wg.Done()
		err := service.Serve(ctx)
		if err != nil {
			t.Error(err)
		}
	}()
	client, err := Dial[*learnerClient](ctx, &ClientOptions[*learnerClient]{
		Addr:   addr,
		Attach: connectClient,
		RPCOptions: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := fakeLearnRequest()
	bb, err := proto.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	res, err := client.PutToken(ctx, &petri2.PutTokenRequest{
		PlaceId: "result",
		Token: &petri2.Token{
			Id:   petri.ID(),
			Data: bb,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatal("expected success")
	}
	tr := true
	m, err := client.GetMarking(ctx, &petri2.GetMarkingRequest{WithValue: &tr})
	if err != nil {
		t.Fatal(err)
	}
	learnedBytes, err := proto.Marshal(req.Learned)
	if err != nil {
		t.Fatal(err)
	}
	res, err = client.PutToken(ctx, &petri2.PutTokenRequest{
		PlaceId: "learned",
		Token: &petri2.Token{
			Id:   petri.ID(),
			Data: learnedBytes,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatal("expected success")
	}
	m, err = client.GetMarking(ctx, &petri2.GetMarkingRequest{WithValue: &tr})
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Marking.PlaceMarkings) != 2 {
		t.Fatalf("expected 2 place markings, got %d", len(m.Marking.PlaceMarkings))
	}
	fmt.Println(m)
	time.Sleep(1 * time.Second)
	m, err = client.GetMarking(ctx, &petri2.GetMarkingRequest{WithValue: &tr})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(m)
	service.Close()
	wg.Wait()
}
