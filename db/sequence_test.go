package prisma_test

import (
	"context"
	"github.com/jt05610/petri/db"
	"github.com/jt05610/petri/prisma/db"
	"testing"
)

func GetMock() (*prisma.RunClient, *db.Mock, func(t *testing.T)) {
	client, mock, ensure := db.NewMock()
	return &prisma.RunClient{PrismaClient: client}, mock, ensure
}

func TestRunClient_Load(t *testing.T) {
	client, mock, ensure := GetMock()
	defer ensure(t)

	expected := &db.RunModel{
		InnerRun: db.InnerRun{
			ID:          "123",
			Name:        "",
			Description: "",
			CreatedAt:   db.DateTime{},
			UpdatedAt:   db.DateTime{},
			NetID:       "",
		},
		RelationsRun: db.RelationsRun{
			Net: &db.NetModel{
				InnerNet: db.InnerNet{
					ID:        "",
					Name:      "",
					CreatedAt: db.DateTime{},
					UpdatedAt: db.DateTime{},
				},
				RelationsNet: db.RelationsNet{
					Places: []db.PlaceModel{
						{
							InnerPlace: db.InnerPlace{
								ID:          "",
								Name:        "",
								Description: nil,
								CreatedAt:   db.DateTime{},
								UpdatedAt:   db.DateTime{},
								Bound:       0,
							},
						},
					},
					Transitions: []db.TransitionModel{
						{
							InnerTransition: db.InnerTransition{
								ID:          "",
								Condition:   nil,
								Description: nil,
								Name:        "",
								CreatedAt:   db.DateTime{},
								UpdatedAt:   db.DateTime{},
							},
						},
					},
					Arcs: []db.ArcModel{
						{
							InnerArc: db.InnerArc{
								ID:           "",
								NetID:        "",
								FromPlace:    false,
								PlaceID:      "",
								TransitionID: "",
								CreatedAt:    db.DateTime{},
								UpdatedAt:    db.DateTime{},
							},
						},
					},
					Children: []db.NetModel{}},
			},
			Steps: []db.StepModel{
				{
					InnerStep: db.InnerStep{
						ID:       "",
						Order:    0,
						RunID:    "",
						ActionID: "",
					},
					RelationsStep: db.RelationsStep{
						Action: &db.ActionModel{
							InnerAction: db.InnerAction{
								ID:        "",
								Input:     nil,
								Output:    nil,
								DeviceID:  "",
								EventID:   "",
								CreatedAt: db.DateTime{},
								UpdatedAt: db.DateTime{},
							},
							RelationsAction: db.RelationsAction{
								Constants: []db.ConstantModel{
									{
										InnerConstant: db.InnerConstant{
											ID:       "",
											ActionID: "",
											FieldID:  "",
										},
										RelationsConstant: db.RelationsConstant{
											Field: &db.FieldModel{
												InnerField: db.InnerField{
													ID:        "",
													Name:      "",
													Type:      "",
													Condition: nil,
													EventID:   "",
													CreatedAt: db.DateTime{},
													UpdatedAt: db.DateTime{},
												},
											},
										},
									},
								},
								Device: &db.DeviceModel{
									InnerDevice: db.InnerDevice{
										ID:          "",
										AuthorID:    "",
										Name:        "",
										Description: "",
										CreatedAt:   db.DateTime{},
										UpdatedAt:   db.DateTime{},
									},
									RelationsDevice: db.RelationsDevice{
										Nets: []db.NetModel{
											{
												InnerNet: db.InnerNet{
													ID:             "",
													Name:           "",
													Description:    "",
													InitialMarking: nil,
													CreatedAt:      db.DateTime{},
													UpdatedAt:      db.DateTime{},
													AuthorID:       "",
													ParentID:       nil,
												},
											},
										},
										Instances: []db.InstanceModel{
											{
												InnerInstance: db.InnerInstance{
													ID:        "",
													CreatedAt: db.DateTime{},
													UpdatedAt: db.DateTime{},
													AuthorID:  "",
													Language:  "",
													Name:      "",
													DeviceID:  "",
													Addr:      "",
												},
											},
										},
									},
								},
								Event: &db.EventModel{
									InnerEvent: db.InnerEvent{
										ID:                    "",
										Name:                  "",
										Description:           nil,
										CreatedAt:             db.DateTime{},
										UpdatedAt:             db.DateTime{},
										PlaceInterfaceID:      nil,
										TransitionInterfaceID: nil,
									},
									RelationsEvent: db.RelationsEvent{
										Fields: []db.FieldModel{
											{
												InnerField: db.InnerField{
													ID:        "",
													Name:      "",
													Type:      "",
													Condition: nil,
													EventID:   "",
													CreatedAt: db.DateTime{},
													UpdatedAt: db.DateTime{},
												},
											},
										},
										Transitions: []db.TransitionModel{
											{
												InnerTransition: db.InnerTransition{
													ID:          "",
													Condition:   nil,
													Description: nil,
													Name:        "",
													CreatedAt:   db.DateTime{},
													UpdatedAt:   db.DateTime{},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	mock.Run.Expect(
		client.Run.FindUnique(
			db.Run.ID.Equals("123"),
		).With(
			db.Run.Net.Fetch().With(
				db.Net.Places.Fetch(),
			).With(
				db.Net.Transitions.Fetch(),
			).With(
				db.Net.Arcs.Fetch(),
			).With(
				db.Net.Children.Fetch(),
			),
		).With(
			db.Run.Steps.Fetch().With(
				db.Step.Action.Fetch().With(
					db.Action.Constants.Fetch().With(
						db.Constant.Field.Fetch().With(),
					),
				).With(
					db.Action.Device.Fetch().With(
						db.Device.Nets.Fetch(),
					).With(
						db.Device.Instances.Fetch(),
					).With(),
				).With(
					db.Action.Event.Fetch().With(
						db.Event.Fields.Fetch().With(),
					).With(
						db.Event.Transitions.Fetch(),
					),
				),
			).OrderBy(
				db.Step.Order.Order(db.SortOrderAsc),
			),
		),
	).Returns(*expected)

	run, err := client.Load(context.Background(), "123")
	if err != nil {
		t.Fatal(err)
	}
	Check(t, run.ID, "123")
}

func Check[T comparable](t *testing.T, actual T, expected T) {
	if actual != expected {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}
