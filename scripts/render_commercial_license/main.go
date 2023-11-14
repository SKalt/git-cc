package main

import (
	"flag"
	"log"
	"os"
	"path"
	"regexp"
	"text/template"
	"time"
)

const licensePath = "./licenses/COMMERCIAL.template.md"

type Options struct {
	Date             string
	CustomerName     string
	CustomerLocation string
	CustomerEmail    string
	Users            int
	Price            int
}

var iso3166_2 *regexp.Regexp = regexp.MustCompile(`^[A-Z]{2}-[A-Z0-9]{2,3}$`)
var emailPattern *regexp.Regexp = regexp.MustCompile(`^[\w\-\.]+@([\w-]+\.)+[\w-]{2,}$`)

func parseArgs() (opts Options) {
	name := flag.String("name", "", "The name of the customer")
	loc := flag.String("location", "", "The ISO 3166-2 location of the customer")
	email := flag.String("email", "", "The email of the customer")
	price := flag.Int("price", 0, "The price of the license")
	nUsers := flag.Int("users", 0, "The number of users")
	flag.Parse()
	if *name == "" {
		log.Fatal("-name is required")
	}
	// ISO 3166-2 matches [A-Z]{2}-[A-Z0-9]{2,3}
	if !iso3166_2.Match([]byte(*loc)) {
		log.Fatal("-location must be in ISO 3166-2 format")
	}
	if *email == "" {
		log.Fatal("-email is required")
	} else if !emailPattern.Match([]byte(*email)) {
		log.Fatal("-email must be a valid email address")
	}
	if *price <= 0 {
		log.Fatal("-price is required")
	}
	if *nUsers <= 0 {
		log.Fatal("-users is required")
	}
	opts = Options{
		Date:             time.Now().Format(time.DateOnly),
		CustomerName:     *name,
		CustomerLocation: *loc,
		CustomerEmail:    *email,
		Price:            *price,
		Users:            *nUsers,
	}
	return opts
}

func main() {
	opts := parseArgs()
	// read the license template
	name := path.Base(licensePath)
	tpl, err := template.New(name).ParseFiles(licensePath)
	if err != nil {
		panic(err)
	}
	err = tpl.Execute(os.Stdout, opts)
	if err != nil {
		panic(err)
	}
	return
}
