package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// Item struct to hold item details.  Using স্বতন্ত্র tags for JSON marshaling.
type Item struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// MemoryStore struct to hold the items in memory.
// Using a mutex to protect concurrent access to the map.
type MemoryStore struct {
	items map[int]Item
	mu    sync.RWMutex // Use RWMutex for better concurrency
	nextID int         // Keep track of the next available ID
}

// NewMemoryStore creates a new MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		items:  make(map[int]Item),
		nextID: 1, // Start IDs from 1
	}
}

// GetItem retrieves an item by ID.
func (s *MemoryStore) GetItem(id int) (Item, error) {
	s.mu.RLock() // Use RLock for read-only access
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	if !ok {
		return Item{}, fmt.Errorf("item with id %d not found", id)
	}
	return item, nil
}

// AddItem adds a new item to the store.
func (s *MemoryStore) AddItem(item Item) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

    // Assign a unique ID.
    newItem := Item{
        ID: s.nextID,
        Name: item.Name,
        Value: item.Value,
    }
	s.items[s.nextID] = newItem
	s.nextID++
	return newItem.ID, nil
}

// UpdateItem updates an existing item.
func (s *MemoryStore) UpdateItem(id int, item Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return fmt.Errorf("item with id %d not found", id)
	}
	item.ID = id // Ensure the ID is not changed.
	s.items[id] = item
	return nil
}

// DeleteItem deletes an item by ID.
func (s *MemoryStore) DeleteItem(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return fmt.Errorf("item with id %d not found", id)
	}
	delete(s.items, id)
	return nil
}

// GetItems returns all items.
func (s *MemoryStore) GetItems() ([]Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Item, 0, len(s.items)) // Pre-allocate for efficiency.
	for _, item := range s.items {
		items = append(items, item)
	}
	return items, nil
}

// App struct to hold the store and any other dependencies.
type App struct {
	store *MemoryStore
}

// NewApp creates a new App.
func NewApp(store *MemoryStore) *App {
	return &App{
		store: store,
	}
}

// getItemHandler handles the retrieval of a single item.
func (a *App) getItemHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the URL.  Assume a simple path like /items/{id}.
	idStr := r.URL.Path[len("/items/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	item, err := a.store.GetItem(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(item); err != nil {
		log.Printf("Error encoding JSON: %v", err) // Log the error.
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// addItemHandler handles the creation of a new item.
func (a *App) addItemHandler(w http.ResponseWriter, r *http.Request) {
	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation.
	if item.Name == "" {
		http.Error(w, "Item name cannot be empty", http.StatusBadRequest)
		return
	}
    if item.Value < 0 {
        http.Error(w, "Item value cannot be negative", http.StatusBadRequest)
        return
    }

	id, err := a.store.AddItem(item)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Use the newly assigned ID in the response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // Use 201 Created for successful creation.
	response := map[string]int{"id": id}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// updateItemHandler handles the updating of an existing item.
func (a *App) updateItemHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/items/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = a.store.UpdateItem(id, item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content for successful update with no body.
}

// deleteItemHandler handles the deletion of an item.
func (a *App) deleteItemHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/items/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	err = a.store.DeleteItem(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content for successful deletion.
}

// getItemsHandler handles the retrieval of all items.
func (a *App) getItemsHandler(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.GetItems()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func main() {
	store := NewMemoryStore()
	app := NewApp(store)

	// Use http.NewServeMux() for more explicit routing.
	mux := http.NewServeMux()

	// Register handlers with the mux.  This makes the routing very clear.
	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path == "/items/" { //list all items
				app.getItemsHandler(w, r)
			} else { //get item by id
				app.getItemHandler(w, r)
			}

		case http.MethodPost:
			app.addItemHandler(w, r)
		case http.MethodPut:
			app.updateItemHandler(w, r)
		case http.MethodDelete:
			app.deleteItemHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
    mux.HandleFunc("/items", app.getItemsHandler) //handles /items

	// Start the server using the mux.
	fmt.Println("Server listening on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
