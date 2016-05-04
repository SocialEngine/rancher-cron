package main

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

var (
	router          = mux.NewRouter()
	healthcheckPort = ":8080"
)

func startHealthcheck() {
	router.HandleFunc("/", healthcheck).Methods("GET", "HEAD").Name("Healthcheck")
	logrus.Info("Healthcheck handler is listening on ", healthcheckPort)
	logrus.Fatal(http.ListenAndServe(healthcheckPort, router))
}

func healthcheck(w http.ResponseWriter, req *http.Request) {
	_, err := m.MetadataClient.GetSelfStack()
	if err != nil {
		logrus.Error("Healthcheck failed: unable to reach metadata")
		http.Error(w, "Failed to reach metadata server", http.StatusInternalServerError)
	} else {
		err = c.TestConnect()
		if err != nil {
			logrus.Error("Healthcheck failed: unable to reach Cattle")
			http.Error(w, "Failed to connect to Cattle ", http.StatusInternalServerError)
		} else {
			w.Write([]byte("OK"))
		}
	}
}
