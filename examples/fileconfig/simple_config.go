package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/exonlabs/go-utils/pkg/fileconfig"
)

func _print(name string, data any) {
	fmt.Printf(name)
	b, _ := json.MarshalIndent(data, "", " ")
	fmt.Println(string(b))
}

var CFGFILE = path.Join(os.TempDir(), "sample_config")
var DEFAULTS = map[string]any{
	"key1": "some value",
	"key2": map[string]string{
		"x": "xxx",
		"y": "yyy",
		"z": "zzz",
	},
	"key3": []int{1, 2, 3},
	"key4": map[string]any{
		"a": []int{1, 2, 3},
		"b": map[string]any{
			"1": 111,
			"2": 222,
			"3": map[string]string{
				"x": "xxx",
				"y": "yyy",
				"z": "zzz",
			},
		},
	},
	"key7":  "عربي",
	"key8":  "عربي",
	"دليل1": "عربي",
	"دليل2": "عربي",
}

func main() {
	init := flag.Bool("init", false, "initialize config file")
	blobcfg := flag.Bool("blobcfg", false, "use binary config files mode")
	flag.Parse()

	var cfg fileconfig.FileConfig
	var err error

	if *blobcfg {
		// cfg, err = fileconfig.NewBlobConfig(CFGFILE, DEFAULTS)
		// if err != nil {
		// 	log.Panic(err)
		// }
	} else {
		cfg, err = fileconfig.NewJsonConfig(CFGFILE, DEFAULTS)
		if err != nil {
			log.Panic(err)
		}
	}

	if *init {
		fmt.Println("\n* using cfg file:", CFGFILE)
		fmt.Println()
		_print("\n-- default config:\n", DEFAULTS)

		cfg.Set("key4.b.4", []int{4, 44, 444}, true)
		for _, jKey := range []string{"key2.y", "key3"} {
			val, err := cfg.Get(jKey, false)
			if err != nil {
				log.Println(err)
			}
			fmt.Println("key", jKey)
			fmt.Println("val", val)
			cfg.Set(jKey, val, true)
		}
		cfg.Save()

		fmt.Println("\n-- config saved")
		fmt.Println(strings.Repeat("-", 50))
		d, err := (cfg.Dump())
		if err != nil {
			log.Println(err)
		}
		fmt.Println(string(d))
		fmt.Println(strings.Repeat("-", 50))
	} else {
		if *blobcfg {
			// cfg, err = fileconfig.NewBlobConfig(CFGFILE, nil)
			// if err != nil {
			// 	log.Panic(err)
			// }
		} else {
			cfg, err = fileconfig.NewJsonConfig(CFGFILE, nil)
			if err != nil {
				log.Panic(err)
			}
		}
	}
	cfg.Set("new_key", 999, false)
	cfg.Set("key4.b.دليل", "vvv", false)
	cfg.Set("key4.b.3.t", "ttt", true)
	cfg.Delete("key1")

	fmt.Printf("\n-- active config:\n")
	fmt.Printf("%T\n", cfg)
	for _, jKey := range cfg.Keys() {
		val, err := cfg.Get(jKey, false)
		if err != nil && err != fileconfig.ErrValNotExists {
			log.Println(err)
		}
		_print(jKey+": ", val)
	}
	fmt.Println()

	_print("-- Buffer --\n", cfg.Buffer())

	fmt.Println("\n-- read encoded keys")
	for _, jKey := range []string{"key2.y", "key3", "key4.b.4", "key4.b.3.t"} {
		val, err := cfg.Get(jKey, true)
		if err != nil && err != fileconfig.ErrValNotExists {
			log.Println(err)
		}
		fmt.Println(jKey, "=", val)
	}

	fmt.Println()

	cfg.Purge()
}
