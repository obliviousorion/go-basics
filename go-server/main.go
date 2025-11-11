package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"log" // Added for better error logging
)

// --- Data Structures and Global State ---

// User defines the structure for a user object.
// The `json:"name"` tag is crucial, telling the `encoding/json` package
// how to map the struct field to the JSON key when encoding/decoding.
type User struct {
	Name string `json:"name"`
}

// userCache acts as our in-memory "database" to store User objects.
// Keys are integers (acting as user IDs), and values are User structs.
var userCache = make(map[int]User)

// cacheMutex is a Read-Write Mutex (RWMutex) used to protect userCache
// from race conditions when multiple goroutines (requests) try to read or write concurrently.
var cacheMutex sync.RWMutex

// nextID tracks the next ID to assign to a new user.
var nextID = 1

// --- Main Function and Server Setup ---

func main() {
	// Initialize a new HTTP request multiplexer (router).
	// This is responsible for matching incoming requests to their appropriate handlers.
	mux := http.NewServeMux()

	// 1. Root Handler: A simple health check or welcome message.
	mux.HandleFunc("/", handleRoot)

	// 2. RESTful API Handlers: Using the new Go 1.22 routing features (HTTP method + path pattern).
	// POST /users: Create a new user.
	mux.HandleFunc("POST /users", handleCreateUser)
	// GET /users/{id}: Fetch a user by their ID (the {id} is a path variable).
	mux.HandleFunc("GET /users/{id}", handleGetUser)
	// DELETE /users/{id}: Delete a user by their ID.
	mux.HandleFunc("DELETE /users/{id}", handleDeleteUser)

	// Start the HTTP server. http.ListenAndServe blocks execution until the server stops.
	fmt.Println("Server is listening on port 8080...")
	// We use log.Fatal to ensure any error during server startup (e.g., port already in use) is logged.
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// --- Handlers Implementation ---

// handleRoot simply responds with a static "Hello, World" message.
func handleRoot(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Fprintf writes the formatted string to the response writer (w).
	fmt.Fprintf(w, "Hello, Go API World!")
}

// handleCreateUser handles POST requests to /users to add a new user.
func handleCreateUser(
	w http.ResponseWriter,
	r *http.Request,
) {
	var user User
	
	// 1. Decode the JSON request body into the User struct.
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		// If JSON decoding fails (e.g., malformed JSON), return 400 Bad Request.
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Input Validation
	if user.Name == "" {
		// If the required 'name' field is missing, return 400 Bad Request.
		http.Error(w, "Name field is required", http.StatusBadRequest)
		return
	}

	// 3. Acquire Write Lock
	// We use Lock() because we are modifying the shared resource (userCache and nextID).
	cacheMutex.Lock()
	
	// Assign the current nextID as the new user's ID.
	userID := nextID
	// Store the new user in the cache.
	userCache[userID] = user
	// Increment the ID counter for the next user.
	nextID++
	
	// Release the lock immediately after modifying the shared resources.
	cacheMutex.Unlock()
	
	// 4. Send Response
	// Set the status code to 201 Created to indicate successful resource creation.
	w.WriteHeader(http.StatusCreated) 
	
	// Write a response body indicating the success and the assigned ID.
	// NOTE: The previous code had a bug where fmt.Fprintf was called before WriteHeader,
	// which would incorrectly set the status to 200 OK. This is now corrected.
	fmt.Fprintf(w, "User successfully created with ID: %d", userID)
	
	// Optional: In a real-world scenario, you might return the full created resource object 
	// or the location header (w.Header().Set("Location", "/users/"+strconv.Itoa(userID))).
}

// handleGetUser handles GET requests to /users/{id} to retrieve a user by ID.
func handleGetUser(
	w http.ResponseWriter,
	r *http.Request,
) {
	// 1. Extract and Convert Path Variable
	// r.PathValue("id") retrieves the value from the {id} segment in the route pattern.
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// If the 'id' is not a valid integer, return 400 Bad Request.
		http.Error(w, "Invalid user ID format: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Acquire Read Lock
	// We use RLock() because we are only reading the shared resource (userCache).
	// This allows multiple readers to access the map simultaneously.
	cacheMutex.RLock()
	user, ok := userCache[id]
	cacheMutex.RUnlock() // Release the lock immediately after reading.

	// 3. Check for User Existence
	if !ok {
		// If the user ID is not found in the map, return 404 Not Found.
		http.Error(
			w,
			fmt.Sprintf("User with ID %d not found", id),
			http.StatusNotFound,
		)
		return
	}

	// 4. Encode and Send Response
	// Set the Content-Type header to inform the client that the response body is JSON.
	w.Header().Set("Content-Type", "application/json")

	// Marshal the User struct into a JSON byte slice.
	j, err := json.Marshal(user)
	if err != nil {
		// If JSON encoding fails (shouldn't happen with simple structs), return 500 Internal Server Error.
		http.Error(
			w,
			"Error encoding JSON response",
			http.StatusInternalServerError,
		)
		return
	}
	
	// Set the status code to 200 OK for a successful GET.
	w.WriteHeader(http.StatusOK) 
	// Write the JSON byte slice as the response body.
	w.Write(j)
}

// handleDeleteUser handles DELETE requests to /users/{id} to remove a user.
func handleDeleteUser(
	w http.ResponseWriter,
	r *http.Request,
) {
	// 1. Extract and Convert Path Variable
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		// If the 'id' is not a valid integer, return 400 Bad Request.
		http.Error(w, "Invalid user ID format: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Acquire Write Lock
	// We use Lock() because we are modifying the shared resource (userCache).
	cacheMutex.Lock()
	
	// delete() is safe to call even if the key doesn't exist; it simply does nothing.
	delete(userCache, id) 
	
	// Release the lock.
	cacheMutex.Unlock()

	// 3. Send Response
	// HTTP 204 No Content is the standard successful response for DELETE operations.
	// It indicates the action was successful but there is no body to return.
	w.WriteHeader(http.StatusNoContent)
	
	// NOTE: If you write any content to the response writer (w) after setting 204,
	// HTTP clients might ignore it, as 204 responses are expected to be empty.
}