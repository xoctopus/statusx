package statusx_test

import (
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	. "github.com/xoctopus/statusx"
)

type TestError struct {
	e *StatusErr
}

func (e *TestError) StatusErr() *StatusErr {
	return e.e
}

func (e *TestError) Error() string {
	return e.e.Error()
}

func TestStatusErr(t *testing.T) {
	t.Run("Is", func(t *testing.T) {
		NewWithT(t).Expect(Is(nil)).To(BeFalse())

		e1 := NewStatusErr("KEY", http.StatusNotFound, "error message")
		NewWithT(t).Expect(Is(e1)).To(BeTrue())
		NewWithT(t).Expect(e1.Is(e1)).To(BeTrue())

		e2 := &TestError{e: e1}
		NewWithT(t).Expect(Is(e2)).To(BeTrue())

		e3 := errors.New("message")
		NewWithT(t).Expect(Is(e3)).To(BeFalse())
		NewWithT(t).Expect(e1.Is(nil)).To(BeFalse())
		NewWithT(t).Expect(e1.Is(e3)).To(BeFalse())
		NewWithT(t).Expect((*StatusErr)(nil).Is(e1)).To(BeFalse())
	})

	t.Run("As", func(t *testing.T) {
		NewWithT(t).Expect(As(nil)).To(BeNil())

		e1 := NewStatusErr("KEY", http.StatusNotFound, "error message")
		NewWithT(t).Expect(As(e1)).To(Equal(e1))

		e2 := &TestError{e: e1}
		NewWithT(t).Expect(As(e2)).To(Equal(e1))

		e3 := errors.New("message")
		as := As(e3)
		NewWithT(t).Expect(as.Error()).To(Equal(UnknownErr(e3).Error()))
	})

	t.Run("Wrap", func(t *testing.T) {
		e := Wrap(errors.New("raw"), http.StatusNotFound, "SpecifiedResourceNotFound")
		expect := "[]@StatusErr[SpecifiedResourceNotFound][404000000][SpecifiedResourceNotFound] raw"
		NewWithT(t).Expect(e.Error()).To(Equal(expect))

		e = Wrap(errors.New("raw"), http.StatusNotFound, "SpecifiedResourceNotFound", "msg0", "msg1")
		expect = "[]@StatusErr[SpecifiedResourceNotFound][404000000][msg0] msg1"
		NewWithT(t).Expect(e.Error()).To(Equal(expect))
		NewWithT(t).Expect(e.Desc).To(Equal("msg1"))
		NewWithT(t).Expect(e.Unwrap().Error()).To(Equal("raw"))

		e = Wrap(nil, http.StatusNotFound, "AnyKey")
		NewWithT(t).Expect(e).To(BeNil())
	})

	t.Run("ParseSummary", func(t *testing.T) {
		summary := "[src1,src2]@StatusErr[Key][404000001][Message]! talk"
		se, err := ParseSummary(summary)
		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(se.Key).To(Equal("Key"))
		NewWithT(t).Expect(se.Code).To(Equal(404000001))
		NewWithT(t).Expect(se.Msg).To(Equal("Message"))
		NewWithT(t).Expect(se.CanBeTalk).To(BeTrue())
		t.Run("InvalidSummary", func(t *testing.T) {
			summary2 := t.Name()
			_, err2 := ParseSummary(summary2)
			NewWithT(t).Expect(err2.Error()).To(ContainSubstring(t.Name()))
		})

		se = se.AppendSource("src1").
			AppendSource("src2").
			WithMsg("message").
			WithDesc("desc").
			WithID("100").
			AppendErrorField("body", "field1", "msg1").
			AppendErrorFields(NewErrorField("param", "field2", "msg2")).
			EnableErrTalk()

		summary = "[src1,src2]@StatusErr[Key][404000001][message]!<field1 in body - msg1, field2 in param - msg2> desc"
		NewWithT(t).Expect(se.Error()).To(Equal(summary))

		se = se.DisableErrTalk()
		NewWithT(t).Expect(se.Error()).To(Equal(strings.Replace(summary, "!", "", -1)))
		NewWithT(t).Expect(se.StatusCode()).To(Equal(404))

		se.Code = 10
		NewWithT(t).Expect(se.StatusCode()).To(Equal(0))

	})
}

