package main

import (
    "context"
    "testing"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func TestDatabase(t *testing.T) {
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        t.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    defer client.Disconnect(context.TODO())

    collection := client.Database("packsdb").Collection("packs")
    db := Database{client: client, collection: collection}

    // Clean up before tests
    collection.DeleteMany(context.TODO(), bson.M{})

    // Test CreatePack
    pack := Pack{Size: 10}
    createdPack, err := db.CreatePack(pack)
    if err != nil {
        t.Fatalf("Failed to create pack: %v", err)
    }
    
    if createdPack.ID == "" {
        t.Error("Expected a valid ID for the created pack")
    }

    // Test GetAllPacks
    packs, err := db.GetAllPacks()
    if err != nil {
        t.Fatalf("Failed to get all packs: %v", err)
    }
    
    if len(packs) != 1 {
        t.Errorf("Expected 1 pack, got %d", len(packs))
    }

    // Test GetPack
    retrievedPack, err := db.GetPack(createdPack.ID)
    if err != nil {
        t.Fatalf("Failed to get pack: %v", err)
    }

    if retrievedPack.ID != createdPack.ID {
        t.Errorf("Expected pack ID %s, got %s", createdPack.ID, retrievedPack.ID)
    }

    // Test UpdatePack
    createdPack.Size = 20
    updatedPack, err := db.UpdatePack(createdPack)
    if err != nil {
        t.Fatalf("Failed to update pack: %v", err)
    }

    if updatedPack.Size != 20 {
        t.Errorf("Expected updated size 20, got %d", updatedPack.Size)
    }

    // Test DeletePack
    err = db.DeletePack(createdPack.ID)
    if err != nil {
        t.Fatalf("Failed to delete pack: %v", err)
    }

    packsAfterDelete, _ := db.GetAllPacks()
    
    if len(packsAfterDelete) != 0 {
        t.Errorf("Expected 0 packs after deletion, got %d", len(packsAfterDelete))
    }
}