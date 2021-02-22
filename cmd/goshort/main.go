package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

const (
	configPath  = "/etc/routes.json"
	secretsPath = "secrets.json"
	hashLen     = 6
	postRoute   = "/post/new-route"
)

type Routing struct {
	Config Config
	Routes Routes
	mu *sync.Mutex
}

type Config struct {
	Urls []string `json:"urls"`
}

type Secrets struct {
	AuthTok    string
	SecretAuth string
}

// "sha-1 short hash": "https://someurl.com/some/path"
type Routes map[string]string

type AddRouteRequest struct {
	AuthTok    string   `json:"authTok"`
	SecretAuth string   `json:"secretAuth"`
	Routes     []string `json:"routes"`
}

func main() {
	log.Println("starting url shortener")
	ro, err := readConfig(configPath)

	if err != nil {
		log.Fatalf("config error - %+v", err)
	}

	http.HandleFunc(postRoute, ro.AddNewRoute)
	http.HandleFunc("/", ro.GenericServer)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (r *Routes) makeRoutes(cfg Config) error {
	for _, val := range cfg.Urls {
		raw := sha1.New()
		raw.Write([]byte(val))
		h := raw.Sum(nil)
		hash := fmt.Sprintf("%x", h)

		ro := *r
		ro[hash[:hashLen]] = val
		r = &ro
		log.Printf("made: %s --> %s", hash[:hashLen], val)
	}

	return nil
}

func (ro *Routing) GenericServer(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/"):]
	if val, ok := ro.Routes[title]; ok {
		log.Printf("request: %s --> maps to: %s", title, val)
		http.Redirect(w, r, val, http.StatusFound)
	} else {
		log.Printf("hash was %s which is not in the map. \n%+v", title, ro.Routes)
		http.Redirect(w, r, "https://brennanmacaig.com/notfound", http.StatusFound)
	}
}

func (ro *Routing) AddNewRoute(w http.ResponseWriter, r *http.Request) {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	decoder := json.NewDecoder(r.Body)
	var req AddRouteRequest
	err := decoder.Decode(&req)
	if err != nil {
		log.Printf("non-fatal error: could not read request body for new route into struct - %+v", err)
		return
	}

	// check keys
	// if keys are good, add to config
	// and then slap config back into file

	sec, err := readSecrets(secretsPath)
	if err != nil {
		log.Printf("non-fatal error: could not read secrets file - %+v", err)
		return
	}
	if sec.AuthTok != req.AuthTok && sec.SecretAuth != req.SecretAuth {
		log.Printf("login attempt denied - user/pass does not match.")
		return
	}

	for _, val := range req.Routes {
		log.Printf("adding route: %s", val)
		ro.Config.Urls = append(ro.Config.Urls, val)
	}

	byt, err := json.Marshal(ro.Config)
	if err != nil {
		log.Fatalf("fatal error: could not marshal json - %+v", err)
	}

	err = ioutil.WriteFile(configPath, byt, 0664)
	if err != nil {
		log.Fatalf("fatal error: could not write config file - %+v", err)
	}

	err = ro.Routes.makeRoutes(ro.Config)
	log.Printf("ro.Routes: %+v", ro.Routes)
	if err != nil {
		log.Fatalf("fatal error: could not construct routes from config - %+v", err)
	}

	return
}

func readConfig(path string) (Routing, error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return Routing{}, fmt.Errorf("could not read config file - %w", err)
	}
	var cfg Config
	err = json.Unmarshal(dat, &cfg)
	if err != nil {
		return Routing{}, fmt.Errorf("could not unmarshal - %w", err)
	}

	if len(cfg.Urls) < 1 {
		return Routing{}, fmt.Errorf("must be some routes in config file")
	}

	if err != nil {
		return Routing{}, fmt.Errorf("could not construct routes")
	}
	r := Routing{
		Config: cfg,
		Routes: make(map[string]string),
		mu: &sync.Mutex{},
	}
	err = r.Routes.makeRoutes(r.Config)
	return r, nil
}

func readSecrets(path string) (Secrets, error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return Secrets{}, fmt.Errorf("could not read secrets file - %w", err)
	}
	var sec Secrets
	err = json.Unmarshal(dat, &sec)
	if err != nil {
		return Secrets{}, fmt.Errorf("could not unmarshal - %w", err)
	}
	return sec, nil
}
