// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
	"github.com/exonlabs/go-utils/pkg/jconfig"
)

var (
	CFGFILE = filepath.Join(os.TempDir(), "sample_config.json")
	SECRET  = "12345678"
)

// mixed Dict and map[string]any definitions
var DEFAULTS = map[string]any{
	"key1": "some value",
	"key2": map[string]any{
		"1": "xxx",
		"2": "yyy",
		"3": "zzz",
	},
	"key3": []int{1, 2, 3},
	"key4": dictx.Dict{
		"a": []int{1, 2, 3},
		"b": map[string]any{
			"1": 111,
			"2": 222,
			"3": dictx.Dict{
				"x": "xxx",
				"y": "yyy",
				"z": "zzz",
			},
			"4": nil,
		},
	},
	"key7":  "عربي 7",
	"key8":  "عربي 8",
	"دليل1": "عربي 1",
	"دليل2": "عربي 2",
}

func _print(msg string, data any) {
	fmt.Println(msg)
	b, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(b))
}

func NewConfig() *jconfig.JConfig {
	cfg, _ := jconfig.New(CFGFILE, DEFAULTS)
	cfg.InitAES128(SECRET)
	return cfg
}

func main() {
	init := flag.Bool("init", false, "initialize config file")
	flag.Parse()

	if *init {
		cfg := NewConfig()
		fmt.Println("\n* using cfg file:", CFGFILE)

		// add/set default secure data
		data := map[string]any{
			"key2.y":   cfg.Get("key2.y", nil),
			"key3":     cfg.Get("key3", nil),
			"key4.b.5": []any{1, "2", true, 1.2345},
			"key5":     nil,
		}
		for k, v := range data {
			if err := cfg.SetSecure(k, v); err != nil {
				panic(err)
			}
		}

		_print("\n-- initial config:", cfg.Buffer)

		fmt.Println("\n-- saving config")
		if err := cfg.Save(); err != nil {
			panic(err)
		}

		fmt.Println("\n" + strings.Repeat("-", 50))
		fmt.Println("file contents")
		fmt.Println(strings.Repeat("-", 50))
		if data, err := os.ReadFile(CFGFILE); err != nil {
			panic(err)
		} else {
			fmt.Println(string(data))
		}
		fmt.Println(strings.Repeat("-", 50))
		return
	}

	cfg := NewConfig()
	fmt.Println("\n* using cfg file:", CFGFILE)
	if cfg.IsExist() {
		if err := cfg.Load(); err != nil {
			panic(err)
		}
	}
	_print("\n-- loaded config:", cfg.Buffer)

	cfg.Set("new_key", 999)
	cfg.Set("key4.b.دليل", "vvv")
	cfg.Set("key4.b.3.t", "ttt")
	cfg.Delete("key1")
	_print("\n-- modified config:", cfg.Buffer)

	fmt.Println("\n-- config keys/values:")
	for _, k := range cfg.Keys() {
		fmt.Printf("%s = %v\n", k, cfg.Get(k, nil))
	}

	fmt.Println("\n-- read secured keys")
	keys := []string{
		"key2.y", "key3", "key4.b.5", "key5",
	}
	for _, k := range keys {
		if val, err := cfg.GetSecure(k, nil); err != nil {
			fmt.Printf("%s = err: %v\n", k, err.Error())
		} else {
			fmt.Printf("%s = %v\n", k, val)
		}
	}

	cfg.Purge()
	fmt.Println("\n-- config purged")

	fmt.Println()
}
