package types_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/exonlabs/go-utils/pkg/types"
	"github.com/exonlabs/go-utils/tests"
)

type NDictAlias = map[string]any
type NDictAlt map[string]any
type NDictSlAlias = []map[string]any
type NDictSlAlt []map[string]any

func print_data(d any) string {
	return fmt.Sprintf(
		"\n%#v\n-----------------------------------------------", d)
}

func TestCreateNDict_MixedTypes(t *testing.T) {
	d := types.NewNDict(NDictAlt{
		"k1": "some value",
		"k2": map[string]any{"1": "111", "2": true, "3": 333},
		"k3": []int{1, 2, 3},
		"k4": NDictAlias{
			"a": []any{
				"1", 2, false, nil,
				NDictAlias{"x": "xxx", "y": nil, "z": "zzz"}},
			"b": map[string]any{
				"1": 111, "2": "222", "3": nil,
				"4": NDictAlt{"x": "xxx", "y": nil, "z": "zzz"},
				"5": NDictAlias{"x": "xxx", "y": nil, "z": "zzz"},
			},
		},
		"k5": []types.NDict{
			{"x": "xxx", "y": nil, "z": "zzz"},
			{"x": "xxx", "y": nil, "z": "zzz"},
			nil,
		},
	})
	t.Logf(">> input: %s", print_data(d))

	validation := types.NDict{
		"k1": "some value",
		"k2": types.NDict{"1": "111", "2": true, "3": 333},
		"k3": []int{1, 2, 3},
		"k4": types.NDict{
			"a": []any{
				"1", 2, false, nil,
				types.NDict{"x": "xxx", "y": nil, "z": "zzz"}},
			"b": types.NDict{
				"1": 111, "2": "222", "3": nil,
				"4": NDictAlt{"x": "xxx", "y": nil, "z": "zzz"},
				"5": types.NDict{"x": "xxx", "y": nil, "z": "zzz"}}},
		"k5": []types.NDict{
			{"x": "xxx", "y": nil, "z": "zzz"},
			{"x": "xxx", "y": nil, "z": "zzz"},
			{}},
	}
	if reflect.DeepEqual(d, validation) {
		t.Logf(tests.ValidMsg())
	} else {
		t.Errorf(tests.FailMsg()+" -- expecting: %s", print_data(validation))
	}
}

func TestCreateNDict_DeepNesting(t *testing.T) {
	d := types.NewNDict(NDictAlt{
		"k1": "some value",
		"k2": map[string]any{"1": "xxx", "2": "yyy"},
		"k3": NDictAlt{"1": "xxx", "2": "yyy"},
		"k4": NDictAlias{
			"1": "xxx",
			"2": NDictAlt{"1": "xxx", "2": "yyy"},
			"3": NDictAlias{
				"1": "xxx",
				"2": map[string]any{
					"1": "xxx",
					"2": map[string]any{"1": "xxx", "2": "yyy"},
					"3": []NDictAlias{
						{"1": "xxx",
							"2": NDictAlt{"1": "xxx", "2": "yyy"},
							"3": NDictAlias{
								"1": "xxx",
								"2": NDictAlt{"1": "xxx", "2": "yyy"},
							}},
						{"1": "xxx",
							"2": NDictAlt{"1": "xxx", "2": "yyy"},
							"3": NDictAlias{
								"1": "xxx",
								"2": NDictAlt{"1": "xxx", "2": "yyy"},
							}},
					},
				},
				"3": map[string]any{"1": "xxx", "2": "yyy"},
			},
		},
	})
	t.Logf(">> input: %s", print_data(d))

	validation := types.NDict{
		"k1": "some value",
		"k2": types.NDict{"1": "xxx", "2": "yyy"},
		"k3": NDictAlt{"1": "xxx", "2": "yyy"},
		"k4": types.NDict{
			"1": "xxx",
			"2": NDictAlt{"1": "xxx", "2": "yyy"},
			"3": types.NDict{
				"1": "xxx",
				"2": types.NDict{
					"1": "xxx",
					"2": types.NDict{"1": "xxx", "2": "yyy"},
					"3": []types.NDict{
						{"1": "xxx",
							"2": NDictAlt{"1": "xxx", "2": "yyy"},
							"3": types.NDict{
								"1": "xxx",
								"2": NDictAlt{"1": "xxx", "2": "yyy"},
							}},
						{"1": "xxx",
							"2": NDictAlt{"1": "xxx", "2": "yyy"},
							"3": types.NDict{
								"1": "xxx",
								"2": NDictAlt{"1": "xxx", "2": "yyy"},
							}},
					},
				},
				"3": types.NDict{"1": "xxx", "2": "yyy"},
			},
		},
	}
	if reflect.DeepEqual(d, validation) {
		t.Logf(tests.ValidMsg())
	} else {
		t.Errorf(tests.FailMsg()+" -- expecting: %s", print_data(validation))
	}
}

