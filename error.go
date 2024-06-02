package statusx

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/xoctopus/x/ptrx"
)

type Error interface {
	StatusErr() *StatusErr
	Error() string
}

type ServiceCode interface {
	ServiceCode() int
}

func Is(err error) bool {
	if err == nil {
		return false
	}

	return errors.As(err, ptrx.Ptr(Error(nil))) || errors.As(err, ptrx.Ptr(&StatusErr{}))
}

func As(err error) *StatusErr {
	if err == nil {
		return nil
	}

	var e Error
	if errors.As(err, &e) {
		return e.StatusErr()
	}

	var se *StatusErr
	if ok := errors.As(err, &se); ok {
		return se
	}

	return UnknownErr(err)
}

func Wrap(err error, code int, key string, message ...string) *StatusErr {
	if err == nil {
		return nil
	}
	if code >= 100 && code <= 999 {
		code = code * 1e6
	}
	msg := key
	if len(message) > 0 {
		msg = message[0]
	}

	desc := err.Error()
	if len(message) > 1 {
		desc = strings.Join(message[1:], "\n")
	}

	// err = errors.WithMessage(err, "xxx")
	s := &StatusErr{
		Key:   key,
		Code:  code,
		Msg:   msg,
		Desc:  desc,
		error: errors.WithStack(err),
	}

	return s
}

func UnknownErr(err error) *StatusErr {
	se := NewStatusErr("UnknownError", 500, "unknown error")
	se.error = errors.WithStack(err)
	return se
}

func NewStatusErr(key string, code int, msg string) *StatusErr {
	if code >= 100 && code <= 999 {
		code = code * 1e6
	}
	return &StatusErr{
		Key:  key,
		Code: code,
		Msg:  msg,
	}
}

// @err[UnknownError][500000000][unknown error]
var regexpStatusErrSummary = regexp.MustCompile(`@StatusErr\[(.+)\]\[(.+)\]\[(.+)\](!)?`)

func ParseSummary(s string) (*StatusErr, error) {
	if !regexpStatusErrSummary.Match([]byte(s)) {
		return nil, fmt.Errorf("unsupported status err summary: %s", s)
	}

	matched := regexpStatusErrSummary.FindStringSubmatch(s)

	code, _ := strconv.ParseInt(matched[2], 10, 64)

	se := NewStatusErr(matched[1], int(code), matched[3])
	se.CanBeTalk = matched[4] != ""
	return se, nil
}

type StatusErr struct {
	Key       string      `json:"key"       xml:"key"`
	Code      int         `json:"code"      xml:"code"`      // unique err code
	Msg       string      `json:"msg"       xml:"msg"`       // msg of err
	Desc      string      `json:"desc"      xml:"desc"`      // desc of err
	CanBeTalk bool        `json:"canBeTalk" xml:"canBeTalk"` // can be task error; for client to should error msg to end user
	ID        string      `json:"id"        xml:"id"`        // request ID or other request context
	Sources   []string    `json:"sources"   xml:"sources"`   // error tracing
	Fields    ErrorFields `json:"fields"    xml:"fields"`    // error in where fields
	error     error
}

func (se *StatusErr) Unwrap() error {
	return se.error
}

func (se *StatusErr) Summary() string {
	s := fmt.Sprintf(`@StatusErr[%s][%d][%s]`, se.Key, se.Code, se.Msg)

	if se.CanBeTalk {
		return s + "!"
	}
	return s
}

func (se *StatusErr) Is(err error) bool {
	if se == nil {
		return false
	}
	e := As(err)
	return e != nil && e.Key == se.Key && e.Code == se.Code
}

func StatusCodeFromCode(code int) int {
	sc := fmt.Sprintf("%d", code)
	if len(sc) < 3 {
		return 0
	}
	statusCode, _ := strconv.Atoi(sc[:3])
	return statusCode
}

func (se *StatusErr) StatusCode() int {
	return StatusCodeFromCode(se.Code)
}

func (se *StatusErr) Error() string {
	s := fmt.Sprintf(
		"[%s]%s%s",
		strings.Join(se.Sources, ","),
		se.Summary(),
		se.Fields,
	)

	if se.Desc != "" {
		s += " " + se.Desc
	}

	return s
}

func (se StatusErr) WithMsg(msg string) *StatusErr {
	se.Msg = msg
	return &se
}

func (se StatusErr) WithDesc(desc string) *StatusErr {
	se.Desc = desc
	return &se
}

func (se StatusErr) WithID(id string) *StatusErr {
	se.ID = id
	return &se
}

func (se StatusErr) AppendSource(source string) *StatusErr {
	length := len(se.Sources)
	if length == 0 || se.Sources[length-1] != source {
		se.Sources = append(se.Sources, source)
	}
	return &se
}

func (se StatusErr) EnableErrTalk() *StatusErr {
	se.CanBeTalk = true
	return &se
}

func (se StatusErr) DisableErrTalk() *StatusErr {
	se.CanBeTalk = false
	return &se
}

func (se StatusErr) AppendErrorField(in string, field string, msg string) *StatusErr {
	se.Fields = append(se.Fields, NewErrorField(in, field, msg))
	return &se
}

func (se StatusErr) AppendErrorFields(errorFields ...*ErrorField) *StatusErr {
	se.Fields = append(se.Fields, errorFields...)
	return &se
}
