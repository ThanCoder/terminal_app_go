package serverapp

import (
	"fmt"
	"log"
	"net/http"
)

func RunApp() {
	port := ":8080"
	fmt.Printf("Starting Server on http://localhost%s\n", port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello from Go Server!")
	})
	log.Fatal(http.ListenAndServe(port, nil))
}
