package orchestration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/bububa/atomic-agents/tools/searxng"
)

func startSearxngServer(port int, results *searxng.Output) *http.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		json.NewEncoder(buf).Encode(results)
		io.Copy(w, buf)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/search", handler)
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("start searxng server failed: %v", err)
		}
	}()
	return srv
}
