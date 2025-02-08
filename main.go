package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"

	_ "github.com/Esbaevnurdos/hackaton/docs"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Taraz Tourist Places API
// @version 1.0
// @description This API manages tourist places in Taraz, allowing CRUD operations, ratings, comments, and photos.
// @host localhost:8080
// @BasePath /

// Place represents a tourist place
type Place struct {
	ID          string   `json:"id"`
	PlaceName   string   `json:"placeName"`
	Rating      float64  `json:"rating"`
	Description string   `json:"description"`
	PhotoURLs   []string `json:"photoURLs"`
	Comments    []string `json:"comments"`
	Longitude   float64  `json:"longitude"`
	Latitude    float64  `json:"latitude"`
}

var (
	places   []Place
	mutex    sync.Mutex
	jsonFile = "places.json"
	lastID   int
)

// loadPlaces loads places from JSON file
func loadPlaces() {
	file, err := os.Open(jsonFile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	data, _ := ioutil.ReadAll(file)
	json.Unmarshal(data, &places)

	// Set lastID based on the highest ID found
	for _, p := range places {
		id, _ := strconv.Atoi(p.ID)
		if id > lastID {
			lastID = id
		}
	}
}

// savePlaces saves places to JSON file
func savePlaces() {
	data, _ := json.MarshalIndent(places, "", "  ")
	ioutil.WriteFile(jsonFile, data, 0644)
}

// @Summary Get all places
// @Description Retrieves all tourist places in Taraz
// @Tags places
// @Produce json
// @Success 200 {array} Place
// @Router /places [get]
func getPlaces(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(places)
}

// @Summary Get a place by ID
// @Description Retrieves details of a specific place
// @Tags places
// @Param id path string true "Place ID"
// @Produce json
// @Success 200 {object} Place
// @Failure 404
// @Router /places/{id} [get]
func getPlaceByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for _, p := range places {
		if p.ID == params["id"] {
			json.NewEncoder(w).Encode(p)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

// @Summary Add a new place
// @Description Adds a new tourist place
// @Tags places
// @Accept json
// @Produce json
// @Param place body Place true "New Place"
// @Success 201 {object} Place
// @Router /places [post]
func addPlace(w http.ResponseWriter, r *http.Request) {
	var newPlace Place
	json.NewDecoder(r.Body).Decode(&newPlace)
	mutex.Lock()
	lastID++
	newPlace.ID = strconv.Itoa(lastID)
	places = append(places, newPlace)
	savePlaces()
	mutex.Unlock()
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPlace)
}

// @Summary Update a place
// @Description Updates the details of a specific place
// @Tags places
// @Param id path string true "Place ID"
// @Param place body Place true "Updated Place"
// @Success 200
// @Failure 404
// @Router /places/{id} [put]
func updatePlace(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var updatedPlace Place
	json.NewDecoder(r.Body).Decode(&updatedPlace)
	mutex.Lock()
	for i, p := range places {
		if p.ID == params["id"] {
			updatedPlace.ID = p.ID // Ensure ID remains unchanged
			places[i] = updatedPlace
			savePlaces()
			mutex.Unlock()
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	mutex.Unlock()
	w.WriteHeader(http.StatusNotFound)
}

// @Summary Delete a place
// @Description Deletes a place by ID
// @Tags places
// @Param id path string true "Place ID"
// @Success 200
// @Failure 404
// @Router /places/{id} [delete]
func deletePlace(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	mutex.Lock()
	for i, p := range places {
		if p.ID == params["id"] {
			places = append(places[:i], places[i+1:]...)
			savePlaces()
			break
		}
	}
	mutex.Unlock()
	w.WriteHeader(http.StatusOK)
}

// @Summary Add a comment to a place
// @Description Adds a comment to a tourist place
// @Tags places
// @Param id path string true "Place ID"
// @Param comment body string true "Comment"
// @Success 200
// @Router /places/{id}/comment [post]
func addComment(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var comment struct {
		Text string `json:"text"`
	}
	json.NewDecoder(r.Body).Decode(&comment)
	mutex.Lock()
	for i, p := range places {
		if p.ID == params["id"] {
			places[i].Comments = append(places[i].Comments, comment.Text)
			savePlaces()
			break
		}
	}
	mutex.Unlock()
	w.WriteHeader(http.StatusOK)
}

// @Summary Add a rating to a place
// @Description Updates the rating of a place
// @Tags places
// @Param id path string true "Place ID"
// @Param rating body float64 true "New Rating"
// @Success 200
// @Router /places/{id}/rating [post]
func addRating(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var rating struct {
		Value float64 `json:"value"`
	}
	json.NewDecoder(r.Body).Decode(&rating)
	mutex.Lock()
	for i, p := range places {
		if p.ID == params["id"] {
			places[i].Rating = (places[i].Rating + rating.Value) / 2
			savePlaces()
			break
		}
	}
	mutex.Unlock()
	w.WriteHeader(http.StatusOK)
}

// @Summary Add a photo to a place
// @Description Adds a photo URL to a tourist place
// @Tags places
// @Accept json
// @Produce json
// @Param id path string true "Place ID"
// @Param photo body object{url=string} true "Photo URL"
// @Success 200
// @Failure 400
// @Router /places/{id}/photo [post]
func addPhoto(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var photo struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&photo); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()
	for i, p := range places {
		if p.ID == params["id"] {
			places[i].PhotoURLs = append(places[i].PhotoURLs, photo.URL)
			savePlaces()
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}


// @Summary Serve Swagger documentation
// @Description Provides API documentation
// @Tags docs
// @Success 200
// @Router /swagger/ [get]
func main() {
	loadPlaces()
	r := mux.NewRouter()

	r.HandleFunc("/places", getPlaces).Methods("GET")
	r.HandleFunc("/places/{id}", getPlaceByID).Methods("GET")
	r.HandleFunc("/places", addPlace).Methods("POST")
	r.HandleFunc("/places/{id}", updatePlace).Methods("PUT")
	r.HandleFunc("/places/{id}", deletePlace).Methods("DELETE")
	r.HandleFunc("/places/{id}/comment", addComment).Methods("POST")
	r.HandleFunc("/places/{id}/rating", addRating).Methods("POST")

	// Swagger route
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", r)
}
