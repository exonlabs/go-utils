package types_test

import (
	"reflect"
	"testing"

	"github.com/exonlabs/go-utils/pkg/types"
	"github.com/exonlabs/go-utils/tests"
)

type DictAlias = map[string]any
type DictAlt map[string]any
type DictSlAlias = []map[string]any
type DictSlAlt []map[string]any

func TestCreateDict_Clone(t *testing.T) {
	d := types.NewDict(DictAlt{
		"k1": "some value",
		"":   "empty key name",
		"k2": map[string]any{"1": "111", "2": true, "3": 333},
		"k3": []int{1, 2, 3},
		"k4": "444",
		"k5": []types.Dict{
			{"x": "xxx", "y": nil, "z": "zzz", "": "empty key name"},
			{"x": "xxx", "y": nil, "z": "zzz"},
			nil,
		},
	})
	t.Logf(">> input source: %s", tests.PrintData(d))

	v, err := types.CloneDict(d)
	if err != nil {
		t.Errorf(tests.FailMsg()+" -- %s", err.Error())
	}
	t.Logf(">> input Clone: %s", tests.PrintData(v))

	v["k4"] = "4444444444"
	v["k5"].([]types.Dict)[1]["t"] = "ttt"
	t.Logf(">> input Clone Modified: %s", tests.PrintData(v))

	validation1 := types.Dict{
		"k1": "some value",
		"k2": types.Dict{"1": "111", "2": true, "3": 333},
		"k3": []int{1, 2, 3},
		"k4": "444",
		"k5": []types.Dict{
			{"x": "xxx", "y": nil, "z": "zzz"},
			{"x": "xxx", "y": nil, "z": "zzz"},
			{}},
	}
	if reflect.DeepEqual(d, validation1) {
		t.Logf(tests.ValidMsg() + " -- orginal data")
	} else {
		t.Errorf(tests.FailMsg()+" -- expecting source: %s",
			tests.PrintData(validation1))
	}

	validation2 := types.Dict{
		"k1": "some value",
		"k2": types.Dict{"1": "111", "2": true, "3": 333},
		"k3": []int{1, 2, 3},
		"k4": "4444444444",
		"k5": []types.Dict{
			{"x": "xxx", "y": nil, "z": "zzz"},
			{"x": "xxx", "y": nil, "z": "zzz", "t": "ttt"},
			{}},
	}
	if reflect.DeepEqual(v, validation2) {
		t.Logf(tests.ValidMsg() + " -- cloned and modified data")
	} else {
		t.Errorf(tests.FailMsg()+" -- expecting clone: %s",
			tests.PrintData(validation1))
	}
}
