package prisma

import (
	"context"
	"errors"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/device"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/marked"
	"github.com/jt05610/petri/prisma/db"
	"strings"
)

var _ device.Service = (*DeviceClient)(nil)

type DeviceClient struct {
	*db.PrismaClient
	Seen map[string]*db.DeviceModel
}

func toSnakeCase(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "_")
}

func makePetriNet(model *db.NetModel, handlers control.Handlers) *labeled.Net {
	tt := make([]*petri.Transition, len(model.Transitions()))
	pp := make([]*petri.Place, len(model.Places()))
	nodeIndex := make(map[string]petri.Node)
	aa := make([]*petri.Arc, len(model.Arcs()))
	ee := make([]*labeled.Event, 0)
	// eventIndex maps the events name to the index of the transition in the tt slice
	eventIndex := make(map[string]int)
	initialMarking := make(map[string]int)

	for i, transition := range model.Transitions() {
		tt[i] = &petri.Transition{
			ID:   transition.ID,
			Name: transition.Name,
		}
		nodeIndex[transition.ID] = tt[i]
		if transition.Events() != nil {
			for _, event := range transition.Events() {
				fields := make([]*labeled.Field, len(event.Fields()))
				for i, field := range event.Fields() {
					fields[i] = &labeled.Field{
						Name: field.Name,
						Type: labeled.FieldType(field.Type),
					}
				}
				ee = append(ee, &labeled.Event{
					Name:   event.Name,
					Fields: fields,
				})
				eventIndex[event.Name] = i
			}
		}
	}

	for i, place := range model.Places() {
		pp[i] = &petri.Place{
			ID:    place.ID,
			Name:  place.Name,
			Bound: place.Bound,
		}
		nodeIndex[place.ID] = pp[i]

		if model.InitialMarking != nil {
			if len(model.InitialMarking) > i {
				initialMarking[place.ID] = model.InitialMarking[i]
			}
		} else {
			initialMarking[place.ID] = 0
		}
	}

	for i, arc := range model.Arcs() {
		if arc.FromPlace {
			aa[i] = &petri.Arc{
				ID:   arc.ID,
				Src:  nodeIndex[arc.PlaceID],
				Dest: nodeIndex[arc.TransitionID],
			}
		} else {
			aa[i] = &petri.Arc{
				ID:   arc.ID,
				Src:  nodeIndex[arc.TransitionID],
				Dest: nodeIndex[arc.PlaceID],
			}
		}
	}
	pNet := petri.New(pp, tt, aa)
	mNet := marked.NewFromMap(pNet, initialMarking)
	lNet := labeled.New(mNet)

	for _, event := range ee {
		var err error
		if handlers == nil {
			err = lNet.AddEventHandler(event, tt[eventIndex[event.Name]], func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
				return data, nil
			})
		} else {
			err = lNet.AddEventHandler(event, tt[eventIndex[event.Name]], handlers[event.Name])
		}
		if err != nil {
			panic(err)
		}
	}
	return lNet
}

func (d *DeviceClient) List(ctx context.Context) ([]*device.ListItem, error) {
	devices, err := d.Device.FindMany().Exec(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*device.ListItem, len(devices))
	for i, dev := range devices {
		list[i] = &device.ListItem{
			ID:   dev.ID,
			Name: dev.Name,
		}
	}
	return list, nil
}

func ConvertDevice(dev *db.DeviceModel, handlers control.Handlers) (*device.Device, error) {
	nets := make([]*labeled.Net, len(dev.Nets()))
	for i, net := range dev.Nets() {
		nets[i] = makePetriNet(net.Net(), handlers)
	}
	return device.New(dev.ID, dev.Name, nets), nil
}

func (d *DeviceClient) Load(ctx context.Context, id string, handlers control.Handlers) (*device.Device, error) {
	dev, err := d.Device.FindUnique(
		db.Device.ID.Equals(id),
	).With(
		db.Device.Nets.Fetch().With(
			db.DevicesOnNets.Net.Fetch().With(
				db.Net.Places.Fetch(),
				db.Net.Arcs.Fetch(),
				db.Net.Transitions.Fetch().With(
					db.Transition.Events.Fetch().With(
						db.Event.Fields.Fetch(),
					),
				),
			),
		),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}
	lD, err := ConvertDevice(dev, handlers)
	if err != nil {
		return nil, err
	}
	if d.Seen == nil {
		d.Seen = make(map[string]*db.DeviceModel)
	}
	d.Seen[dev.ID] = dev
	return lD, nil
}

func requireAuthorID(ctx context.Context) (string, error) {
	authorID, ok := ctx.Value("authorID").(string)
	if !ok {
		return "", errors.New("authorID not found in context")
	}
	return authorID, nil
}

func (d *DeviceClient) Flush(ctx context.Context, dev *device.Device) (string, error) {
	if dev.Instance == nil {
		return "", errors.New("device instance is nil")
	}
	authorID, err := requireAuthorID(ctx)
	if err != nil {
		return "", err
	}
	i, err := d.Instance.CreateOne(
		db.Instance.Author.Link(
			db.User.ID.Equals(authorID),
		),
		db.Instance.Language.Set(db.Language(strings.ToUpper(dev.Instance.Language))),
		db.Instance.Name.Set(dev.Name),
		db.Instance.Device.Link(
			db.Device.ID.Equals(dev.ID),
		),
		db.Instance.Addr.Set(dev.Instance.Addr()),
	).Exec(ctx)

	return i.ID, nil
}
