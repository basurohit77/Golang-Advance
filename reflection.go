// Example program to show difference between
// Type and Kind and to demonstrate use of
// Methods provided by Go reflect Package
// Method used to find the variable name of inner package of a sturct
// A use case to avoid:: panic: reflect.Value.Interface: cannot return value obtained from unexported field or method
package main

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

type Person1 struct {
	W3ID string
	Name string
}

type Address1 struct {
	city    string
	country string
}

type User1 struct {
	name      string
	age       int
	address   Address1
	manager   Person1
	developer Person1
	tech      Person1
}

func showDetails(load, email interface{}) {
	//fmt.Println("The values in the first argument are :")
	if reflect.ValueOf(load).Kind() == reflect.Struct {
		typ := reflect.TypeOf(load)
		value := reflect.ValueOf(load)
		value2 := reflect.New(value.Type()).Elem() // #1 For struct, not addressable create a copy With Elements.
		value2.Set(value)                          //#2 Value2 is addressable and can be set
		for i := 0; i < typ.NumField(); i++ {
			//fmt.Println(value.Field(i).Kind(), "---------")
			if value.Field(i).Kind() == reflect.Struct {
				rf := value2.Field(i)
				rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
				irf := rf.Interface() 
				typrf := reflect.TypeOf(irf)
				nameP := typrf.String()
				if strings.Contains(nameP, "Person") {
					//fmt.Println(nameP, "FOUND !!!!!!! ")
					for j := 0; j < typrf.NumField(); j++ {
						//fmt.Println(j)
						re := rf.Field(j)
						nameW := typrf.Field(j).Name
						//fmt.Println(nameW)
						if strings.Contains(nameW, "W3ID") {
							//fmt.Println(nameW)
							valueW := re.Interface()
							fetchEmail := valueW.(string)
							//fmt.Println(fetchEmail)
							if fetchEmail == email {
								fmt.Println(fetchEmail, " MATCH!!!!")
							}
						}
					}
				}
				//fmt.Println(rf.Interface(), "++=====")
				showDetails(irf, email)
			} else {
				// fmt.Printf("%d.Type:%T || Value:%#v\n",
				// 	(i + 1), value.Field(i), value.Field(i))

				// fmt.Println("Kind is ", value.Field(i).Kind())
			}
		}
	}
}

func main() {
	iD := "tsumi@in.org.com"

	load := User1{
		name: "John Doe",
		age:  34,
		address: Address1{
			city:    "New York",
			country: "USA",
		},
		manager: Person1{
			W3ID: "jBult@in.org.com",
			Name: "Bualt",
		},
		developer: Person1{
			W3ID: "tsumi@in.org.com",
			Name: "Sumi",
		},
		tech: Person1{
			W3ID: "lPaul@in.org.com",
			Name: "Paul",
		},
	}

	showDetails(load, iD)
}
