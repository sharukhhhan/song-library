package main

import (
	_ "effective_mobile_tz/docs"
	"effective_mobile_tz/internal/app"
)

const configPath = "config/config.yaml"

func main() {
	app.Run(configPath)
}
