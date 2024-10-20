package main

import (
    "context"
    "net/http"
    "os"

    // Importing necessary packages
    "github.com/gin-contrib/cors" // Middleware for CORS support
    "github.com/gin-gonic/gin"    // Gin framework for HTTP routing
    "github.com/google/uuid"       // Package for generating unique IDs
    "github.com/joho/godotenv"     // Package for loading environment variables from .env file
    "go.mongodb.org/mongo-driver/bson" // BSON encoding/decoding for MongoDB
    "go.mongodb.org/mongo-driver/mongo" // MongoDB driver for Go
    "go.mongodb.org/mongo-driver/mongo/options" // Options for MongoDB client
)

// Pack represents the data model for a pack with ID and Size fields.
type Pack struct {
    ID   string `json:"id" bson:"id"`   // Unique identifier for the pack
    Size int    `json:"size" bson:"size"` // Size of the pack
}

// Database encapsulates the MongoDB client and collection.
type Database struct {
    client     *mongo.Client       // MongoDB client
    collection *mongo.Collection    // Collection to perform operations on
}

// InitDatabase initializes the database connection and returns a Database instance.
func InitDatabase() Database {
    // Load environment variables from .env file
    err := godotenv.Load()
    if err != nil {
        panic(err) // Panic if loading .env file fails
    }
    
    // Get the MongoDB connection URL from environment variables
    mongoURL := os.Getenv("MONGO_URL")
    
    // Set up MongoDB client options with the provided URL
    clientOptions := options.Client().ApplyURI(mongoURL)
    
    // Connect to MongoDB using the specified options
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        panic(err) // Panic if connection fails
    }

    // Initialize the collection for packs in the packsdb database
    collection := client.Database("packsdb").Collection("packs")
    
    return Database{client: client, collection: collection} // Return the initialized database instance
}

// CreatePack inserts a new pack into the database and returns it.
func (db Database) CreatePack(pack Pack) (Pack, error) {
    pack.ID = uuid.New().String() // Generate a new unique ID for the pack

    _, err := db.collection.InsertOne(context.TODO(), pack) // Insert the pack into the collection
    if err != nil {
        return Pack{}, err // Return an error if insertion fails
    }

    return pack, nil // Return the created pack on success
}

// GetAllPacks retrieves all packs from the database.
func (db Database) GetAllPacks() ([]Pack, error) {
    var packs []Pack

    cursor, err := db.collection.Find(context.TODO(), bson.M{}) // Find all packs in the collection
    if err != nil {
        return nil, err // Return an error if retrieval fails
    }

    if err = cursor.All(context.TODO(), &packs); err != nil { // Decode all packs into the packs slice
        return nil, err // Return an error if decoding fails
    }

    return packs, nil // Return the retrieved packs on success
}

// GetPack retrieves a specific pack by its ID.
func (db Database) GetPack(id string) (Pack, error) {
    var pack Pack
    
    // Find one pack by its ID and decode it into the pack variable
    err := db.collection.FindOne(context.TODO(), bson.M{"id": id}).Decode(&pack)
    
    if err != nil {
        return Pack{}, err // Return an error if retrieval fails or pack not found
    }

    return pack, nil // Return the found pack on success
}

// UpdatePack updates an existing pack in the database.
func (db Database) UpdatePack(pack Pack) (Pack, error) {
   _, err := db.collection.UpdateOne(context.TODO(), bson.M{"id": pack.ID}, bson.M{"$set": pack}) 
   // Update the pack in the collection based on its ID

   if err != nil {
       return Pack{}, err // Return an error if update fails
   }

   return pack, nil // Return the updated pack on success
}

// DeletePack removes a specific pack from the database by its ID.
func (db Database) DeletePack(id string) error {
   _, err := db.collection.DeleteOne(context.TODO(), bson.M{"id": id}) 
   // Delete one pack from the collection based on its ID

   return err // Return any errors that occurred during deletion
}

