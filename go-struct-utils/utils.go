package gostructutils

import (
	"encoding/json"
	"reflect"
)

func StructToMapJSON(s any) (map[string]any, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func StructToMapUsingReflection(s any) map[string]any {
	result := make(map[string]any)
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		return result
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if !field.IsExported() {
			continue
		}

		key := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			if commaIdx := len(jsonTag); commaIdx > 0 {
				for j, char := range jsonTag {
					if char == ',' {
						commaIdx = j
						break
					}
				}
				key = jsonTag[:commaIdx]
			}
		}

		result[key] = value.Interface()
	}

	return result
}

func StructToMapUsingAdvancedReflection(s any) map[string]any {
	result := make(map[string]any)
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		return result
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		key := field.Name
		if jsonTag != "" {
			tagParts := []string{}
			current := ""
			for _, char := range jsonTag {
				if char == ',' {
					tagParts = append(tagParts, current)
					current = ""
				} else {
					current += string(char)
				}
			}
			if current != "" {
				tagParts = append(tagParts, current)
			}

			if len(tagParts) > 0 && tagParts[0] != "" {
				key = tagParts[0]
			}
		}

		result[key] = value.Interface()
	}

	return result
}
