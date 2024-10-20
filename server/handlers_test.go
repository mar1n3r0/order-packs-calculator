package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestPostPack(t *testing.T) {
   router := InitRouter()
   pack := Pack{Size: 10}
   jsonData, _ := json.Marshal(pack)

   req, _ := http.NewRequest("POST", "/packs", bytes.NewBuffer(jsonData))
   req.Header.Set("Content-Type", "application/json")

   w := httptest.NewRecorder()
   router.ServeHTTP(w, req)

   if w.Code != http.StatusOK {
       t.Errorf("Expected status code 200, got %d", w.Code)
   }
}

func TestGetAllPacks(t *testing.T) {
   router := InitRouter()

   req, _ := http.NewRequest("GET", "/packs", nil)
   w := httptest.NewRecorder()
   router.ServeHTTP(w, req)

   if w.Code != http.StatusOK {
       t.Errorf("Expected status code 200, got %d", w.Code)
   }
}

func TestGetPack(t *testing.T) {
   router := InitRouter()
   
   // First create a pack to retrieve it later
   pack := Pack{Size: 10}
   jsonData, _ := json.Marshal(pack)

   reqPost, _ := http.NewRequest("POST", "/packs", bytes.NewBuffer(jsonData))
   reqPost.Header.Set("Content-Type", "application/json")
   
   wPost := httptest.NewRecorder()
   router.ServeHTTP(wPost, reqPost)

   var createdPack Pack
   json.Unmarshal(wPost.Body.Bytes(), &createdPack)

   reqGet, _ := http.NewRequest("GET", "/packs/"+createdPack.ID, nil)
   wGet := httptest.NewRecorder()
   router.ServeHTTP(wGet, reqGet)

   if wGet.Code != http.StatusOK {
       t.Errorf("Expected status code 200 for getting pack, got %d", wGet.Code)
   }
}

func TestUpdatePack(t *testing.T) {
   router := InitRouter()

   // Create a pack first
   pack := Pack{Size: 10}
   jsonData, _ := json.Marshal(pack)

   reqPost, _ := http.NewRequest("POST", "/packs", bytes.NewBuffer(jsonData))
   reqPost.Header.Set("Content-Type", "application/json")
   
   wPost := httptest.NewRecorder()
   router.ServeHTTP(wPost, reqPost)

   var createdPack Pack
   json.Unmarshal(wPost.Body.Bytes(), &createdPack)

   // Update the pack
   createdPack.Size = 20
   updatedData, _ := json.Marshal(createdPack)

   reqPut, _ := http.NewRequest("PUT", "/packs/"+createdPack.ID, bytes.NewBuffer(updatedData))
   reqPut.Header.Set("Content-Type", "application/json")

   wPut := httptest.NewRecorder()
   router.ServeHTTP(wPut, reqPut)

   if wPut.Code != http.StatusOK {
       t.Errorf("Expected status code 200 for updating pack, got %d", wPut.Code)
   }
}

func TestDeletePack(t *testing.T) {
   router := InitRouter()

   // Create a pack first
   pack := Pack{Size: 10}
   jsonData, _ := json.Marshal(pack)

   reqPost, _ := http.NewRequest("POST", "/packs", bytes.NewBuffer(jsonData))
   reqPost.Header.Set("Content-Type", "application/json")
   
   wPost := httptest.NewRecorder()
   router.ServeHTTP(wPost, reqPost)

   var createdPack Pack
   json.Unmarshal(wPost.Body.Bytes(), &createdPack)
}