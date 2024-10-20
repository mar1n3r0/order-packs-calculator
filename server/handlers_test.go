package main

import (
    "bytes"
    "net/http"
    "net/http/httptest"
    "testing"
    "encoding/json"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

// Mock database functions for testing
var mockDatabase = make(map[string]Pack)

func TestPostPack(t *testing.T) {
    router := gin.Default()
    router.POST("/packs", postPack)

    // Sample pack data
    pack := Pack{ID: "1", Size: 10}

    // Convert pack to JSON
    jsonValue, _ := json.Marshal(pack)

    // Create a request
    req, _ := http.NewRequest("POST", "/packs", bytes.NewBuffer(jsonValue))
    req.Header.Set("Content-Type", "application/json")

    // Record the response
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    // Assert the response
    assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetAllPacks(t *testing.T) {
    router := gin.Default()
    router.GET("/packs", getPacks)

    // Populate mock database
    mockDatabase["1"] = Pack{ID: "1", Size: 20}

    // Create a request
    req, _ := http.NewRequest("GET", "/packs", nil)

    // Record the response
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    // Assert the response
    assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetPack(t *testing.T) {
    router := gin.Default()
    router.GET("/packs/:id", getPack)

    // Populate mock database
    mockDatabase["1"] = Pack{ID: "1", Size: 10}

    // Create a request
    req, _ := http.NewRequest("GET", "/packs/1", nil)

    // Record the response
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    // Assert the response
    assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdatePack(t *testing.T) {
    router := gin.Default()
    router.PUT("/packs/:id", updatePack)

    // Populate mock database
    mockDatabase["1"] = Pack{ID: "1", Size: 100}

    // Sample updated pack data
    updatedPack := Pack{Size: 100}

    // Convert updated pack to JSON
    jsonValue, _ := json.Marshal(updatedPack)

    // Create a request
    req, _ := http.NewRequest("PUT", "/packs/1", bytes.NewBuffer(jsonValue))
    req.Header.Set("Content-Type", "application/json")

    // Record the response
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    // Assert the response
    assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeletePack(t *testing.T) {
   router := gin.Default()
   router.DELETE("/packs/:id", deletePack)

   // Populate mock database
   mockDatabase["1"] = Pack{ID: "1", Size: 100}

   // Create a request
   req, _ := http.NewRequest("DELETE", "/packs/1", nil)

   // Record the response
   w := httptest.NewRecorder()
   router.ServeHTTP(w, req)

   // Assert the response
   assert.Equal(t, http.StatusNoContent, w.Code)
}