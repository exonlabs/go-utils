// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package xdict_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/exonlabs/go-utils/pkg/types/xdict"
)

type DictAlias = map[string]any
type DictAlt map[string]any
type DictSlAlias = []map[string]any
type DictSlAlt []map[string]any

func Test_KeysN(t *testing.T) {
	d := DictAlt{
		"k1": "some value",
		"k2": map[string]any{"1": "xxx", "2": "yyy"},
		"k3": DictAlt{"1": "xxx", "2": "yyy"},
		"k4": DictAlias{
			"1": "xxx",
			"2": DictAlt{"1": "xxx", "2": "yyy"},
			"3": DictAlias{
				"1": "xxx",
				"2": map[string]any{
					"1": "xxx",
					"2": map[string]any{"1": "xxx", "2": "yyy"},
					"3": []DictAlias{
						{
							"1": "xxx",
							"2": DictAlt{"1": "xxx", "2": "yyy"},
							"3": DictAlias{
								"1": "xxx",
								"2": DictAlt{"1": "xxx", "2": "yyy"},
							},
						},
						{
							"1": "xxx",
							"2": DictAlt{"1": "xxx", "2": "yyy"},
							"3": DictAlias{
								"1": "xxx",
								"2": DictAlt{"1": "xxx", "2": "yyy"},
							},
						},
					},
				},
				"3": map[string]any{"1": "xxx", "2": "yyy"},
			},
		},
	}
	t.Logf(">> input: \n%#v\n", d)

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
		for lvl, keys := range validation {
			res := xdict.KeysN(d, lvl)
			t.Logf("--- lvl %v = %#v", lvl, res)
			if reflect.DeepEqual(res, keys) {
				t.Logf("VALID")
			} else {
				t.Errorf("FAILED check for lvl: %v", lvl)
			}
		}
	}
}

func Test_IsExist(t *testing.T) {
	d := DictAlias{
		"k1": "some value",
		"k2": map[string]any{"1": "xxx", "2": "yyy"},
		"k3": DictAlt{"1": "xxx", "2": "yyy"},
		"k4": DictAlias{
			"1": "xxx",
			"2": DictAlt{"1": "xxx", "2": "yyy"},
			"3": DictAlias{
				"1": "xxx",
				"2": map[string]any{
					"1": "xxx",
					"2": map[string]any{"1": "xxx", "2": "yyy"},
				},
				"3": map[string]any{"1": "xxx", "2": "yyy"},
			},
		},
	}
	t.Logf(">> input: %v", d)

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
			res := xdict.IsExist(d, din)
			t.Logf("--- %v exist %#v", din, res)
			if reflect.DeepEqual(res, dout) {
				t.Logf("VALID")
			} else {
				t.Errorf("FAILED check for key: %v", din)
			}
		}
	}
}

func Test_Get(t *testing.T) {
	d := DictAlias{
		"k1": "some value",
		"k2": map[string]any{"1": "xxx", "2": "yyy"},
		"k3": DictAlt{"1": "xxx", "2": "yyy"},
		"k4": DictAlias{"1": "xxx", "2": "yyy"},
		"k5": nil,
		"k6": []int{1, 2, 3},
		"k7": DictAlias{
			"t": []DictAlias{
				{
					"1": "xxx",
					"2": DictAlt{"1": "xxx", "2": "yyy"},
					"3": DictAlias{
						"1": "xxx",
						"2": DictAlias{"1": "xxx", "2": "yyy"},
					},
				},
				{
					"1": "xxx",
					"2": DictAlt{"1": "xxx", "2": "yyy"},
					"3": DictAlias{
						"1": "xxx",
						"2": DictAlias{"1": "xxx", "2": "yyy"},
					},
				},
			},
		},
	}
	t.Logf(">> input: %v", d)

	validations := []map[string]any{
		{"k1": "some value"}, {"k1.xx": nil},
		{"k2": map[string]any{"1": "xxx", "2": "yyy"}},
		{"k2.1": "xxx"}, {"k2.2": "yyy"}, {"k2.xx.yy": nil},
		{"k3": DictAlt{"1": "xxx", "2": "yyy"}},
		{"k3.xx": nil}, {"k3.yy.zz": nil}, {"k3.2.3": nil},
		{"k4": DictAlias{"1": "xxx", "2": "yyy"}},
		{"k4.1": "xxx"}, {"k4.2": "yyy"}, {"k4.xx.yy": nil},
		{"k5": nil}, {"k5.xx": nil}, {"k5.yy.zz": nil},
		{"k6": []int{1, 2, 3}}, {"k6.xx": nil},
	}
	for _, validation := range validations {
		for k, v := range validation {
			res := xdict.Get(d, k, nil)
			t.Logf("--- %v = %#v", k, res)
			if reflect.DeepEqual(res, v) {
				t.Logf("VALID")
			} else {
				t.Errorf("FAILED check for key: %v", k)
			}
		}
	}

	// sub dicts checks
	for i, v := range xdict.Get(d, "k7.t", nil).([]DictAlias) {
		k := fmt.Sprintf("k7.t[%v].3.2.1", i)
		res := xdict.Get(v, "3.2.1", nil)
		t.Logf("--- %v = %#v", k, res)
		if reflect.DeepEqual(res, "xxx") {
			t.Logf("VALID")
		} else {
			t.Errorf("FAILED check for key: %v", k)
		}
	}
}

