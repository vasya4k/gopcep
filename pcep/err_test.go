package pcep

import "testing"

func TestParseErrObj(t *testing.T) {
	data := []byte{
		0: 13,
		1: 1 << 4,
		2: 0,
		3: 8,
		4: 0,
		5: 0,
		6: 11,
		7: 0,
	}
	e, err := parseErrObj(data)
	if err != nil {
		t.Errorf("must not see any errors, instead got: %s", err.Error())
		return
	}
	if e.ErrValueStr == "" {
		t.Errorf("ErrValueStr must not be empty")
	}
}