func TestCreateNDictSlice_MixedTypes(t *testing.T) {
	d := types.NewNDictSlice([]NDictAlias{
		{
			"k1": "some value",
			"k2": map[string]any{"1": "111", "2": true, "3": 333},
			"k3": []int{1, 2, 3},
			"k4": NDictAlias{
				"a": []any{
					"1", 2, false, nil,
					NDictAlias{"x": "xxx", "y": nil, "z": "zzz"}},
			},
		},
		{
			"k1": "some value",
			"k4": NDictAlias{
				"b": map[string]any{
					"1": 111, "2": "222", "3": nil,
					"4": NDictAlt{"x": "xxx", "y": nil, "z": "zzz"},
					"5": NDictAlias{"x": "xxx", "y": nil, "z": "zzz"}},
			},
		},
	})
	t.Logf(">> input: %s", print_data(d))

	validation := []types.NDict{
		{
			"k1": "some value",
			"k2": types.NDict{"1": "111", "2": true, "3": 333},
			"k3": []int{1, 2, 3},
			"k4": types.NDict{
				"a": []any{
					"1", 2, false, nil,
					types.NDict{"x": "xxx", "y": nil, "z": "zzz"}},
			},
		},
		{
			"k1": "some value",
			"k4": types.NDict{
				"b": types.NDict{
					"1": 111, "2": "222", "3": nil,
					"4": NDictAlt{"x": "xxx", "y": nil, "z": "zzz"},
					"5": types.NDict{"x": "xxx", "y": nil, "z": "zzz"}},
			},
		},
	}
	if reflect.DeepEqual(d, validation) {
		t.Logf(tests.ValidMsg())
	} else {
		t.Errorf(tests.FailMsg()+" -- expecting: %s", print_data(validation))
	}
}
func TestStripNDict_MixedTypes(t *testing.T) {
	buff := NDictAlias{
		"k1": "some value",
		"k2": map[string]any{"1": "111", "2": true, "3": 333},
		"k3": []int{1, 2, 3},
		"k4": NDictAlias{
			"a": []any{
				"1", 2, false, nil,
				NDictAlias{"x": "xxx", "y": nil, "z": "zzz"}},
			"b": map[string]any{
				"1": 111, "2": "222", "3": nil,
				"4": NDictAlt{"x": "xxx", "y": nil, "z": "zzz"},
				"5": NDictAlias{"x": "xxx", "y": nil, "z": "zzz"},
			},
		},
		"k5": []NDictAlias{
			{"x": "xxx", "y": nil, "z": "zzz"},
			{"x": "xxx", "y": nil, "z": "zzz"},
			nil,
		},
	}
	t.Logf(">> input: %s", print_data(buff))

	rbuff := types.StripNDict(types.NewNDict(buff))
	if !reflect.DeepEqual(rbuff, buff) {
		t.Errorf(tests.RedText("FAILED, 1 level strip check")+
			" -- got: %s", print_data(rbuff))
	} else {
		t.Logf(tests.ValidMsg())
	}
}

