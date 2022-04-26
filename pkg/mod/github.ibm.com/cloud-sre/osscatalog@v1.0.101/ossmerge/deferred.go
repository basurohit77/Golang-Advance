package ossmerge

// Utilities for running deferred functions at the very of a merge, for example to remove or void some records

var deferredFunctions []func()

// registerDeferredFunction registers one function (closure) to be executed by RunAllDeferredFunctions()
func registerDeferredFunction(f func()) {
	deferredFunctions = append(deferredFunctions, f)
}

// runAllDeferredFunctions runs all the functions (closures) that have been registered with registerDeferredFunction()
func runAllDeferredFunctions() {
	for _, f := range deferredFunctions {
		f()
	}
	deferredFunctions = nil
}
