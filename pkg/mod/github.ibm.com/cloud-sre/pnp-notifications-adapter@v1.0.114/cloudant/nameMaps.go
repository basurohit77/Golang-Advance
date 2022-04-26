package cloudant

import (
	"log"
)

// NameMap carries the name maps from cloudant for notification category ID to service name
type NameMap struct {
	//Maps []*NameLookup
	CategoryMap map[string]*NameComponent
}

// NameLookup is a cache item corresponding to one of the cloudant name mapping databases
// We rearrange things a bit and create hashmaps keyed by the notification category ID
type NameLookup struct {
	CategoryMap map[string]*NameComponent
}

// NewNameMap creates a new name map
func NewNameMap(id, pw string, mappingURLs ...string) (nameMap *NameMap, err error) {

	nameMap = new(NameMap)
	//nameMap.Maps = make([]*NameLookup, 0, len(mappingURLs))
	nameMap.CategoryMap = make(map[string]*NameComponent)

	for _, u := range mappingURLs {

		//nl := new(NameLookup)
		//nl.CategoryMap = make(map[string]*NameComponent)

		myMap, err := GetNameMapping(u, id, pw)
		if err != nil {
			return nil, err
		}

		for _, p := range myMap.Components {
			//fmt.Println("Saving, ", p.ID)
			//nl.CategoryMap[p.ID] = p
			if nameMap.CategoryMap[p.ID] != nil {
				log.Println("ERROR: unexpected duplicate notification category ID: ", p.ID)
			}
			nameMap.CategoryMap[p.ID] = p
		}

		//nameMap.Maps = append(nameMap.Maps, nl)
	}

	return nameMap, nil
}

// MatchCategoryID will take a category ID and return a struct containing names about the service.
func (nm *NameMap) MatchCategoryID(categoryID string) (result *NameComponent) {

	if nm != nil {
		result = nm.CategoryMap[categoryID]
	}

	return result
}
