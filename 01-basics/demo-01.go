package main

import "fmt"

func main() {
	/* 1. Functions can be assigned to variables */
	/*
		var fn = func() {
			fmt.Println("fn is invoked")
		}
	*/
	/*
		fn := func() {
			fmt.Println("fn is invoked")
		}
	*/

	var fn func()
	fn = func() {
		fmt.Println("fn is invoked")
	}
	fn()

	/*
		add := func(x, y int) int {
			return x + y
		}
	*/

	var add func(int, int) int
	add = func(x, y int) int {
		return x + y
	}
	fmt.Println(add(100, 200))
	//fmt.Println(add(100, 200))

	/* 2. Functions can be passed as arguments to other functions */
	exec(fn)
	execOper(add, 100, 200)

	/* 3. Functions can be retured from other function */
	adderFor100 := getAdderFor(100)
	fmt.Println(adderFor100(200))
	fmt.Println(adderFor100(300))
	fmt.Println(adderFor100(300))

	fmt.Println("closure")
	/* 4. Closures */
	increment := incrementor()
	fmt.Println(increment())
	fmt.Println(increment())
	fmt.Println(increment())
	fmt.Println(increment())

	/* Anonymous Function */
	func() {
		fmt.Println("Anonymous function invoked")
	}()
}

func exec(fn func()) {
	fn()
}

func execOper(operFn func(int, int) int, x, y int) {
	fmt.Println(operFn(x, y))
}

func getAdderFor(x int) func(int) int {
	return func(y int) int {
		return x + y
	}
}

/*
func getHttpClient(baseUrl string){
	return func(url, reqType, payload){

	}
}

httpClient(url, reqType, payload)

GET 	http://myNewService.com/products
GET 	http://myNewService.com/products/1
POST 	http://myNewService.com/products
PUT 	http://myNewService.com/products/1
DELETE 	http://myNewService.com/products/1
*/

/* Closures */
func incrementor() func() int {
	var count = 0
	return func() int {
		count++
		return count
	}
}

//curl -X GET http://localhost:8080/v1/catalog_sync/missing_emails -H "Authorization: Bearer eyJraWQiOiIyMDIyMDQxNjA4MjQiLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJJQk1pZC01NTAwMDVDWE5TIiwiaWQiOiJJQk1pZC01NTAwMDVDWE5TIiwicmVhbG1pZCI6IklCTWlkIiwic2Vzc2lvbl9pZCI6IkMtMGNkYTZiMmEtMDU0Ny00MGRmLWIxYmQtMjM4YWI4NmU1OTRlIiwic2Vzc2lvbl9leHBfbWF4IjoxNjUxMjA1NDE4LCJzZXNzaW9uX2V4cF9uZXh0IjoxNjUxMTI2NTQ4LCJqdGkiOiI0YmZhYTA3Ny0zNmFhLTQ5ZWQtYmE4NS0zNmMyMGQwZDc5ZTkiLCJpZGVudGlmaWVyIjoiNTUwMDA1Q1hOUyIsImdpdmVuX25hbWUiOiJSb2hpdCIsImZhbWlseV9uYW1lIjoiQmFzdSIsIm5hbWUiOiJSb2hpdCBCYXN1IiwiZW1haWwiOiJyb2hiYXMxMUBpbi5pYm0uY29tIiwic3ViIjoicm9oYmFzMTFAaW4uaWJtLmNvbSIsImF1dGhuIjp7InN1YiI6InJvaGJhczExQGluLmlibS5jb20iLCJpYW1faWQiOiJJQk1pZC01NTAwMDVDWE5TIiwibmFtZSI6IlJvaGl0IEJhc3UiLCJnaXZlbl9uYW1lIjoiUm9oaXQiLCJmYW1pbHlfbmFtZSI6IkJhc3UiLCJlbWFpbCI6InJvaGJhczExQGluLmlibS5jb20ifSwiYWNjb3VudCI6eyJib3VuZGFyeSI6Imdsb2JhbCIsInZhbGlkIjp0cnVlLCJic3MiOiI1YjdhNmY5NDZkMmY0OWZjYjM4NDUxY2NmZTViMjViNiIsImltc191c2VyX2lkIjoiODg2MzY2NCIsImZyb3plbiI6dHJ1ZSwiaW1zIjoiMjExNzUzOCJ9LCJpYXQiOjE2NTExMTkzNDgsImV4cCI6MTY1MTEyMDU0OCwiaXNzIjoiaHR0cHM6Ly9pYW0uY2xvdWQuaWJtLmNvbS9pZGVudGl0eSIsImdyYW50X3R5cGUiOiJ1cm46aWJtOnBhcmFtczpvYXV0aDpncmFudC10eXBlOmFwaWtleSIsInNjb3BlIjoiaWJtIG9wZW5pZCIsImNsaWVudF9pZCI6ImJ4IiwiYWNyIjozLCJhbXIiOlsidG90cCIsImNvb2tpZSIsIm1mYSIsIm90cCJdfQ.DGnDmO47HdhQDfccDhp2VYfa1tWJEckh_MxRRDZcz345WWeLxSRVWFFwyJbm0l670l80KAh5mJ0fmKK5QtWxdiCJ5mwG3WQyjQ83QUJXUGSTEGo-etEHRuAGmPqLSxBBEkSmtTe5siynCOORbKJYvjFgubBpSN-pFhoDHx_H55ot8Cgi_ZnKWYfAXtLKmF2tlPXSjlQ6pl1Lvv-6drAb6ZY5TR4LvYdkQ2NY-hbVk77Wx5SmYnAH3MegJqyQgU8PoaDTj4CJknWF7ElcEQ45beSeYSf6RwfqObMQwL8XERVXS3yNPe3fz-bUcceqMQFc5DWAEVrqID-mqTYjQpztpQ"
