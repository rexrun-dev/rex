package generate

import "rexrun.dev/rex/internal/detect"

// Compose generates a docker-compose.yml for the detected stack.
func Compose(d *detect.Detection) string {
	service := d.Stack
	port := "8080"

	switch d.Stack {
	case "go":
		port = "8080"
	case "node":
		port = "3000"
	case "python":
		port = "8000"
	case "ruby":
		port = "3000"
	case "php":
		port = "80"
	case "elixir":
		port = "4000"
	}

	return `version: "3.8"

services:
  ` + service + `:
    build: .
    ports:
      - "` + port + `:` + port + `"
    environment:
      - PORT=` + port + `
    volumes:
      - .:/app
`
}
