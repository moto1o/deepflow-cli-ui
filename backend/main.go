package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/moto1o/deepflow-cli-ui/backend/handler"
)

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Df-Ip, X-Df-Api-Port, X-Df-Org-Id, X-Df-Timeout")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func main() {
	mux := http.NewServeMux()

	// ---- agent-groups ----
	mux.HandleFunc("/api/agent-groups", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.ListAgentGroups(w, r)
		case http.MethodPost:
			handler.CreateAgentGroup(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/agent-groups/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/agent-groups/")
		if r.Method == http.MethodDelete {
			handler.DeleteAgentGroup(w, r, name)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	// ---- agents ----
	mux.HandleFunc("/api/agents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.ListAgents(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/agents/rebalance", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.RebalanceAgent(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/agents/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/agents/")
		switch r.Method {
		case http.MethodDelete:
			handler.DeleteAgent(w, r, name)
		case http.MethodPatch:
			handler.UpdateAgent(w, r, name)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// ---- agent-group-configs ----
	mux.HandleFunc("/api/agent-group-configs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.ListConfigs(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/agent-group-configs/template", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetConfigTemplate(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/agent-group-configs/", func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, "/api/agent-group-configs/")
		// 形如 {shortUuid} 或 {shortUuid}/yaml
		shortUUID := rest
		isYAML := false
		if strings.HasSuffix(rest, "/yaml") {
			shortUUID = strings.TrimSuffix(rest, "/yaml")
			isYAML = true
		}
		switch {
		case r.Method == http.MethodGet && isYAML:
			handler.GetConfigYAML(w, r, shortUUID)
		case r.Method == http.MethodPost && isYAML:
			handler.SaveConfigYAML(w, r, shortUUID, http.MethodPost)
		case r.Method == http.MethodPut && isYAML:
			handler.SaveConfigYAML(w, r, shortUUID, http.MethodPut)
		case r.Method == http.MethodDelete:
			handler.DeleteConfig(w, r, shortUUID)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// 包一层 CORS
	handlerWithCORS := withCORS(mux.ServeHTTP)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("deepflow-cli-ui backend listening on %s", addr)
	if err := http.ListenAndServe(addr, handlerWithCORS); err != nil {
		log.Fatal(err)
	}
}
