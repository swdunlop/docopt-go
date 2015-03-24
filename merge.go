// Licensed under terms of MIT license (see LICENSE-MIT)
// Copyright (c) 2013 Keith Batten, kbatten@gmail.com
// Contributed by Scott W. Dunlop, swdunlop@gmail.com

package docopt

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

/*
Merge the results of a previous Parse into `dst` using Go reflection.

Given a pointer to a structure, `dst`, and a the results of a previous Parse, `src`, docopt will merge
values from `src` into the fields of `dst`.  Each public field in `dst` is examined using Go's `reflect` package
and updated unless it has been tagged `docopt:"-"`.

Merge selects a value from `src` for each field in `dst`, based on the following:

- If the field is private, it is skipped.  (Go reflection won't indicate it exists, anyway.)

- If the docopt tag is "-", it is skipped.

- If the field has an all uppercase name, it will be updated; e.g. `FILES` will use `src["FILES"]`

- If there is a docopt tag, it is used as the key for `dst`; e.g. `docopt:"-d"` will use `src["-d"]`

- Otherwise, docopt will ignore the field.

Given a value selected from `src` based on the rules above, Merge will update the field based on the field's type.
If the type is not supported, Merge will panic, unless it has been bypassed using the `docopt:"-"` tag.
The following field types are currently supported:

- Integers and Floating Point numbers are parsed using `encoding/json` and updated.

- Boolean fields are updated based on the presence or absence of a flag.

- Strings are updated with the value provided by dst.

- Slices of any of the previous types will be updated with the values found in dst.

- Merger implementations will be permitted to Merge themselves.  (Slices of Mergers are not currently supported.)

The following example defines bindings for `-j`, `-w` and `-n` flags, and accepts zero or more URL values:

	var opt struct {
		Home 	  string `docopt:"-"`  // not updated by Merge
		JsonInput bool   `docopt:"-j"`
		WriteFile string `docopt:"-w"`
		N         int    `docopt:"-n"`
		URL       []string
	}

If a value in `dst` does not have a field associated with it in `src`, it is silently ignored.
However, if a field in `dst` does not have a corresponding value in `src`, a panic is produced,
since the type is no longer consistent with the documented command line interface.
*/
func Merge(dst interface{}, src map[string]interface{}) error {
	dv := reflect.Indirect(reflect.ValueOf(dst))
	dt := dv.Type()
	nf := dt.NumField()
	for i := 0; i < nf; i++ {
		ft := dt.Field(i)
		tag := ft.Tag.Get("docopt")
		switch {
		case tag == "-":
			continue
		case tag != "":
			// okay.
		case strings.ToUpper(ft.Name) == ft.Name:
			tag = ft.Name
		default:
			continue
		}
		val, ok := src[tag]
		if !ok {
			panic(fmt.Errorf("value %#v not defined in documentation", tag))
		}
		val = val

		fv := dv.Field(i).Addr().Interface()
		fv = fv
		// fmt.Printf(".. for field %v (type %T), docopt provides %#v\n", ft.Name, fv, val)

		var err error
		switch fv := fv.(type) {
		case *string:
			switch val := val.(type) {
			case string:
				*fv = val
			case []string:
				switch len(val) {
				case 0:
				case 1:
					*fv = val[0]
				default:
					err = fmt.Errorf("too many values")
				}
			}

		case *[]string:
			switch val := val.(type) {
			case string:
				*fv = []string{val}
			case []string:
				*fv = val
			}

		case *bool:
			switch val := val.(type) {
			case bool:
				*fv = val
			default:
				panic(fmt.Errorf("expected bool for %v, got %#v", tag, val))
			}

		case *[]int, *[]int32, *[]int64, *[]float32, *[]float64:
			js := ""
			switch val := val.(type) {
			case string:
				js = val
			case []string:
				js = "[" + strings.Join(val, ",") + "]"
			}
			if js != "" {
				err = json.Unmarshal([]byte(js), fv)
			}

		case *int, *int32, *int64, *float32, *float64:
			js := ""
			switch val := val.(type) {
			case string:
				js = val
			case []string:
				switch len(val) {
				case 0:
				case 1:
					js = val[0]
				default:
					err = fmt.Errorf("too many values")
				}
			}
			if js != "" {
				err = json.Unmarshal([]byte(js), fv)
			}

		case Merger:
			err = fv.MergeDocopt(val)

			//TODO(scott): support []Merger

		default:
			panic(fmt.Errorf("field %#v not supported by docopt", ft.Name))
		}

		if err != nil {
			return fmt.Errorf("%v: %v", tag, err.Error())
		}
	}
	return nil
}

/*
Merger indicates fields that know how to merge a docopt flag or argument value for Merge.

Implementations of Merger should generally use pointer or slice types as the recipient, so they
can update themselves with the provided value.

Example:

	// Base16 is a hexadecimal encoded string
	type Base16 string
	...
	func (b16 *Base16) MergeDocopt(v interface{}) error {
		str, ok := v.(string)
		if !ok {
			return fmt.Errorf("expected base16 string, got %v", v)
		}
		p, err = hex.DecodeString(str)
		*b16 = string(p)
		return err
	}

*/
type Merger interface {
	MergeDocopt(v interface{}) error
}
