// Copyright 2012, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*  Filename:    validate.go
 *  Author:      Bryan Matsuo <bryan.matsuo [at] gmail.com>
 *  Created:     2012-07-03 23:27:22.212922 -0700 PDT
 *  Description: Main source file in confighelper
 */

// Package validate helps with validation.
package validate

import (
	"errors"
	"fmt"
)

// The interface that validatable types should satisfy.
type Interface interface {
	Validate() error
}

// Call Validate() on v if v is validatable.
func V(v interface{}) error {
	switch v.(type) {
	case Interface:
		return v.(Interface).Validate()
	}
	return nil
}

// Validate property values.
//		type Qux int
//		func (qux Qux) Validate() error { return errors.New("qux") }
//		type Foo struct { Bar Qux }
//      func (foo *Foo) Validate() error {
//			return validator.Property("Bar", foo.Bar)
//		}
//		validator.V(&Foo{1}) // `Bar: qux`
func Property(property string, value interface{}) error {
	return PropertyFunc(property, func() error { return V(value) })
}

// Used in tricker validation cases.
// 
// Try to use Property() instead.
//
//		type Qux int
//		func (qux Qux) Validate() error { return errors.New("qux") }
//		type Foo struct { Bars []struct { Baz Qux } }
//		func (foo *Foo) Validate() error {
//			return validator.PropertyFunc("Bars", func() (err error) {
//				for i, bar := range foo.Bars {
//					if err = validator.IndexFunc(i, func() error {
//						return validator.Property("Baz", bar.Baz)
//					})
//					if err != nil {
//						return
//					}
//				}
//				return
//			})
//		}
//		validator.V(&Foo{[]struct{Baz string}{"qux"}}) // `Bar[0].Baz: qux`
func PropertyFunc(property interface{}, validate func() error) error {
	if err := validate(); err != nil {
		return PropertyError{fmt.Sprint(property), nil, err}
	}
	return nil
}

// A validation error from by a (possibly nested) property.
type PropertyError struct {
	property string
	index    interface{}
	err      error
}

// The validation error.
func (err PropertyError) OriginatingError() error {
	switch err.err.(type) {
	case PropertyError:
		return err.err.(PropertyError).OriginatingError()
	}
	return err.err
}

// The name of the invalid property.
func (err PropertyError) Property() string {
	prefix := err.property
	if err.index != nil {
		prefix = fmt.Sprintf("%s[%#v]", prefix, err.index)
	}
	switch err.err.(type) {
	case PropertyError:
		return fmt.Sprintf("%s.%s", prefix, err.property)
	}
	return prefix
}

// The invalid property concatenated with the validation error message.
func (err PropertyError) Error() string {
	prefix := err.property
	if err.index != nil {
		prefix = fmt.Sprintf("%s[%#v]", prefix, err.index)
	}
	switch err.err.(type) {
	case PropertyError:
		if err.err.(PropertyError).property == "" {
			return fmt.Sprintf("%s%v", prefix, err.err)
		}
		return fmt.Sprintf("%s.%v", prefix, err.err)
	}
	return fmt.Sprintf("%s: %v", prefix, err.err)
}

// Validate property element values values (see Property).
func Index(index, value interface{}) error {
	return IndexFunc(index, func() error {
		if err := V(value); err != nil {
			return err
		}
		return nil
	})
}

// Used in tricker cases of validating elements of properties that are slices/maps
//
// Try to use Index() instead.
//
//		type Qux int
//		func (qux Qux) Validate() error { return errors.New("qux") }
//		type Foo struct { Bars []struct { Baz struct { Qux } } }
//		func (foo *Foo) Validate() error {
//			return validator.PropertyFunc("Bars", func(k *interface{}) error {
//				for i, bar := range foo.Bars {
//					err := validator.IndexFunc(i, func() error {
//						return validator.Property("Baz", bar.Baz)
//					})
//					if err != nil {
//						return err
//					}
//				}
//				return nil
//			})
//		}
//		validator.V(&Foo{[]struct{Baz string}{"qux"}}) // `Bar[0].Baz: qux`
func IndexFunc(index interface{}, validate func() error) (err error) {
	if err = validate(); err != nil {
		return PropertyError{"", index, err}
	}
	return nil
}

// An error describing an invalid value.
//		Invalid("foo")               // `Invalid: "foo"`
//		Invalid("foo", "bar")        // `Invalid foo: "bar"`
//		Invalid("foo", "bar", "baz") // `Invalid foo bar: "baz"`
//		...
func Invalid(v ...interface{}) error {
	prefix, size := "Invalid", len(v)
	switch {
	case size > 1:
		prefix = fmt.Sprint(prefix, v[:size-1])
		fallthrough
	case size == 1:
		return fmt.Errorf("%s: %#v", prefix, v[:size-1])
	}
	return errors.New(prefix)
}
