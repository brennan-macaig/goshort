package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	configPath = "C:\\Users\\Brennan\\go\\src\\goshort\\routes.json"
	hashLen    = 6
)

type Config struct {
	Urls []string `json:"urls"`
}

// "sha-1 short hash": "https://someurl.com/some/path"
type Routes map[string]string

func main() {
	log.Println("starting url shortener")

	routes, err := makeRoutes()
	if err != nil {
		log.Fatalf("fatal error - %+v", err)
	}

	buildRoutes(routes)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func makeRoutes() (Routes, error) {
	cfg, err := readConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("config error - %w", err)
	}

	r := make(Routes)
	for _, val := range cfg.Urls {
		raw := sha1.New()
		raw.Write([]byte(val))
		h := raw.Sum(nil)
		hash := fmt.Sprintf("%x", h)

		r[hash[:hashLen]] = val

		log.Printf("made: %s --> %s", hash[:hashLen], val)
	}

	return r, nil
}

func (ro Routes) GenericServer(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/"):]
	if val, ok := ro[title]; ok {
		log.Printf("request: %s --> maps to: %s", title, val)
		http.Redirect(w, r, val, http.StatusFound)
	} else {
		http.Redirect(w, r, "https://brennanmacaig.com/notfound", http.StatusFound)
	}
}

func buildRoutes(r Routes) {
	for idx, _ := range r {
		route := fmt.Sprintf("/%s", idx)
		http.HandleFunc(route, r.GenericServer)
	}
}

func readConfig(path string) (Config, error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("could not read config file - %w", err)
	}
	var cfg Config
	err = json.Unmarshal(dat, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("could not unmarshal - %w", err)
	}

	if len(cfg.Urls) < 1 {
		return Config{}, fmt.Errorf("must be some routes in config file")
	}

	return cfg, nil
}
