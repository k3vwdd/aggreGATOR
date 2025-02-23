package main

import (
	"fmt"
	"github.com/k3vwdd/aggreGATOR/internal/config"
)





func main() {
    config, err := config.Read()
    if err != nil {
        fmt.Println("error reading config:", err)
    }
    config.DBUrl = "postgres://example"
    config.SetUser("Pam")

    fmt.Println(config.DBUrl)
    fmt.Println(config.CurrentUserName)
}
