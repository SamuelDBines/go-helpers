# Filestore



# Usage

```

package main

import (
	"fmt"
	"log"

	"./filestore"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	fs := filestore.New("./data")

	_ = fs.WriteHTML("pages/index.html", "<h1>Hello</h1>")
	_ = fs.WriteYAML("config/app.yaml", "port: 3000\nname: demo")

	u := User{ID: "u1", Name: "Sam"}
	if err := fs.WriteJSON("db/users/u1.json", u, true); err != nil {
		log.Fatal(err)
	}

	var read User
	if err := fs.ReadJSON("db/users/u1.json", &read); err != nil {
		log.Fatal(err)
	}
	fmt.Println(read.Name)

	_ = fs.AppendString("logs/app.log", "started\n")

	ok := fs.Exists("db/users/u1.json")
	fmt.Println("exists:", ok)

	_ = fs.Delete("db/users/u1.json")
}
```