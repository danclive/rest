package main

import (
	"fmt"

	"github.com/danclive/rest"
)

var Rest *rest.Rest

func main2() {

	Rest = rest.NewRest()

	/*
		rest.SetBefore(func(req rest.Request) {

		})

		rest.SetAfter(func(rep rest.Response) {

		})

		rest.SetBaseUrl("http://xxx.xxx.xxx")

		response, err := rest.Post("http://httpbin.org/post").
			Header("key", "value").
			Query("key", "value").
			Querys(map[string]interface{}).
			Param("key", "value").
			Params(map[string]interface{}).
			File("key", "file.text").
			FileBytes("key", "file.text", []byte("file")).
			Auth("username", "password").
			Body([]byte("a")).
			Send()

		if err != nil {
			panic(err)
		}

		response.Body()
		response.Code()
		response.Header("key")
		response.Headers()

		response.Bind(interface{})
	*/
}

/*
type User struct {
	Id        int64  `query:"aaa"`
	Username  string `query:"bbb"`
	Password  string
	Logintime time.Time
}

func Struct2Map(obj interface{}) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		fmt.Println(v.Field(i).Tag.)
		data[t.Field(i).Name] = v.Field(i).Interface()
	}
	return data
}

func main() {
	user := User{5, "zhangsan", "pwd", time.Now()}
	data := Struct2Map(user)
	fmt.Println(data)
}
*/
/*
const tagName = "rest"

type User struct {
	Id        int    `rest:"id"`
	Name      string `rest:"name"`
	Email     string `rest:"email"`
	Logintime time.Time
}

func main() {
	user := User{
		Id:        1,
		Name:      "John Doe",
		Email:     "john@example",
		Logintime: time.Now(),
	}

	// TypeOf returns the reflection Type that represents the dynamic type of variable.
	// If variable is a nil interface value, TypeOf returns nil.
	t := reflect.TypeOf(user)
	v := reflect.ValueOf(user)

	//Get the type and kind of our user variable
	// fmt.Println("Type: ", t.Name())
	// fmt.Println("Kind: ", t.Kind())

	var data = make(map[string]string)

	for i := 0; i < t.NumField(); i++ {
		// Get the field, returns https://golang.org/pkg/reflect/#StructField
		field := t.Field(i)

		//Get the field tag value
		tag := field.Tag.Get(tagName)

		var key string

		if tag == "" {
			key = field.Name
		} else if tag == "-" {
			continue
		} else {
			key = tag
		}

		data[key] = fmt.Sprintf("%v", v.Field(i))

		// fmt.Printf("%d. %v(%v), tag:'%v'\n", i+1, field.Name, field.Type.Name(), tag)
		// fmt.Println(fmt.Sprintf("%v", v.Field(i)))
	}

	fmt.Println(data)
}
*/

func main() {
	Rest = rest.NewRest().BaseUrl("https://main.danclive.com")

	res, err := Rest.Get("/article").Query("page", "1").Query("per_page", "12").Send()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(res.Body()))

	response := Response{}

	fmt.Println(res.Json(&response))

	fmt.Println(response)
}

type Response struct {
	Success bool    `json:"success"`
	Message Message `json:"message"`
	Data    Data    `json:"data"`
}

type Message struct {
	Code  int `json:"success"`
	Error string
}

type Data struct {
	Articles []Article `json:"articles"`
}

type Article struct {
	Id    string   `json:"id"`
	Title string   `json:"title"`
	Image []string `json:"image"`
	//Summary string   `json:"summary"`
}
