package marlin

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
	E float32
}

type StatusUpdate interface {
	IsStatusUpdate()
}

type Processing struct {
}

func (p *Processing) IsStatusUpdate() {}

type Count struct {
	X int
	Y int
	Z int
}

type Status struct {
	Error    *int
	Alarm    *int
	Position *Position
	Count    *Count
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
	Space
	Newline
	Colon
	Comma
	Bar
	Identifier
	Float
)

var tokens = []string{
	Return:     "RETURN",
	Space:      "SPACE",
	Newline:    "NEWLINE",
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
		case ' ':
			return l.pos, Space, Space.String()
		case '\n':
			return l.pos, Newline, Newline.String()
		case '\r':
			return l.pos, Return, Return.String()
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
	case Colon:
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
	ret := new(Position)
	for i := 0; i < 4; {
		pos, tok, lit := p.lexer.Lex()
		switch tok {
		case Newline:
			panic(p.errorf(pos, "unexpected newline"))
		case Colon, Space, Identifier:
			continue
		default:
			if tok != Float {
				panic(p.errorf(pos, "expected float, got %q", lit))
			}
			f, err := strconv.ParseFloat(lit, 32)
			if err != nil {
				panic(p.errorf(pos, "expected float, got %q", lit))
			}
			switch i {
			case 0:
				ret.X = float32(f)
			case 1:
				ret.Y = float32(f)
			case 2:
				ret.Z = float32(f)
			case 3:
				ret.E = float32(f)
			}
			i++
		}

	}
	return ret
}

func (p *Parser) parseString() string {
	pos, tok, lit := p.lexer.Lex()
	switch tok {
	case Newline:
		panic(p.errorf(pos, "unexpected newline"))
	case Colon, Comma, Identifier, Space:
		return p.parseString()
	}
	if tok != Identifier {
		panic(p.errorf(pos, "expected identifier, got %q", lit))
	}
	return strings.TrimSuffix(lit, "\r")
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

func (p *Parser) parseCount() *Count {
	ret := new(Count)
	for i := 0; i < 3; {
		pos, tok, lit := p.lexer.Lex()
		switch tok {
		case Newline:
			return ret
		case Colon, Comma, Space, Bar, Identifier:
			continue
		}
		if tok != Float {
			panic(p.errorf(pos, "expected float, got %q", lit))
		}
		f, err := strconv.ParseInt(lit, 10, 32)
		if err != nil {
			panic(p.errorf(pos, "expected float, got %q", lit))
		}
		i++
		switch i {
		case 0:
			ret.X = int(f)
		case 1:
			ret.Y = int(f)
		case 2:
			ret.Z = int(f)
		}
	}
	return ret
}
func (p *Parser) parseStatusUpdate(s *Status) (*Status, error) {
	for {
		pos, tok, lit := p.lexer.Lex()
		switch tok {
		case Colon:
			s.Position = p.parsePosition()
		case Identifier:
			switch lit {
			case "Count":
				s.Count = p.parseCount()
			default:
				return nil, p.errorf(pos, "unknown identifier %q", lit)
			}
		case Bar, Comma, Space:
			continue
		case Newline:
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
		case Identifier:
			switch lit {
			case "X":
				return p.parseStatusUpdate(status)
			case "ok":
				return &Ack{}, p.discard()
			case "echo":
				return &Processing{}, p.discard()
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
