package util

import (
	"fmt"
	"reflect"

	"k8s.io/klog/v2"
)

type validator interface {
	Validate() error
}

func ValidateFields(obj interface{}) error {
	rv := reflect.Indirect(reflect.ValueOf(obj))
	if rv.Type().Kind() != reflect.Struct {
		return fmt.Errorf("kind of %s expect a struct, got %s",
			reflect.TypeOf(obj).Name(), reflect.TypeOf(obj).Kind())
	}

	for i := 0; i < rv.Type().NumField(); i++ {
		if rv2 := reflect.Indirect(rv.Field(i)); rv2.CanAddr() && rv2.CanInterface() {
			switch rv2.Kind() {
			case reflect.Struct:
				klog.Infof("name %s kind %s", rv2.Type().Name(), rv2.Type().Kind())
				if v, ok := rv2.Addr().Interface().(validator); ok && v != nil {
					if err := v.Validate(); err != nil {
						return err
					}
				}
			case reflect.Array, reflect.Slice:
				for j := 0; j < rv2.Len(); j++ {
					rv3 := reflect.Indirect(rv2.Index(j))
					if !rv3.IsValid() {
						continue
					}
					if !rv3.CanAddr() || !rv3.CanInterface() {
						break
					}
					if v, ok := rv3.Addr().Interface().(validator); !ok {
						break
					} else {
						if v != nil {
							if err := v.Validate(); err != nil {
								return err
							}
						}
					}

				}
			}
		}
	}

	return nil
}
