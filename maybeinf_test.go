package png

import (
	"math"
	"testing"
)

func TestMarshalJson(t *testing.T) {
	cases := []struct {
		value float64
		want  string
	}{
		{0, "0"},
		{math.Inf(1), "null"},
	}

	for _, c := range cases {
		mi := MaybeInf(c.value)
		got, err := mi.MarshalJSON()
		if err != nil {
			t.Errorf("MarshalJson(%v) got error: %v", c.value, err)
		}
		if string(got) != c.want {
			t.Errorf("MarshalJson(%v) = %v, want %v", c.value, got, c.want)
		}
	}
}

func TestUnmarshalJson(t *testing.T) {
	cases := []struct {
		data string
		want MaybeInf
	}{
		{"0", 0},
		{"null", MaybeInf(math.Inf(1))},
	}

	for _, c := range cases {
		var mi MaybeInf
		err := mi.UnmarshalJSON([]byte(c.data))
		if err != nil {
			t.Errorf("UnmarshalJson(%v) got error: %v", c.data, err)
		}
		if mi != c.want {
			t.Errorf("UnmarshalJson(%v) = %v, want %v", c.data, mi, c.want)
		}
	}
}
