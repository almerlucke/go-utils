// Package unmarshal can be used to unmarshal httprouter params, query params
// and JSON body params together in one data structure
package unmarshal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/almerlucke/go-utils/reflection/structural"

	"github.com/julienschmidt/httprouter"
)

func addRouterParamsToMap(pm httprouter.Params, mp map[string]string) {
	for _, param := range pm {
		mp[param.Key] = param.Value
	}
}

func addQueryParamsToMap(values url.Values, mp map[string]string) {
	for key, value := range values {
		if len(value) > 0 {
			mp[key] = value[0]
		}
	}
}

// setFieldValue - convert param value to reflect.Value
func setFieldValue(paramValue string, field reflect.Value) error {
	switch field.Kind() {
	case reflect.Int:
		intValue, err := strconv.ParseInt(paramValue, 10, strconv.IntSize)
		if err != nil {
			return err
		}

		field.SetInt(intValue)
	case reflect.Int8:
		intValue, err := strconv.ParseInt(paramValue, 10, 8)
		if err != nil {
			return err
		}

		field.SetInt(intValue)
	case reflect.Int16:
		intValue, err := strconv.ParseInt(paramValue, 10, 16)
		if err != nil {
			return err
		}

		field.SetInt(intValue)
	case reflect.Int32:
		intValue, err := strconv.ParseInt(paramValue, 10, 32)
		if err != nil {
			return err
		}

		field.SetInt(intValue)
	case reflect.Int64:
		intValue, err := strconv.ParseInt(paramValue, 10, 64)
		if err != nil {
			return err
		}

		field.SetInt(intValue)
	case reflect.Uint:
		uintValue, err := strconv.ParseUint(paramValue, 10, strconv.IntSize)
		if err != nil {
			return err
		}

		field.SetUint(uintValue)
	case reflect.Uint8:
		uintValue, err := strconv.ParseUint(paramValue, 10, 8)
		if err != nil {
			return err
		}

		field.SetUint(uintValue)
	case reflect.Uint16:
		uintValue, err := strconv.ParseUint(paramValue, 10, 16)
		if err != nil {
			return err
		}

		field.SetUint(uintValue)
	case reflect.Uint32:
		uintValue, err := strconv.ParseUint(paramValue, 10, 32)
		if err != nil {
			return err
		}

		field.SetUint(uintValue)
	case reflect.Uint64:
		uintValue, err := strconv.ParseUint(paramValue, 10, 64)
		if err != nil {
			return err
		}

		field.SetUint(uintValue)
	case reflect.Float32:
		floatValue, err := strconv.ParseFloat(paramValue, 32)
		if err != nil {
			return err
		}

		field.SetFloat(floatValue)
	case reflect.Float64:
		floatValue, err := strconv.ParseFloat(paramValue, 64)
		if err != nil {
			return err
		}

		field.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(paramValue)
		if err != nil {
			return err
		}

		field.SetBool(boolValue)
	case reflect.String:
		field.SetString(paramValue)

	default:
		return fmt.Errorf("Unsupported request value type %v", field.Type())
	}

	return nil
}

// unmarshalParamsMap unmarshals params map to object structure fields
func unmarshalParamsMap(paramsMap map[string]string, obj interface{}) error {
	desc, ok := structural.NewStructDescriptor(obj)
	if !ok {
		return errors.New("object is not a struct or struct ptr")
	}

	if !desc.CanSet() {
		return errors.New("object fields can not be set")
	}

	err := desc.ScanFields(true, true, nil, func(field structural.FieldDescriptor, context interface{}) error {
		fieldName := field.Name()
		fieldTag := field.Tag().Get("param")
		lowercaseFieldName := strings.ToLower(fieldName)

		for key, value := range paramsMap {
			if strings.ToLower(key) == lowercaseFieldName || key == fieldTag {
				err := setFieldValue(value, field.Value())
				if err != nil {
					return err
				}

				break
			}
		}

		return nil
	})

	return err
}

// unmarshalRequestParams unmarshal request query and router params to obj
func unmarshalParams(r *http.Request, pm httprouter.Params, obj interface{}) error {
	// Param map
	mp := make(map[string]string)

	// Add query params
	addQueryParamsToMap(r.URL.Query(), mp)

	// Add router params
	addRouterParamsToMap(pm, mp)

	// Unmarshal query and router params
	return unmarshalParamsMap(mp, obj)
}

// Unmarshal query params, httprouter params and optional JSON body to object.
// Object needs to be a structure
func Unmarshal(r *http.Request, pm httprouter.Params, decodeBody bool, obj interface{}) error {
	var err error

	// Check if we need to decode the request JSON body (POST or PUT)
	if decodeBody {
		// Start decoding json body
		decoder := json.NewDecoder(r.Body)

		// Always close body
		defer r.Body.Close()

		// Decode to object
		err = decoder.Decode(obj)
		if err != nil && err != io.EOF {
			return err
		}
	}

	// Unmarshal query & router params
	err = unmarshalParams(r, pm, obj)
	if err != nil {
		return err
	}

	return nil
}
