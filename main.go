package main


import ( 
	"fmt"
	"net/http"
	"log"
)


func main() {
	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.HandleFunc("/healthz", handlerReadyCheck)

	srv := http.Server {
		Addr:		":" + port,
		Handler:	mux,
	}

	if err := srv.ListenAndServe(); err != nil {
		// Error starting or closing listener
		fmt.Errorf("HTTP server ListenAndServe: %v", err)
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
}

func handlerReadyCheck(w http.ResponseWriter, req *http.Request){
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

