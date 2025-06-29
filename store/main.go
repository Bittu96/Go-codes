package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"
)

type Product struct {
	Id          int
	Name        string
	Description string
	Price       float64
	Quantity    int
	Discount    float64
}

func (p Product) CheckStock() bool {
	return p.Quantity > 0
}

type Store struct {
	Products map[int]*Product
	isOpen   bool
}

func (s Store) AddProduct(p Product) error {
	if _, exists := s.Products[p.Id]; exists {
		return errors.New("product already exists")
	}
	s.Products[p.Id] = &p
	return nil
}
func (s Store) GetProduct(id int) *Product {
	return s.Products[id]
}
func (s *Store) Close() {
	s.isOpen = false
}
func (s *Store) Open() {
	s.isOpen = true
}
func addProductHandler(w http.ResponseWriter, r *http.Request) {
	var p Product
	json.NewDecoder(r.Body).Decode(&p)
	store := getStoreInstance()
	if err := store.AddProduct(p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
func getProductHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	productId, _ := strconv.Atoi(id)
	store := getStoreInstance()
	product := store.GetProduct(productId)
	json.NewEncoder(w).Encode(product)
}
func openStoreHandler(w http.ResponseWriter, r *http.Request) {
	store := getStoreInstance()
	if store.isOpen {
		json.NewEncoder(w).Encode("Store is already opened")
	}
	store.Open()
	json.NewEncoder(w).Encode("Store opened")
}
func closeStoreHandler(w http.ResponseWriter, r *http.Request) {
	store := getStoreInstance()
	store.Close()
	json.NewEncoder(w).Encode("Store closed")
}
func MinnMax(min float64, max float64) []Product {
	store := getStoreInstance()
	var list_of_prod []Product
	for _, prod_obj := range store.Products {
		if max > prod_obj.Price*(1-prod_obj.Discount/100) && min < prod_obj.Price*(1-prod_obj.Discount/100) {
			list_of_prod = append(list_of_prod, *prod_obj)
		}
	}
	return list_of_prod
}

var storeInstance *Store
var once sync.Once

func getStoreInstance() *Store {
	once.Do(func() {
		storeInstance = &Store{
			Products: make(map[int]*Product),
		}
	})
	return storeInstance
}

func main() {
	http.HandleFunc("/add", addProductHandler)
	http.HandleFunc("/get", getProductHandler)
	http.HandleFunc("/open", openStoreHandler)
	http.HandleFunc("/close", closeStoreHandler)
	http.ListenAndServe(":8080", nil)
}
