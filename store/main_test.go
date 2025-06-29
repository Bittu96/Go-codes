package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddProduct(t *testing.T) {
	store := Store{
		Products: make(map[int]*Product),
	}
	err := store.AddProduct(
		Product{
			Id:    1,
			Name:  "Laptop",
			Price: 1000,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(store.Products))
	err = store.AddProduct(
		Product{
			Id:    1,
			Name:  "Laptop",
			Price: 1000,
		},
	)
	assert.ErrorContains(t, err, "product already exists")
}
func TestGetProduct(t *testing.T) {
	store := Store{
		Products: make(map[int]*Product),
	}
	p := Product{
		Id:    1,
		Name:  "Laptop",
		Price: 1000,
	}
	store.AddProduct(p)
	product := store.GetProduct(1)
	assert.Equal(t, &p, product)
}
func TestStoreClose(t *testing.T) {
	store := Store{
		Products: make(map[int]*Product),
	}
	store.Close()
	assert.False(t, store.isOpen)
}

// func TestStoreOpen(t *testing.T) {
// store := Store{
// Products: make(map[int]*Product),
// }
// store.Open()
// assert.True(t, store.isOpen)
// } main_test.go file
