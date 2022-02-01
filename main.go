package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func main() {
	configFilename := flag.String("config", "config.yaml", "config filename")

	flag.Parse()
	//signal catch
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt) //notify when Ctrl+C
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	configPaths := []string{"config", "../../config", "/go/bin/config"}
	viper.SetConfigName(*configFilename)
	viper.SetConfigType("yaml")
	for _, configPath := range append(configPaths, ".") {
		viper.AddConfigPath(configPath)
	}
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	log.SetOutput(file)

	log.Println("Hello world!")

	ctx, cancel := context.WithCancel(context.Background())

	server, err := initServer(ctx)
	if err != nil {
		log.Println("can not initialise server : %#v --> exit", err)
		os.Exit(1)
	}
	//Run monitoring server
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("listen:%+s\n", err)
		}
	}()

	log.Println("waiting interrupt signal")
	<-signals

	cancel()
}
func setRouter(service *Frx) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/frx/record/{customerID:[0-9]+}/{transactionType}/", HandleRecording(service)).Methods("POST")

	router.HandleFunc("/frx/authorization/{customerID:[0-9]+}/{transactionType}/", HandleAuthorization(service)).Methods("POST")

	router.HandleFunc("/frx/risk/{customerID:[0-9]+}/{transactionType}/", HandleRiskManagement(service)).Methods("POST")

	router.HandleFunc("/admin/health/", HandleServiceAvailability(service)).Methods("POST")

	router.HandleFunc("/frx/initialize/{customerID:[0-9]+}/", HandleInitialization(service)).Methods("POST")
	return router
}
func HandleInitialization(service *Frx) func(http.ResponseWriter, *http.Request) {
	return handleHTTPError(func(resp http.ResponseWriter, req *http.Request) error {
		params := mux.Vars(req)
		customerID := params["customerID"]
		resp.Header()["RequestUuid"] = []string{req.Header.Get("RequestUuid")}
		c := http.Client{}
		fmt.Println("req is", req)
		url := "http://" + service.Config.Host + ":" + service.Config.Port + "/frx/initialize/" + customerID + "/"
		request, err := http.NewRequest("POST", url, req.Body)
		if err != nil {
			log.Println("can not Initiliaze request : %#v --> exit", err)
		}

		respVir, err := c.Do(request)
		if err != nil {
			log.Println("can not Do request : %#v --> exit", err)
		}

		defer respVir.Body.Close()
		// headers
		for name, values := range respVir.Header {
			resp.Header()[name] = values
			fmt.Println(name, "  ", values)
		}
		// status (must come after setting headers and before copying body)
		resp.WriteHeader(respVir.StatusCode)
		// body
		io.Copy(resp, respVir.Body)

		return nil
	})
}
func HandleRecording(service *Frx) func(http.ResponseWriter, *http.Request) {
	return handleHTTPError(func(resp http.ResponseWriter, req *http.Request) error {
		params := mux.Vars(req)
		customerID := params["customerID"]
		transactionType := params["transactionType"]

		resp.Header()["RequestUuid"] = []string{req.Header.Get("RequestUuid")}
		c := http.Client{}
		url := "http://" + service.Config.Host + ":" + service.Config.Port + "/frx/record/" + customerID + "/" + transactionType + "/"

		request, err := http.NewRequest("POST", url, req.Body)
		if err != nil {
			log.Println("can not Initiliaze request : %#v --> exit", err)
		}
		request.Header.Add("Accept", `application/json`)

		respVir, err := c.Do(request)
		if err != nil {
			log.Println("can not Do request : %#v --> exit", err)
		}

		defer respVir.Body.Close()
		// headers
		for name, values := range respVir.Header {
			resp.Header()[name] = values
			fmt.Println(name, "  ", values)
		}
		// status (must come after setting headers and before copying body)
		resp.WriteHeader(respVir.StatusCode)
		// body
		io.Copy(resp, respVir.Body)

		return nil
	})
}
func HandleRiskManagement(service *Frx) func(http.ResponseWriter, *http.Request) {
	return handleHTTPError(func(resp http.ResponseWriter, req *http.Request) error {
		resp.Header()["RequestUuid"] = []string{req.Header.Get("RequestUuid")}
		params := mux.Vars(req)
		customerID := params["customerID"]
		transactionType := params["transactionType"]
		c := http.Client{}
		url := "http://" + service.Config.Host + ":" + service.Config.Port + "/frx/risk/" + customerID + "/" + transactionType + "/"

		request, err := http.NewRequest("POST", url, req.Body)
		fmt.Println(url)
		if err != nil {
			log.Println("can not Initiliaze request : %#v --> exit", err)
		}
		request.Header.Add("Accept", `application/json`)
		respVir, err := c.Do(request)
		if err != nil {
			log.Println("can not Do request : %#v --> exit", err)
		}

		defer respVir.Body.Close()
		// headers
		for name, values := range respVir.Header {
			resp.Header()[name] = values
			fmt.Println(name, "  ", values)
		}
		// status (must come after setting headers and before copying body)
		resp.WriteHeader(respVir.StatusCode)
		// body
		io.Copy(resp, respVir.Body)

		return nil
	})
}
func HandleServiceAvailability(service *Frx) func(http.ResponseWriter, *http.Request) {
	return handleHTTPError(func(resp http.ResponseWriter, req *http.Request) error {
		resp.Header()["RequestUuid"] = []string{req.Header.Get("RequestUuid")}
		c := http.Client{}
		url := "http://" + service.Config.Host + ":" + service.Config.Port + "/admin/health/"
		request, err := http.NewRequest("POST", url, req.Body)
		if err != nil {
			log.Println("can not Initiliaze request : %#v --> exit", err)
		}
		request.Header.Add("Accept", `application/json`)
		respVir, err := c.Do(request)
		if err != nil {
			log.Println("can not Do request : %#v --> exit", err)
		}

		defer respVir.Body.Close()
		// headers
		for name, values := range respVir.Header {
			resp.Header()[name] = values
			fmt.Println(name, "  ", values)
		}
		// status (must come after setting headers and before copying body)
		resp.WriteHeader(respVir.StatusCode)
		// body
		io.Copy(resp, respVir.Body)

		return nil
	})
}
func HandleAuthorization(service *Frx) func(http.ResponseWriter, *http.Request) {
	return handleHTTPError(func(resp http.ResponseWriter, req *http.Request) error {

		params := mux.Vars(req)
		customerID := params["customerID"]
		transactionType := params["transactionType"]
		resp.Header()["RequestUuid"] = []string{req.Header.Get("RequestUuid")}
		c := http.Client{}
		url := "http://" + service.Config.Host + ":" + service.Config.Port + "/frx/authorization/" + customerID + "/" + transactionType + "/"

		request, err := http.NewRequest("POST", url, req.Body)
		if err != nil {
			log.Println("can not Initiliaze request : %#v --> exit", err)
		}
		request.Header.Add("Accept", `application/json`)
		respVir, err := c.Do(request)
		if err != nil {
			log.Println("can not Do request : %#v --> exit", err)
		}

		defer respVir.Body.Close()
		// headers
		for name, values := range respVir.Header {
			resp.Header()[name] = values
			fmt.Println(name, "  ", values)
		}
		// status (must come after setting headers and before copying body)
		resp.WriteHeader(respVir.StatusCode)
		// body
		io.Copy(resp, respVir.Body)

		return nil
	})
}

