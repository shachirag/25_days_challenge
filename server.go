package main

import (
	"challenge/app"
	"challenge/database"
	graph "challenge/graph/resolvers"
	"challenge/middleware"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8080"

func main() {

	err := app.SetupAndRunApp()
	if err != nil {
		panic(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	db := database.Connect()
	s3 := database.GetS3Uploader()
	ses := database.GetSesClient()
	resolver := &graph.Resolver{
		DB:        db,
		S3Client:  s3,
		SESClient: ses,
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	})

	http.Handle("/query", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		opName := r.Header.Get("X-GraphQL-Operation-Name")
		if opName == "" {
			http.Error(w, "Operation name header is required", http.StatusBadRequest)
			return
		}

		authRequiredOperations := map[string]bool{
			"CompleteTask":              true,
			"SelfCareForm":              true,
			"GetSelfCareFormData":       true,
			"LoginCustomer":             false,
			"ForgotPassword":            false,
			"ResetPassword":             false,
			"SignUpUser":                false,
			"SocialLoginCustomer":       false,
			"VerifyOtpForResetPassword": false,
			"VerifyOtp":                 false,
		}
		if authRequiredOperations[opName] {
			middleware.ValidateJWT(srv).ServeHTTP(w, r)
		} else {
			srv.ServeHTTP(w, r)
		}
	}))

	http.Handle("/schema", srv)

	http.Handle("/", playground.Handler("GraphQL Playground", "/query"))

	log.Printf("connect to http://localhost:%s/ for GraphQL Playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