func TestStripNDictSlice_MixedTypes(t *testing.T) {
	buff := []NDictAlias{
		{
			"k1": "some value",
			"k2": map[string]any{"1": "111", "2": true, "3": 333},
			"k3": []int{1, 2, 3},
			"k4": NDictAlias{
				"a": []any{
					"1", 2, false, nil,
					NDictAlias{"x": "xxx", "y": nil, "z": "zzz"}},
			},
		},
		{
			"k1": "some value",
			"k4": NDictAlias{
				"b": map[string]any{
					"1": 111, "2": "222", "3": nil,
					"4": NDictAlt{"x": "xxx", "y": nil, "z": "zzz"},
					"5": NDictAlias{"x": "xxx", "y": nil, "z": "zzz"}},
			},
		},
	}
	t.Logf(">> input: %s", print_data(buff))

	rbuff := types.StripNDictSlice(types.NewNDictSlice(buff))
	if !reflect.DeepEqual(rbuff, buff) {
		t.Errorf(tests.RedText("FAILED, 1 level strip check")+
			" -- got: %s", print_data(rbuff))
	} else {
		t.Logf(tests.ValidMsg())
	}
}
func TestNDict_Keys(t *testing.T) {
	d := types.NewNDict(NDictAlt{
		"k1": "some value",
		"k2": map[string]any{"1": "xxx", "2": "yyy"},
		"k3": NDictAlt{"1": "xxx", "2": "yyy"},
		"k4": NDictAlias{
			"1": "xxx",
			"2": NDictAlt{"1": "xxx", "2": "yyy"},
			"3": NDictAlias{
				"1": "xxx",
				"2": map[string]any{
					"1": "xxx",
					"2": map[string]any{"1": "xxx", "2": "yyy"},
					"3": []NDictAlias{
						{"1": "xxx",
							"2": NDictAlt{"1": "xxx", "2": "yyy"},
							"3": NDictAlias{
								"1": "xxx",
								"2": NDictAlt{"1": "xxx", "2": "yyy"},
							}},
						{"1": "xxx",
							"2": NDictAlt{"1": "xxx", "2": "yyy"},
							"3": NDictAlias{
								"1": "xxx",
								"2": NDictAlt{"1": "xxx", "2": "yyy"},
							}},
					},
				},
				"3": map[string]any{"1": "xxx", "2": "yyy"},
			},
		},
	})
	t.Logf(">> input: %s", print_data(d))

	validations := []map[int][]string{
		{1: {"k1", "k2", "k3", "k4"}},
		{2: {"k1", "k2.1", "k2.2", "k3", "k4.1", "k4.2", "k4.3"}},
		{3: {"k1", "k2.1", "k2.2", "k3", "k4.1", "k4.2",
			"k4.3.1", "k4.3.2", "k4.3.3"}},
		{4: {"k1", "k2.1", "k2.2", "k3", "k4.1", "k4.2",
			"k4.3.1", "k4.3.2.1", "k4.3.2.2", "k4.3.2.3",
			"k4.3.3.1", "k4.3.3.2"}},
		{5: {"k1", "k2.1", "k2.2", "k3", "k4.1", "k4.2",
			"k4.3.1", "k4.3.2.1", "k4.3.2.2.1", "k4.3.2.2.2",
			"k4.3.2.3", "k4.3.3.1", "k4.3.3.2"}},
		{6: {"k1", "k2.1", "k2.2", "k3", "k4.1", "k4.2",
			"k4.3.1", "k4.3.2.1", "k4.3.2.2.1", "k4.3.2.2.2",
			"k4.3.2.3", "k4.3.3.1", "k4.3.3.2"}},
		{0: {"k1", "k2.1", "k2.2", "k3", "k4.1", "k4.2",
			"k4.3.1", "k4.3.2.1", "k4.3.2.2.1", "k4.3.2.2.2",
			"k4.3.2.3", "k4.3.3.1", "k4.3.3.2"}},
		{-1: {"k1", "k2.1", "k2.2", "k3", "k4.1", "k4.2",
			"k4.3.1", "k4.3.2.1", "k4.3.2.2.1", "k4.3.2.2.2",
			"k4.3.2.3", "k4.3.3.1", "k4.3.3.2"}},
	}
	for _, validation := range validations {
		for din, dout := range validation {
			res := d.KeysN(din)
			t.Logf("--- lvl %v = %#v", din, res)
			if reflect.DeepEqual(res, dout) {
				t.Logf(tests.ValidMsg())
			} else {
				t.Errorf(tests.RedText(
					fmt.Sprintf("FAILED check for lvl: %v", din)))
			}
		}
	}
}

