package main

import (
    "context"
    "testing"
    "time"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func RunMongo(ctx context.Context, t *testing.T) testcontainers.Container {
    // Define the container request
    req := testcontainers.ContainerRequest{
        Image:        "mongo:latest",
        ExposedPorts: []string{"27017/tcp"},
        WaitingFor:   wait.ForListeningPort("27017"),
    }

    // Start the MongoDB container
    mongoContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    
    if err != nil {
        t.Fatalf("Failed to start MongoDB container: %v", err)
    }

    return mongoContainer
}

func TestDatabase(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    mongoContainer := RunMongo(ctx, t)
    defer mongoContainer.Terminate(ctx)

    host, err := mongoContainer.Host(ctx)
    if err != nil {
        t.Fatalf("Failed to get container host: %v", err)
    }

    port, err := mongoContainer.MappedPort(ctx, "27017")
    if err != nil {
        t.Fatalf("Failed to get mapped port: %v", err)
    }
    
    // Create a MongoDB client
    mongoURI := "mongodb://" + host + ":" + port.Port()
    clientOptions := options.Client().ApplyURI(mongoURI)
    
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        t.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    
    collection := client.Database("packsdb").Collection("packs")
    db := Database{client: client, collection: collection}

    // Clean up before tests
    collection.DeleteMany(ctx, bson.M{})

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