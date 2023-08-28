package grbl

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"unicode"
)

type Position struct {
	X float64
	Y float64
	Z float64
}

type StatusUpdate interface {
	IsStatusUpdate()
}

type State string

const (
	Alarm State = "Alarm"
	Idle  State = "Idle"
	Run   State = "Run"
)

type Active struct {
	Spindle bool
}

type Status struct {
	Alarm           State
	Active          *Active
	MachinePosition *Position
	Feed            float64
	WorkPosition    *Position
	Override        *struct {
		X float64
		Y float64
		Z float64
	}
}

func (s *Status) IsStatusUpdate() {}

type Ack struct {
}

func (a *Ack) String() string {
	return "ok"
}

func (a *Ack) IsStatusUpdate() {}

type StatusType string

type Token int

const (
	Newline Token = iota
	OpenAngle
	CloseAngle
	Colon
	Comma
	Bar
	Identifier
	Float
)

var tokens = []string{
	Newline:    "NEWLINE",
	OpenAngle:  "<",
	CloseAngle: ">",
	Colon:      ":",
	Comma:      ",",
	Bar:        "|",
	Identifier: "IDENT",
	Float:      "FLOAT",
}

func (t Token) String() string {
	return tokens[t]
}

type Lexer struct {
	pos int
	rdr *bufio.Reader
}

func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		rdr: bufio.NewReader(r),
		pos: 0,
	}
}
func (l *Lexer) Lex() (int, Token, string) {
	for {
		l.pos++
		r, _, err := l.rdr.ReadRune()
		if err != nil {
			return l.pos, Newline, Newline.String()
		}
		switch r {
		case '\n':
			return l.pos, Newline, Newline.String()
		case '<':
			return l.pos, OpenAngle, OpenAngle.String()
		case '>':
			return l.pos, CloseAngle, CloseAngle.String()
		case ':':
			return l.pos, Colon, Colon.String()
		case ',':
			return l.pos, Comma, Comma.String()
		case '|':
			return l.pos, Bar, Bar.String()
		default:
			startPos := l.pos
			if l.isFloatPart(r) {
				l.backup()
				return startPos, Float, l.lexFloat()
			}
			if unicode.IsLetter(r) {
				l.backup()
				return startPos, Identifier, l.lexString()
			}
		}
	}
}

func (l *Lexer) lexInt() string {
	var lit string
	for {
		r, _, err := l.rdr.ReadRune()
		if err != nil {
			return lit
		}
		if unicode.IsDigit(r) {
			lit += string(r)
		} else {
			l.backup()
			return lit
		}
	}
}

func (l *Lexer) isFloatPart(r rune) bool {
	return unicode.IsDigit(r) || r == '.' || r == '-'
}

func (l *Lexer) lexFloat() string {
	var lit string
	for {
		r, _, err := l.rdr.ReadRune()
		if err != nil {
			return lit
		}
		if l.isFloatPart(r) {
			lit += string(r)
		} else {
			l.backup()
			return lit
		}
	}
}

func (l *Lexer) lexString() string {
	var lit string
	for {
		r, _, err := l.rdr.ReadRune()
		if err != nil {
			return lit
		}
		if unicode.IsLetter(r) {
			lit += string(r)
		} else {
			l.backup()
			return lit
		}
	}
}

func (l *Lexer) backup() {
	l.pos--
	err := l.rdr.UnreadRune()
	if err != nil {
		panic(err)
	}
}

type Parser struct {
	lexer *Lexer
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		lexer: NewLexer(r),
	}
}

func (p *Parser) errorf(pos int, format string, args ...interface{}) error {
	return fmt.Errorf("%d: %s", pos, fmt.Sprintf(format, args...))
}

func (p *Parser) parseFloat() float64 {
	pos, tok, lit := p.lexer.Lex()
	switch tok {
	case Newline:
		panic(p.errorf(pos, "unexpected newline"))
	case Colon, Comma:
		return p.parseFloat()
	}
	if tok != Float {
		panic(p.errorf(pos, "expected float, got %q", lit))
	}
	f, err := strconv.ParseFloat(lit, 64)
	if err != nil {
		panic(p.errorf(pos, "expected float, got %q", lit))
	}
	return f

}

func (p *Parser) parsePosition() *Position {
	return &Position{
		X: p.parseFloat(),
		Y: p.parseFloat(),
		Z: p.parseFloat(),
	}
}

func (p *Parser) parseOverride() *struct {
	X float64
	Y float64
	Z float64
} {
	return &struct {
		X float64
		Y float64
		Z float64
	}{
		X: p.parseFloat(),
		Y: p.parseFloat(),
		Z: p.parseFloat(),
	}
}

func (p *Parser) parseString() string {
	pos, tok, lit := p.lexer.Lex()
	switch tok {
	case Newline:
		panic(p.errorf(pos, "unexpected newline"))
	case Colon, Comma:
		return p.parseString()
	}
	if tok != Identifier {
		panic(p.errorf(pos, "expected identifier, got %q", lit))
	}
	return lit
}
func (p *Parser) parseActive() *Active {
	s := p.parseString()
	ret := &Active{}
	if s == "S" {
		ret.Spindle = true
	}
	return ret

}

func (p *Parser) parseStatusUpdate(s *Status) (*Status, error) {
	for {
		pos, tok, lit := p.lexer.Lex()
		switch tok {
		case Identifier:
			switch lit {
			case "Alarm", "Idle", "Run":
				s.Alarm = State(lit)
			case "A":
				s.Active = p.parseActive()
			case "MPos":
				s.MachinePosition = p.parsePosition()
			case "F":
				s.Feed = p.parseFloat()
			case "WCO":
				s.WorkPosition = p.parsePosition()
			case "Ov":
				s.Override = p.parseOverride()
			default:
				return nil, p.errorf(pos, "unknown identifier %q", lit)
			}
		case Bar, Comma, Colon:
			continue
		case CloseAngle:
			return s, nil
		}

	}
}

var IllegalIdentifier = fmt.Errorf("illegal identifier")

func (p *Parser) Parse() (StatusUpdate, error) {
	status := &Status{}
	for {
		pos, tok, lit := p.lexer.Lex()
		switch tok {
		case Newline:
			return status, nil
		case OpenAngle:
			return p.parseStatusUpdate(status)
		case Identifier:
			switch lit {
			case "ok":
				return &Ack{}, nil
			default:
				return nil, IllegalIdentifier
			}
		default:
			return nil, p.errorf(pos, "expected identifier, got %q", lit)
		}
	}
}