func initServer(ctx context.Context) (*http.Server, error) {
	var config Config
	viper.UnmarshalKey("service", &config)
	frxService := &Frx{
		Ctx:    ctx,
		Config: config,
	}
	return &http.Server{
		Addr:    ":" + "9100",
		Handler: setRouter(frxService),
	}, nil
}

type HTTPError struct {
	Code    int
	Message string
}

func (err *HTTPError) Error() string {
	return fmt.Sprintf("Http error %d : %s", err.Code, err.Message)
}
func writeError(resp http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	httpErr, ok := err.(*HTTPError)
	if !ok {
		httpErr = simpleHTTPError(http.StatusInternalServerError, err.Error()) // factory, construit une HttpError avec une ErrorDetails en fct des paramètres
	}
	resp.Header().Set("Content-type", "application/json")
	resp.WriteHeader(httpErr.Code)
	json.NewEncoder(resp).Encode(httpErr.Message)
}

func simpleHTTPError(code int, message string) *HTTPError {
	return &HTTPError{code, message}
}

type handlerFuncError func(http.ResponseWriter, *http.Request) error

func handleHTTPError(h handlerFuncError) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) { // retourne http.Handlerfunc, peut donc être utilisé comme handler dans un mux
		err := h(resp, req)   // h retourne nil si tout va bien, une erreur sinon
		writeError(resp, err) // si err == nil, noop, sinon transforme err en status code & json
	}
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
type Frx struct {
	Config Config
	Ctx    context.Context
}
type Config struct {
	Port string
	Host string
}
