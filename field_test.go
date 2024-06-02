package statusx_test

import (
	"testing"

	. "github.com/xoctopus/statusx"
)

func TestErrorField(t *testing.T) {
	fields := ErrorFields{}
	fields = append(fields, NewErrorField("body", "field3", "msg3"))
	fields = append(fields, NewErrorField("param", "field2", "msg2"))
	fields = append(fields, NewErrorField("body", "field1", "msg1"))

	t.Log(fields.String())
}
