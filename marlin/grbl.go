package grbl

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type Position struct {
	X float32
	Y float32
	Z float32
}

type StatusUpdate interface {
	IsStatusUpdate()
}

type Active struct {
	Spindle bool
	Flood   bool
	Mist    bool
}

type Override struct {
	Rapid   byte
	Feed    byte
	Spindle byte
}

type LimitPins struct {
	X bool
	Y bool
	Z bool
}

type Status struct {
	State           string
	Error           *int
	Alarm           *int
	Active          *Active
	MachinePosition *Position
	Feed            float32
	WorkPosition    *Position
	Override        *Override
	LimitPins       *LimitPins
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
	Return Token = iota
	Newline
	OpenAngle
	CloseAngle
	Colon
	Comma
	Bar
	Identifier
	Float
)

var tokens = []string{
	Return:     "RETURN",
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
		case '\r':
			return l.pos, Return, Return.String()
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

func (p *Parser) parseFloat() float32 {
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
	f, err := strconv.ParseFloat(lit, 32)
	if err != nil {
		panic(p.errorf(pos, "expected float, got %q", lit))
	}
	return float32(f)

}

func (p *Parser) parsePosition() *Position {
	return &Position{
		X: p.parseFloat(),
		Y: p.parseFloat(),
		Z: p.parseFloat(),
	}
}

func (p *Parser) parseByte() uint8 {
	pos, tok, lit := p.lexer.Lex()
	switch tok {
	case Newline:
		panic(p.errorf(pos, "unexpected newline"))
	case Colon, Comma:
		return p.parseByte()
	}
	if tok != Float {
		panic(p.errorf(pos, "expected float, got %q", lit))
	}
	f, err := strconv.ParseUint(lit, 10, 8)
	if err != nil {
		panic(p.errorf(pos, "expected float, got %q", lit))
	}
	return uint8(f)
}

func (p *Parser) parseOverride() *Override {
	ret := &Override{}
	idx := map[int]*uint8{
		0: &ret.Rapid,
		1: &ret.Feed,
		2: &ret.Spindle,
	}
	for i := 0; i < 3; i++ {
		pos, tok, lit := p.lexer.Lex()
		switch tok {
		case Colon, Comma:
			continue
		case Float:
			f, err := strconv.ParseUint(lit, 10, 8)
			if err != nil {
				panic(p.errorf(pos, "expected float, got %q", lit))
			}
			*idx[i] = uint8(f)
		default:
			panic(p.errorf(pos, "expected float, got %q", lit))
		}

	}
	return ret
}

func (p *Parser) parseString() string {
	pos, tok, lit := p.lexer.Lex()
	switch tok {
	case Newline:
		panic(p.errorf(pos, "unexpected newline"))
	case Colon, Comma, Bar:
		return p.parseString()
	}
	if tok != Identifier {
		panic(p.errorf(pos, "expected identifier, got %q", lit))
	}
	return strings.TrimSuffix(lit, "\r")
}

func (p *Parser) parseActive() *Active {
	s := p.parseString()
	ret := &Active{}
	charMap := map[string]func(){
		"S": func() {
			ret.Spindle = true
		},
		"F": func() {
			ret.Flood = true
		},
		"M": func() {
			ret.Mist = true
		},
	}
	for _, c := range s {
		f, ok := charMap[string(c)]
		if !ok {
			panic(p.errorf(p.lexer.pos, "unknown identifier %q", c))
		}
		f()
	}

	return ret

}

func (p *Parser) parseInt() int {
	pos, tok, lit := p.lexer.Lex()
	switch tok {
	case Newline:
		panic(p.errorf(pos, "unexpected newline"))
	case Colon, Comma:
		return p.parseInt()
	}
	if tok != Float {
		panic(p.errorf(pos, "expected float, got %q", lit))
	}
	f, err := strconv.ParseInt(lit, 10, 32)
	if err != nil {
		panic(p.errorf(pos, "expected float, got %q", lit))
	}
	return int(f)
}

func (p *Parser) parseLimitPins(s *Status) *LimitPins {
	s.LimitPins = &LimitPins{}
	for {
		pos, tok, lit := p.lexer.Lex()
		switch tok {
		case Identifier:
			switch lit {
			case "X":
				s.LimitPins.X = true
			case "Y":
				s.LimitPins.Y = true
			case "Z":
				s.LimitPins.Z = true
			default:
				panic(p.errorf(pos, "unknown identifier %q", lit))
			}
		case Comma, Colon:
			continue
		case Bar:
			return s.LimitPins
		case CloseAngle:
			p.lexer.backup()
			return s.LimitPins
		}
	}
}

func (p *Parser) parseStatusUpdate(s *Status) (*Status, error) {
	for {
		pos, tok, lit := p.lexer.Lex()
		switch tok {
		case Identifier:
			switch lit {
			case "Idle", "Run", "Hold", "Home":
				s.State = strings.ToLower(lit)
				s.Alarm = nil
				s.Error = nil
			case "Error":
				s.State = "error"
			case "Alarm":
				s.State = "alarm"
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
			case "Pn":
				s.LimitPins = p.parseLimitPins(s)
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

type Error int

func (e Error) IsStatusUpdate() {}

type Alarm int

func (a Alarm) IsStatusUpdate() {}

func (p *Parser) parseError() (Error, error) {
	return Error(p.parseInt()), nil
}

func (p *Parser) parseAlarm() (Alarm, error) {
	return Alarm(p.parseInt()), nil
}

func (p *Parser) discard() error {
	_, err := io.Copy(io.Discard, p.lexer.rdr)
	return err
}
func (p *Parser) Parse() (ret StatusUpdate, err error) {
	status := new(Status)
	for {
		pos, tok, lit := p.lexer.Lex()
		switch tok {
		case Newline:

		case CloseAngle:
			return ret, p.discard()
		case OpenAngle:
			ret, err = p.parseStatusUpdate(status)
			if err != nil {
				return nil, err
			}
			return ret, p.discard()
		case Identifier:
			switch lit {
			case "ok":
				return &Ack{}, p.discard()
			case "error":
				ret, err = p.parseError()
				if err != nil {
					return nil, err
				}
				return ret, p.discard()
			case "alarm":
				ret, err := p.parseAlarm()
				if err != nil {
					return nil, err
				}
				return ret, p.discard()
			default:
				err = fmt.Errorf("unknown identifier %q", lit)
				_ = p.discard()
				return nil, err
			}
		default:
			err = p.errorf(pos, "expected identifier, got %q", lit)
			_ = p.discard()
			return nil, err
		}
	}
}
