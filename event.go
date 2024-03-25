package petri

type EventSchema struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Url          string       `json:"url"`
	InputSchema  *TokenSchema `json:"inputSchema"`
	OutputSchema *TokenSchema `json:"outputSchema"`
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