func TestNDict_KeyExist(t *testing.T) {
	d := types.NewNDict(NDictAlias{
		"k1": "some value",
		"k2": map[string]any{"1": "xxx", "2": "yyy"},
		"k3": NDictAlt{"1": "xxx", "2": "yyy"},
		"k4": NDictAlias{
			"1": "xxx",
			"2": NDictAlt{"1": "xxx", "2": "yyy"},
			"3": NDictAlias{
				"1": "xxx",
				"2": map[string]any{
					"1": "xxx",
					"2": map[string]any{"1": "xxx", "2": "yyy"},
				},
				"3": map[string]any{"1": "xxx", "2": "yyy"},
			},
		},
	})
	t.Logf(">> input: %s", print_data(d))

	validations := []map[string]bool{
		{"k1": true}, {"k1.xx": false},
		{"k2.1": true}, {"k2.1.xx": false},
		{"k2.2": true}, {"k2.2.xx": false},
		{"k3": true}, {"k3.xx": false},
		{"k4.1": true}, {"k4.1.xx": false},
		{"k4.2": true}, {"k4.2.xx": false},
		{"k4.3.1": true}, {"k4.3.1.xx": false},
		{"k4.3.2.1": true}, {"k4.3.2.1.xx": false},
		{"k4.3.2.2.1": true}, {"k4.3.2.2.1.xx": false},
		{"k4.3.2.2.2": true}, {"k4.3.2.2.2.xx": false},
		{"k4.3.3.1": true}, {"k4.3.3.1.xx": false},
		{"k4.3.3.2": true}, {"k4.3.3.2.xx": false},
	}
	for _, validation := range validations {
		for din, dout := range validation {
			res := d.IsExist(din)
			t.Logf("--- %v exist %#v", din, res)
			if reflect.DeepEqual(res, dout) {
				t.Logf(tests.ValidMsg())
			} else {
				t.Errorf(tests.RedText(
					fmt.Sprintf("FAILED check for key: %v", din)))
			}
		}
	}
}

func TestNDict_Get(t *testing.T) {
	d := types.NewNDict(NDictAlias{
		"k1": "some value",
		"k2": map[string]any{"1": "xxx", "2": "yyy"},
		"k3": NDictAlt{"1": "xxx", "2": "yyy"},
		"k4": NDictAlias{"1": "xxx", "2": "yyy"},
		"k5": nil,
		"k6": []int{1, 2, 3},
		"k7": NDictAlias{
			"t": []NDictAlias{
				{"1": "xxx",
					"2": NDictAlt{"1": "xxx", "2": "yyy"},
					"3": NDictAlias{
						"1": "xxx",
						"2": NDictAlias{"1": "xxx", "2": "yyy"},
					}},
				{"1": "xxx",
					"2": NDictAlt{"1": "xxx", "2": "yyy"},
					"3": NDictAlias{
						"1": "xxx",
						"2": NDictAlias{"1": "xxx", "2": "yyy"},
					}},
			},
		},
	})
	t.Logf(">> input: %s", print_data(d))

	validations := []map[string]any{
		{"k1": "some value"}, {"k1.xx": nil},
		{"k2": types.NDict{"1": "xxx", "2": "yyy"}},
		{"k2.1": "xxx"}, {"k2.2": "yyy"}, {"k2.xx.yy": nil},
		{"k3": NDictAlt{"1": "xxx", "2": "yyy"}},
		{"k3.xx": nil}, {"k3.yy.zz": nil}, {"k3.2.3": nil},
		{"k4": types.NDict{"1": "xxx", "2": "yyy"}},
		{"k4.1": "xxx"}, {"k4.2": "yyy"}, {"k4.xx.yy": nil},
		{"k5": nil}, {"k5.xx": nil}, {"k5.yy.zz": nil},
		{"k6": []int{1, 2, 3}}, {"k6.xx": nil},
	}
	for _, validation := range validations {
		for k, v := range validation {
			res := d.Get(k, nil)
			t.Logf("--- %v = %#v", k, res)
			if reflect.DeepEqual(res, v) {
				t.Logf(tests.ValidMsg())
			} else {
				t.Errorf(tests.RedText(
					fmt.Sprintf("FAILED check for key: %v", k)))
			}
		}
	}

	// sub dicts checks
	for i, v := range d.Get("k7.t", []NDictAlias{}).([]types.NDict) {
		k := fmt.Sprintf("k7.t[%v].3.2.1", i)
		res := v.Get("3.2.1", nil)
		t.Logf("--- %v = %#v", k, res)
		if reflect.DeepEqual(res, "xxx") {
			t.Logf(tests.ValidMsg())
		} else {
			t.Errorf(tests.RedText(
				fmt.Sprintf("FAILED check for key: %v", k)))
		}
	}
}

