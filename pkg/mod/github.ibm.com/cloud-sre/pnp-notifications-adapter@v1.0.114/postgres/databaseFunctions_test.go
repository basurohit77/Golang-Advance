package postgres

import "testing"

func TestDBFuncs(t *testing.T) {

	SetupDBFunctions()
	SetupDBFunctionsForUT(nil, nil, nil, nil, nil, nil, nil)
}
