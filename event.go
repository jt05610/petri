package petri

var _ Object = (*EventSchema)(nil)
var _ Input = (*EventInput)(nil)
var _ Update = (*EventUpdate)(nil)
var _ Filter = (*EventFilter)(nil)

type EventSchema struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Url          string      `json:"url"`
	InputSchema  TokenSchema `json:"inputSchema"`
	OutputSchema TokenSchema `json:"outputSchema"`
}

func (e *EventSchema) PostInit() error {
	return nil
}

func (e *EventSchema) Kind() Kind {
	return EventObject
}

func (e *EventSchema) Identifier() string {
	return e.Name
}

func (e *EventSchema) String() string {
	return e.Name
}

func (e *EventSchema) Update(update Update) error {
	up, ok := update.(*EventUpdate)
	if !ok {
		return ErrWrongUpdate
	}
	if up.Input != nil {
		if up.Mask.Name {
			e.Name = up.Input.Name
		}
		if up.Mask.Url {
			e.Url = up.Input.Url
		}
		if up.Mask.InputSchema {
			e.InputSchema = up.Input.InputSchema
		}
		if up.Mask.OutputSchema {
			e.OutputSchema = up.Input.OutputSchema
		}
	}
	return nil
}

func (e *EventSchema) Document() Document {
	return Document{
		"_id":    e.ID,
		"name":   e.Name,
		"url":    e.Url,
		"input":  e.InputSchema,
		"output": e.OutputSchema,
	}
}

type EventInput struct {
	Name         string      `json:"name"`
	Url          string      `json:"url"`
	InputSchema  TokenSchema `json:"inputSchema"`
	OutputSchema TokenSchema `json:"outputSchema"`
}

func (e EventInput) Kind() Kind {
	return EventObject
}

func (e EventInput) Object() Object {
	return &EventSchema{
		ID:           ID(),
		Name:         e.Name,
		Url:          e.Url,
		InputSchema:  e.InputSchema,
		OutputSchema: e.OutputSchema,
	}
}

type EventMask struct {
	Name         bool
	Url          bool
	InputSchema  bool
	OutputSchema bool
}

type EventUpdate struct {
	Input *EventInput
	Mask  *EventMask
}

type EventFilter struct {
	Name         *StringSelector
	Url          *StringSelector
	InputSchema  *StringSelector
	OutputSchema *StringSelector
}

func (e EventFilter) IsFilter() {}
func (e EventUpdate) IsUpdate() {}
func (e EventInput) IsInput()   {}