// Global variable to hold database instance initialized at application start.
var database = InitDatabase()

// InitRouter sets up HTTP routes and middleware for handling requests.
func InitRouter() *gin.Engine {
   router := gin.Default()           // Create a new Gin router instance
   router.Use(cors.Default())        // Use default CORS middleware

   router.POST("/packs", postPack)   // Route for creating a new pack
   router.GET("/packs", getPacks)     // Route for retrieving all packs
   router.GET("/packs/:id", getPack)  // Route for retrieving a specific pack by ID
   router.PUT("/packs/:id", updatePack)  // Route for updating a specific pack by ID
   router.DELETE("/packs/:id", deletePack)  // Route for deleting a specific pack by ID
   
   return router                     // Return configured router instance
}

// postPack handles POST requests to create a new pack.
func postPack(ctx *gin.Context) {
   var pack Pack
   
   if err := ctx.ShouldBindJSON(&pack); err != nil { 
       ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) 
       return  // Return bad request status if JSON binding fails
   }

   res, err := database.CreatePack(pack) 
   if err != nil {
       ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) 
       return  // Return internal server error status if creation fails
   }

   ctx.JSON(http.StatusOK, res)  // Return created pack with OK status on success
}

// getAllPacks handles GET requests to retrieve all packs.
func getAllPacks(ctx *gin.Context) {
   packs, err := database.GetAllPacks() 
   if err != nil {
       ctx.JSON(http.StatusInternalServerError, gin.H{"error": err}) 
       return  // Return internal server error status if retrieval fails
   }
 
   ctx.JSON(http.StatusOK, packs)  // Return all packs with OK status on success
}

// getPack handles GET requests to retrieve a specific pack by ID.
func getPack(ctx *gin.Context) {
   id := ctx.Param("id")  // Extract ID from URL parameters

   pack, err := database.GetPack(id)
   if err != nil {
       ctx.JSON(http.StatusNotFound, gin.H{"error": "Pack not found"}) 
       return  // Return not found status if retrieval fails or no such pack exists
   }

   ctx.JSON(http.StatusOK, pack)  // Return found pack with OK status on success
}

// updatePack handles PUT requests to update a specific pack by ID.
func updatePack(ctx *gin.Context) {
   id := ctx.Param("id")  // Extract ID from URL parameters

   var pack Pack
   
   if err := ctx.ShouldBindJSON(&pack); err != nil { 
       ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) 
       return  // Return bad request status if JSON binding fails
   }

   pack.ID = id  // Ensure that the ID is set correctly for updating

   updatedPack, err := database.UpdatePack(pack)
   if err != nil {
       ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pack"}) 
       return  // Return internal server error status if update fails
   }

   ctx.JSON(http.StatusOK, updatedPack)  // Return updated pack with OK status on success
}

// deletePack handles DELETE requests to remove a specific pack by ID.
func deletePack(ctx *gin.Context) {
   id := ctx.Param("id")  // Extract ID from URL parameters

   if err := database.DeletePack(id); err != nil { 
       ctx.JSON(http.StatusNotFound, gin.H{"error": "Failed to delete pack"}) 
       return  // Return not found status if deletion fails or no such pack exists
   }

   ctx.JSON(http.StatusNoContent, nil)  // Return No Content status (204) on successful deletion
}

// getPacks handles GET requests to retrieve all packs (duplicate function).
func getPacks(ctx *gin.Context) {
   packs, err := database.GetAllPacks()
   if err != nil {
       ctx.JSON(http.StatusInternalServerError, gin.H{"error": err}) 
       return  // Return internal server error status if retrieval fails
   }

   ctx.JSON(http.StatusOK, packs)  // Return all packs with OK status on success
}

// main is the entry point of the application.
func main() {
     r := InitRouter()             // Initialize HTTP router with routes and middleware setup.
     r.Run(":8080")                // Start listening on port 8080 for incoming requests.
}