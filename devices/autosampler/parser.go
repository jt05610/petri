package autosampler

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type Token int

const (
	Return Token = iota
	Space
	Newline
	Comma
	Identifier
	Int
	String
	END
)

var tokens = []string{
	Return:     "RETURN",
	Space:      " ",
	Newline:    "NEWLINE",
	Comma:      ",",
	Identifier: "IDENT",
	Int:        "INT",
	String:     "STRING",
	END:        "END",
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
			if err == io.EOF {
				return l.pos, END, END.String()
			}
			return l.pos, Newline, Newline.String()
		}
		switch r {
		case ' ':
			return l.pos, Space, Space.String()
		case '\n':
			return l.pos, Newline, Newline.String()
		case '\r':
			return l.pos, Return, Return.String()
		case ',':
			return l.pos, Comma, Comma.String()
		default:
			startPos := l.pos
			if unicode.IsDigit(r) {
				l.backup()
				return startPos, Int, l.lexInt()
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

func (p *Parser) parseString(start string) string {
	bld := strings.Builder{}
	bld.WriteString(start)
	for {
		_, tok, lit := p.lexer.Lex()
		switch tok {
		case Newline, Return:
			return bld.String()
		default:
			bld.WriteString(lit)
		}
	}
}

func (p *Parser) parseInt() int {
	str := p.parseString("")
	i, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return i
}

func (p *Parser) discard() error {
	_, err := io.Copy(io.Discard, p.lexer.rdr)
	return err
}

func (p *Parser) parseIntResponse(lit string) *Response {
	ret := new(Response)
	i, err := strconv.Atoi(lit)
	if err != nil {
		panic(err)
	}
	ret.Data = IntData(i)
	return ret
}

func (p *Parser) parseResp(h Header) *Response {
	ret := new(Response)
	ret.Request = new(Request)
	ret.Header = h
	seenInt := 0
	respInts := make([]int, 0)
	for {
		pos, tok, lit := p.lexer.Lex()
		switch tok {
		case Newline, Return:
			if len(respInts) > 0 {
				if len(respInts) == 1 {
					ret.Data = IntData(respInts[0])
					return ret
				}
				ret.Data = IntArrayData(respInts)
				return ret
			}
			return ret
		case Comma:
			continue
		case Int:
			seenInt++
			i, err := strconv.Atoi(lit)
			if err != nil {
				panic(err)
			}
			if seenInt == 1 {
				ret.Id = i
				continue
			}
			respInts = append(respInts, i)
			seenInt++
		case Identifier:
			ret.Data = StringData(p.parseString(lit))
			return ret
		default:
			panic(p.errorf(pos, "expected identifier, got %q", lit))
		}

	}
}

func (p *Parser) Parse() (ret []*Response, err error) {
	ret = make([]*Response, 0)
	for {
		pos, tok, lit := p.lexer.Lex()
		switch tok {
		case Int:
			ret = append(ret, p.parseIntResponse(lit))
		case Identifier:
			switch lit {
			case "G", "S":
				ret = append(ret, p.parseResp(Header(lit[0])))
			case "I":
				// this should be at the front of the list always
				ret = append([]*Response{
					{
						Request: &Request{
							Header: Inject,
						},
					},
				}, ret...)
			default:
				err = fmt.Errorf("unknown identifier %q", lit)
				return nil, err
			}
		case Newline, Return:
			continue
		case END:
			return ret, nil
		default:
			err = p.errorf(pos, "expected identifier, got %q", lit)
			return nil, err
		}
	}
}