func TestNDict_Set(t *testing.T) {
	d := types.NewNDict(nil)
	t.Logf(">> input: %s", print_data(d))

	validations := []map[string]any{
		{"k1": "some value"},
		{"k2": NDictAlias{"1": "xxx", "2": "yyy"}},
		{"k2.3": "333"},
		{"k3": types.NDict{"1": "xxx", "2": "yyy"}},
		{"k3.3": "333"},
		{"k3.3.1.2.1": "111"}, {"k3.3.1.2.2": "222"},
	}
	for _, validation := range validations {
		for k, v := range validation {
			d.Set(k, v)
			t.Logf(">> set %v = %#v\n--- %#v", k, v, d)
			res := d.Get(k, nil)
			if reflect.DeepEqual(res, v) {
				t.Logf(tests.ValidMsg())
			} else {
				t.Errorf(tests.RedText(
					fmt.Sprintf("FAILED 'Set' check for key: %v", k)))
			}
		}
	}
}

func TestNDict_Del(t *testing.T) {
	d := types.NewNDict(NDictAlt{
		"k1": "some value",
		"k2": map[string]any{"1": "111", "2": true, "3": 333},
		"k3": []int{1, 2, 3},
		"k4": NDictAlias{
			"a": []any{"1", 2, false, nil},
			"b": map[string]any{
				"1": 111, "2": "222", "3": nil,
				"4": NDictAlt{"x": "xxx", "y": nil, "z": "zzz"},
				"5": NDictAlias{"x": "xxx", "y": nil, "z": "zzz"},
			},
		},
	})
	t.Logf(">> input: %s", print_data(d))

	validations := []string{
		"k3.2.xxx", "k4.b.3", "k4.b.4",
	}
	for _, k := range validations {
		d.Del(k)
		t.Logf(">> Del %v\n--- %#v", k, d)
		res := d.Get(k, nil)
		if reflect.DeepEqual(res, nil) {
			t.Logf(tests.ValidMsg())
		} else {
			t.Errorf(tests.RedText(
				fmt.Sprintf("FAILED 'Del' check for key: %v", k)))
		}
	}
}

func TestNDict_Update(t *testing.T) {
	d := types.NewNDict(NDictAlt{
		"k1": "some value",
		"k2": map[string]any{"1": "111", "2": true, "3": 333},
		"k3": []int{1, 2, 3},
		"k4": NDictAlias{
			"a": []any{"1", 2, false, nil},
			"b": map[string]any{
				"1": 111, "2": "222", "3": nil,
				"4": NDictAlt{"x": "xxx", "y": nil, "z": "zzz"},
				"5": NDictAlias{"x": "xxx", "y": nil, "z": "zzz"},
			},
		},
	})
	t.Logf(">> input: %s", print_data(d))

	updt := map[string]any{
		"k3": map[string]any{"1": "111", "2": 222},
		"k4": map[string]any{
			"b": map[string]any{"5": 555, "6": 666},
			"c": "ccc",
		},
		"3": 333}
	t.Logf(">> update: %s", print_data(updt))

	validation := types.NDict{
		"k1": "some value",
		"k2": types.NDict{"1": "111", "2": true, "3": 333},
		"k3": types.NDict{"1": "111", "2": 222},
		"k4": types.NDict{
			"a": []any{"1", 2, false, nil},
			"b": types.NDict{
				"1": 111, "2": "222", "3": nil,
				"4": NDictAlt{"x": "xxx", "y": nil, "z": "zzz"},
				"5": 555, "6": 666},
			"c": "ccc",
		},
		"3": 333,
	}

	d.Update(updt)
	if reflect.DeepEqual(d, validation) {
		t.Logf(tests.ValidMsg() +
			fmt.Sprintf(" --- result: %s", print_data(d)))
	} else {
		t.Errorf(tests.RedText("FAILED 'Updt' check") +
			fmt.Sprintf(" --- got: %s", print_data(d)))
	}
}
