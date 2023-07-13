package utils

import "reflect"

func ReplaceStructField(s interface{}, fieldName string, value interface{}) interface{} {
	newFields := make([]reflect.StructField, 0)
	sType := reflect.TypeOf(s)
	for i := 0; i < sType.NumField(); i++ {
		f := sType.Field(i)
		newType := f.Type
		if f.Name == fieldName {
			newType = reflect.TypeOf(value)
		}
		newFields = append(newFields, reflect.StructField{
			Name: f.Name,
			Type: newType,
			Tag:  f.Tag,
		})
	}
	typ := reflect.StructOf(newFields)
	v := reflect.New(typ).Elem()
	return v.Addr().Interface()
}