func Test_Fetch(t *testing.T) {
	d := DictAlias{
		"k1": "some value",
		"k2": map[string]any{"1": "xxx", "2": "yyy"},
		"k3": DictAlt{"1": "xxx", "2": "yyy"},
		"k4": DictAlias{"1": "xxx", "2": "yyy"},
		"k5": nil,
		"k6": []int{1, 2, 3},
		"k7": DictAlias{
			"t": []DictAlias{
				{
					"1": "xxx",
					"2": DictAlt{"1": "xxx", "2": "yyy"},
					"3": DictAlias{
						"1": "xxx",
						"2": DictAlias{"1": "xxx", "2": "yyy"},
					},
				},
				{
					"1": "xxx",
					"2": DictAlt{"1": "xxx", "2": "yyy"},
					"3": DictAlias{
						"1": "xxx",
						"2": DictAlias{"1": "xxx", "2": "yyy"},
					},
				},
			},
		},
	}
	t.Logf(">> input: %v", d)

	validations := []map[string]any{
		{"k1": "some value"},
		{"k2": map[string]any{"1": "xxx", "2": "yyy"}},
		{"k2.1": "xxx"}, {"k2.2": "yyy"},
		{"k3": DictAlt{"1": "xxx", "2": "yyy"}},
		{"k4": DictAlias{"1": "xxx", "2": "yyy"}},
		{"k4.1": "xxx"}, {"k4.2": "yyy"},
		{"k5": nil},
		{"k6": []int{1, 2, 3}},
	}
	for _, validation := range validations {
		for k, v := range validation {
			res := xdict.Fetch(d, k, v)
			t.Logf("--- %v = %#v", k, res)
			if reflect.DeepEqual(res, v) {
				t.Logf("VALID")
			} else {
				t.Errorf("FAILED check for key: %v", k)
			}
		}
	}

	// sub dicts checks
	for i, v := range xdict.Get(d, "k7.t", nil).([]DictAlias) {
		k := fmt.Sprintf("k7.t[%v].3.2.1", i)
		res := xdict.Get(v, "3.2.1", nil)
		t.Logf("--- %v = %#v", k, res)
		if reflect.DeepEqual(res, "xxx") {
			t.Logf("VALID")
		} else {
			t.Errorf("FAILED check for key: %v", k)
		}
	}
}

func Test_Set(t *testing.T) {
	d := xdict.Dict{}
	t.Logf(">> input: %v", d)

	validations := []map[string]any{
		{"k1": "some value"},
		{"k2": DictAlias{"1": "xxx", "2": "yyy"}},
		{"k2.3": "333"},
		{"k3": xdict.Dict{"1": "xxx", "2": "yyy"}},
		{"k3.3": "333"},
		{"k3.3.1.2.1": "111"}, {"k3.3.1.2.2": "222"},
	}
	for _, validation := range validations {
		for k, v := range validation {
			xdict.Set(d, k, v)
			t.Logf(">> set %v = %#v\n--- %#v", k, v, d)
			res := xdict.Get(d, k, nil)
			if reflect.DeepEqual(res, v) {
				t.Logf("VALID")
			} else {
				t.Errorf("FAILED 'Set' check for key: %v", k)
			}
		}
	}
}

func Test_Merge(t *testing.T) {
	d := DictAlias{
		"k1": "some value",
		"k2": map[string]any{"1": "111", "2": true, "3": 333},
		"k3": []int{1, 2, 3},
		"k4": DictAlias{
			"a": []any{"1", 2, false, nil},
			"b": map[string]any{
				"1": 111, "2": "222", "3": nil,
				"4": DictAlt{"x": "xxx", "y": nil, "z": "zzz"},
				"5": DictAlias{"x": "xxx", "y": nil, "z": "zzz"},
			},
		},
	}
	t.Logf(">> input: %v", d)

	updt := map[string]any{
		"k3": map[string]any{"1": "111", "2": 222},
		"k4": map[string]any{
			"b": map[string]any{"5": 555, "6": 666},
			"c": "ccc",
		},
		"3": 333}
	t.Logf(">> update: %v", updt)

	validation := DictAlias{
		"k1": "some value",
		"k2": DictAlias{"1": "111", "2": true, "3": 333},
		"k3": map[string]any{"1": "111", "2": 222},
		"k4": DictAlias{
			"a": []any{"1", 2, false, nil},
			"b": map[string]any{
				"1": 111, "2": "222", "3": nil,
				"4": DictAlt{"x": "xxx", "y": nil, "z": "zzz"},
				"5": 555, "6": 666,
			},
			"c": "ccc",
		},
		"3": 333,
	}

	xdict.Merge(d, updt)
	if reflect.DeepEqual(d, validation) {
		t.Logf("VALID")
	} else {
		t.Errorf("FAILED merge --- got: %#v", d)
	}
}

func Test_Delete(t *testing.T) {
	d := DictAlt{
		"k1": "some value",
		"k2": map[string]any{"1": "111", "2": true, "3": 333},
		"k3": []int{1, 2, 3},
		"k4": DictAlias{
			"a": []any{"1", 2, false, nil},
			"b": map[string]any{
				"1": 111, "2": "222", "3": nil,
				"4": DictAlt{"x": "xxx", "y": nil, "z": "zzz"},
				"5": DictAlias{"x": "xxx", "y": nil, "z": "zzz"},
			},
		},
	}
	t.Logf(">> input: %v", d)

	validations := []string{
		"k3.2.xxx", "k4.b.3", "k4.b.4",
	}
	for _, k := range validations {
		xdict.Delete(d, k)
		t.Logf(">> Delete %v\n--- %#v", k, d)
		res := xdict.Get(d, k, nil)
		if reflect.DeepEqual(res, nil) {
			t.Logf("VALID")
		} else {
			t.Errorf("FAILED check for key: %v", k)
		}
	}
}