/*
import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/sincospro/statusx"
//	"github.com/sincospro/statusx/generator"
//  . "github.com/sincospro/statusx/generator/__examples__"
//	"github.com/sincospro/x/pkgx"
)

func init() {
	cwd, _ := os.Getwd()
	pkg, _ := pkgx.LoadFrom(filepath.Join(cwd, "../statusxgen/__examples__"))

	g := statusxgen.New(pkg)

	g.Scan("StatusError")
	g.Output(cwd)
}

func TestStatusErr(t *testing.T) {
	var (
		unknownSeStr  = "@StatusErr[UnknownError][500000000][unknown error]"
		unauthedSeStr = "@StatusErr[Unauthorized][401999001][Unauthorized]!"
		internalSeStr = "@StatusErr[InternalServerError][500999001][InternalServerError 内部错误]"
		summary       = statusx.NewUnknownErr().Summary()
		se, err       = statusx.ParseStatusErrSummary(summary)
	)

	g := NewWithT(t)

	g.Expect(summary).To(Equal(unknownSeStr))
	g.Expect(err).To(BeNil())
	g.Expect(se).To(Equal(statusx.NewUnknownErr()))
	g.Expect(Unauthorized.StatusErr().Summary()).To(Equal(unauthedSeStr))
	g.Expect(InternalServerError.StatusErr().Summary()).To(Equal(internalSeStr))
	g.Expect(Unauthorized.StatusCode()).To(Equal(401))
	g.Expect(Unauthorized.StatusErr().StatusCode()).To(Equal(401))

	g.Expect(errors.Is(Unauthorized, Unauthorized)).To(BeTrue())
	g.Expect(errors.Is(Unauthorized.StatusErr(), Unauthorized)).To(BeTrue())
	g.Expect(errors.Is(Unauthorized.StatusErr(), Unauthorized.StatusErr())).To(BeTrue())
}

func ExampleStatusErr() {
	fmt.Println(Unauthorized)
	fmt.Println(statusx.FromErr(nil))
	fmt.Println(statusx.FromErr(fmt.Errorf("unknown")))
	fmt.Println(Unauthorized.StatusErr().WithMsg("msg overwrite"))
	fmt.Println(Unauthorized.StatusErr().WithDesc("desc overwrite"))
	fmt.Println(Unauthorized.StatusErr().DisableErrTalk().EnableErrTalk())
	fmt.Println(Unauthorized.StatusErr().WithID("111"))
	fmt.Println(Unauthorized.StatusErr().AppendSource("service-abc"))
	fmt.Println(Unauthorized.StatusErr().AppendErrorField("header", "Authorization", "missing"))
	fmt.Println(Unauthorized.StatusErr().AppendErrorFields(
		statusx.NewErrorField("query", "key", "missing"),
		statusx.NewErrorField("header", "Authorization", "missing"),
	))
	// Output:
	// []@StatusErr[Unauthorized][401999001][Unauthorized]!
	// <nil>
	// []@StatusErr[UnknownError][500000000][unknown error] unknown
	// []@StatusErr[Unauthorized][401999001][msg overwrite]!
	// []@StatusErr[Unauthorized][401999001][Unauthorized]! desc overwrite
	// []@StatusErr[Unauthorized][401999001][Unauthorized]!
	// []@StatusErr[Unauthorized][401999001][Unauthorized]!
	// [service-abc]@StatusErr[Unauthorized][401999001][Unauthorized]!
	// []@StatusErr[Unauthorized][401999001][Unauthorized]!<Authorization in header - missing>
	// []@StatusErr[Unauthorized][401999001][Unauthorized]!<Authorization in header - missing, key in query - missing>
}
*/
