package bodyprovider

import (
	"encoding/json"
	"io"
	"net/url"
	"strings"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestReader(t *testing.T) {
	p := NewReader(strings.NewReader("hello"))
	assert.Eq(t, "", p.ContentType())

	r, err := p.Body()
	assert.NoErr(t, err)
	bs, _ := io.ReadAll(r)
	assert.Eq(t, "hello", string(bs))
}

func TestJSON(t *testing.T) {
	p := NewJSON(map[string]string{"name": "inhere", "age": "30"})
	assert.Eq(t, "application/json; charset=utf-8", p.ContentType())

	r, err := p.Body()
	assert.NoErr(t, err)
	bs, _ := io.ReadAll(r)

	var got map[string]string
	assert.NoErr(t, json.Unmarshal(bs, &got))
	assert.Eq(t, "inhere", got["name"])
	assert.Eq(t, "30", got["age"])
}

func TestForm(t *testing.T) {
	t.Run("url.Values", func(t *testing.T) {
		v := url.Values{}
		v.Set("name", "inhere")
		v.Set("age", "30")
		p := NewForm(v)
		assert.Eq(t, "application/x-www-form-urlencoded; charset=utf-8", p.ContentType())

		r, err := p.Body()
		assert.NoErr(t, err)
		bs, _ := io.ReadAll(r)
		s := string(bs)
		assert.Contains(t, s, "name=inhere")
		assert.Contains(t, s, "age=30")
	})

	t.Run("map[string][]string", func(t *testing.T) {
		m := map[string][]string{"k": {"v"}}
		r, err := NewForm(m).Body()
		assert.NoErr(t, err)
		bs, _ := io.ReadAll(r)
		assert.Eq(t, "k=v", string(bs))
	})

	t.Run("string passthrough", func(t *testing.T) {
		r, err := NewForm("name=inhere&age=30").Body()
		assert.NoErr(t, err)
		bs, _ := io.ReadAll(r)
		assert.Eq(t, "name=inhere&age=30", string(bs))
	})

	t.Run("invalid type", func(t *testing.T) {
		r, err := NewForm(12345).Body()
		assert.Err(t, err)
		assert.Nil(t, r)
		assert.ErrMsg(t, err, "formBodyProvider: invalid form data type: int")
	})
}

func TestMultipart_FieldsOnly(t *testing.T) {
	p := NewMultipart(nil, map[string]string{"name": "inhere"})
	assert.Eq(t, "", p.ContentType()) // not yet built

	r, err := p.Body()
	assert.NoErr(t, err)
	bs, _ := io.ReadAll(r)
	s := string(bs)
	assert.Contains(t, s, `name="name"`)
	assert.Contains(t, s, "inhere")
	// ContentType is populated after build, including boundary.
	assert.Contains(t, p.ContentType(), "multipart/form-data; boundary=")
}
