package unpack

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// roundTripToStruct is our backstop in case mapToStruct fails
// in an edge case
func roundTripToStruct(o any, v any) error {
	b, err := json.Marshal(o)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, v); err != nil {
		return err
	}

	return nil
}

func mapToStruct(m map[string]any, s interface{}) error {
	structValue := reflect.ValueOf(s).Elem()
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		// Get the JSON tag name
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = fieldType.Name
		} else {
			// Handle "name,omitempty" format
			jsonTag = strings.Split(jsonTag, ",")[0]
		}

		if jsonTag == "-" {
			continue // Skip fields marked with json:"-"
		}

		// Get value from map
		if value, exists := m[jsonTag]; exists && field.CanSet() {
			if err := setFieldValue(field, value); err != nil {
				return fmt.Errorf("error setting field %s: %w", fieldType.Name, err)
			}
		}
	}

	return nil
}

func setFieldValue(field reflect.Value, value interface{}) error {
	valueReflect := reflect.ValueOf(value)
	fieldType := field.Type()

	// Handle nil values
	if value == nil {
		if field.Kind() == reflect.Ptr {
			field.Set(reflect.Zero(fieldType))
		}
		return nil
	}

	// Direct assignment if types match
	if valueReflect.Type() == fieldType {
		field.Set(valueReflect)
		return nil
	}

	// Handle convertible types (int to int64, etc.)
	if valueReflect.Type().ConvertibleTo(fieldType) {
		field.Set(valueReflect.Convert(fieldType))
		return nil
	}

	// Handle maps
	if fieldType.Kind() == reflect.Map && valueReflect.Kind() == reflect.Map {
		return setMapValue(field, valueReflect, fieldType)
	}

	// Handle slices
	if fieldType.Kind() == reflect.Slice && valueReflect.Kind() == reflect.Slice {
		return setSliceValue(field, valueReflect, fieldType)
	}

	// Handle structs (nested structures)
	if fieldType.Kind() == reflect.Struct && valueReflect.Kind() == reflect.Map {
		// Create new struct instance
		newStruct := reflect.New(fieldType).Interface()
		if sourceMap, ok := value.(map[string]any); ok {
			if err := mapToStruct(sourceMap, newStruct); err != nil {
				return err
			}
			field.Set(reflect.ValueOf(newStruct).Elem())
			return nil
		}
	}

	// Handle pointers to structs
	if fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Struct && valueReflect.Kind() == reflect.Map {
		newStruct := reflect.New(fieldType.Elem()).Interface()
		if sourceMap, ok := value.(map[string]any); ok {
			if err := mapToStruct(sourceMap, newStruct); err != nil {
				return err
			}
			field.Set(reflect.ValueOf(newStruct))
			return nil
		}
	}

	return fmt.Errorf("cannot convert %T to %s", value, fieldType)
}

func setMapValue(field reflect.Value, valueReflect reflect.Value, fieldType reflect.Type) error {
	// Create new map of the target type
	newMap := reflect.MakeMap(fieldType)

	keyType := fieldType.Key()
	valueType := fieldType.Elem()

	// Iterate over source map
	for _, key := range valueReflect.MapKeys() {
		sourceKey := key
		sourceValue := valueReflect.MapIndex(key)

		// Convert key
		var targetKey reflect.Value
		if sourceKey.Type().ConvertibleTo(keyType) {
			targetKey = sourceKey.Convert(keyType)
		} else {
			return fmt.Errorf("cannot convert map key from %s to %s", sourceKey.Type(), keyType)
		}

		// Convert value - extract concrete value from interface{}
		var targetValue reflect.Value
		sourceValueInterface := sourceValue.Interface()

		if sourceValueInterface == nil {
			targetValue = reflect.Zero(valueType)
		} else {
			// Extract the concrete value from interface{} if needed
			concreteValue := reflect.ValueOf(sourceValueInterface)

			if concreteValue.Type().ConvertibleTo(valueType) {
				targetValue = concreteValue.Convert(valueType)
			} else if valueType.Kind() == reflect.Map && concreteValue.Kind() == reflect.Map {
				// Nested map conversion
				targetValue = reflect.New(valueType).Elem()
				if err := setMapValue(targetValue, concreteValue, valueType); err != nil {
					return err
				}
			} else if valueType.Kind() == reflect.Slice && concreteValue.Kind() == reflect.Slice {
				// Nested slice conversion
				targetValue = reflect.New(valueType).Elem()
				if err := setSliceValue(targetValue, concreteValue, valueType); err != nil {
					return err
				}
			} else if valueType.Kind() == reflect.Struct && concreteValue.Kind() == reflect.Map {
				// Map to struct conversion
				targetValue = reflect.New(valueType).Elem()
				if sourceMap, ok := sourceValueInterface.(map[string]any); ok {
					if err := mapToStruct(sourceMap, targetValue.Addr().Interface()); err != nil {
						return err
					}
				} else {
					return fmt.Errorf("expected map[string]any for struct conversion, got %T", sourceValueInterface)
				}
			} else {
				return fmt.Errorf("cannot convert map value from %s to %s", concreteValue.Type(), valueType)
			}
		}

		newMap.SetMapIndex(targetKey, targetValue)
	}

	field.Set(newMap)
	return nil
}

func setSliceValue(field reflect.Value, valueReflect reflect.Value, fieldType reflect.Type) error {
	sourceLen := valueReflect.Len()

	// Create new slice
	newSlice := reflect.MakeSlice(fieldType, sourceLen, sourceLen)

	for i := 0; i < sourceLen; i++ {
		sourceElement := valueReflect.Index(i)
		targetElement := newSlice.Index(i)

		if err := setFieldValue(targetElement, sourceElement.Interface()); err != nil {
			return fmt.Errorf("error converting slice element at index %d: %w", i, err)
		}
	}

	field.Set(newSlice)
	return nil
}
