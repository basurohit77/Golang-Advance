package main

import (
	"reflect"
)

// StructIterator is an iterator over the fields of a struct
type StructIterator struct {
	data  reflect.Value
	index int
}

// Next returns the next field from a StructIterator, as (name,value) pair (or "",nil) if there is no next field
func (it *StructIterator) Next() (name string, value interface{}) {
	if it.index >= it.data.Type().NumField() {
		return "", nil
	}
	ft := it.data.Type().Field(it.index)
	fv := it.data.Field(it.index)
	it.index++
	return ft.Name, fv.Interface()
}

// Slice returns all the Elements from this iterator in a slice
func (it *StructIterator) Slice() []Element {
	t := it.data.Type()
	len := t.NumField()
	result := make([]Element, len)
	for i := 0; i < len; i++ {
		result[i] = Element{
			Name:  t.Field(i).Name,
			Value: it.data.Field(i),
		}
	}
	return result
}

// NewStructIterator creates a new StructIterator for the given struct
func NewStructIterator(data interface{}) *StructIterator {
	it := StructIterator{}
	v := reflect.ValueOf(data)
	if v.Type().Kind() == reflect.Ptr {
		it.data = v.Elem()
	} else {
		it.data = v
	}
	return &it
}

// Element is one element of a slice containing all the values from the iterator
type Element struct {
	Name  string
	Value interface{}
}
