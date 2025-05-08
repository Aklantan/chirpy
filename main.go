package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/aklantan/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type errorResponse struct {
	Error string `json:"error"`
}

/*
type validResponse struct {
	Valid       bool   `json:"valid`
	CleanedBody string `json:"cleaned_body"`
	Body        string `json:"body"`
}
*/

type chirpResponse struct {
	ID         uuid.UUID `json:"id"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
	Body       string    `json:"body"`
	User_ID    uuid.UUID `json:"user_id"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type ChirpRequest struct {
	Body    string `json:"body"`
	User_ID string `json:"user_id"`
}

/*
API state and methods
*/
type apiConfig struct {
	fileserverHits atomic.Int32
	db_query       *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) writeHits(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	str := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
	_, err := w.Write([]byte(str))
	if err != nil {
		fmt.Println("Cannot write Hits to response")
	}
}

func (cfg *apiConfig) resetHits(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	cfg.db_query.DeleteUser(r.Context())
	cfg.fileserverHits.Swap(0)
}

func (cfg *apiConfig) validateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
	}

	var dat []byte

	if len(params.Body) <= 140 {
		respBody := chirpResponse{
			Body: removeProfanity(params.Body),
		}

		respondWithJSON(w, 200, respBody)
		return
	} else {
		respBody := errorResponse{
			Error: "Chirp is too long",
		}

		dat, err = json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(dat)
		return
	}

}

func (cfg *apiConfig) addUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
	dbUser, err := cfg.db_query.CreateUser(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		w.WriteHeader(500)
		return
	}
	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
	respondWithJSON(w, 201, user)
}

/*
Response handling
*/

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")

	errResp := errorResponse{
		Error: msg,
	}
	dat, err := json.Marshal(errResp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")

	dat, err := json.Marshal(payload)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling JSON: %s", err)
		respondWithError(w, 500, msg)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

/*
profanity Filters
*/

var profanity = []string{"kerfuffle", "sharbert", "fornax"}

func removeProfanity(text string) string {
	splitText := strings.Split(text, " ")
	for i, word := range splitText {
		for _, prof := range profanity {
			if strings.ToLower(word) == prof {
				splitText[i] = "****"
			}
		}
	}
	returnString := strings.Join(splitText, " ")
	return returnString

}

func main() {
	godotenv.Load()

	const port = "8081"

	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		os.Exit(1)
	}

	dbQueries := database.New(db)

	apiCfg := &apiConfig{
		db_query: dbQueries,
	}

	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:    "127.0.0.1:" + port,
		Handler: mux,
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("./"))))
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHits)
	mux.HandleFunc("GET /admin/metrics", apiCfg.writeHits)
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.validateChirp)
	mux.HandleFunc("POST /api/users", apiCfg.addUser)

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))

	})

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())

}
