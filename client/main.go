package main

import (
	"log"
	"net/http"
	"io"
	"sort"
	"bytes"
	"strconv"
	"encoding/json"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// calculator is a component that displays packs and calculates packs for orders. 
// A component is a customizable, independent, and reusable UI element. 
// It is created by embedding app.Compo into a struct.
type calculator struct {
	app.Compo
	packs          []Pack          // List of available packs
	currentPack    Pack            // Currently selected pack
	items          int             // Number of items to pack
	packQuantities []PackQuantity   // Quantities of each pack size used in the calculation
}

// Pack represents a single pack with an ID and size.
type Pack struct {
	ID    string `mapstructure:"id" json:"id" validate:"uuid_rfc4122"` // Unique identifier for the pack
	Size  int    `mapstructure:"size" json:"size" validate:"uuid_rfc4122"` // Size of the pack
}

// PackQuantity holds the quantity of a specific pack size.
type PackQuantity struct {
	Pack     int `mapstructure:"pack" json:"pack" validate:"uuid_rfc4122"`     // Size of the pack
	Quantity int `mapstructure:"quantity" json:"quantity" validate:"uuid_rfc4122"` // Number of packs of this size
}

// OnMount fetches the available packs when the component mounts.
func (c *calculator) OnMount(ctx app.Context) {
	c.getPacks(ctx)
}

// getPacks retrieves the list of packs from the server.
func (c *calculator) getPacks(ctx app.Context) {
	ctx.Async(func() {
		r, err := http.Get("http://localhost:8080/packs") // Fetch packs from server
		if err != nil {
			app.Log(err)
			return
		}
		defer r.Body.Close()

		resp, err := io.ReadAll(r.Body) // Read response body
		if err != nil {
			app.Log(err)
			return
		}

		var packs []Pack
		err = json.Unmarshal([]byte(resp), &packs) // Unmarshal JSON response into packs slice
		if err != nil {
			log.Fatalf("Unable to marshal JSON due to %s", err)
		}

		sort.Slice(packs, func(i, j int) bool { // Sort packs by size in descending order
			return packs[i].Size > packs[j].Size
		})

		ctx.Dispatch(func(ctx app.Context) { // Update component state with fetched packs
			c.packs = packs
		})
	})
}

// postPack sends a new pack to the server.
func (c *calculator) postPack(ctx app.Context, pack Pack) {
	ctx.Async(func() {
		payload, err := json.Marshal(map[string]interface{}{
			"size": pack.Size,
		})
		if err != nil {
			log.Fatal(err)
		}

		client := &http.Client{}
		url := "http://localhost:8080/packs"

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload)) // Create POST request
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			log.Fatal(err)
		}

		resp, err := client.Do(req) // Send request to server
		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()

		_, err = io.ReadAll(resp.Body) // Read response body
		if err != nil {
			log.Fatal(err)
		}

        c.getPacks(ctx) // Refresh packs after adding new one
    })
}

// putPack updates an existing pack on the server.
func (c *calculator) putPack(ctx app.Context, pack Pack) {
	ctx.Async(func() {
        payload, err := json.Marshal(map[string]interface{}{
            "id":   pack.ID,
            "size": pack.Size,
        })
        if err != nil {
            log.Fatal(err)
        }

        client := &http.Client{}
        url := "http://localhost:8080/packs/" + pack.ID

        req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(payload)) // Create PUT request
        req.Header.Set("Content-Type", "application/json")
        if err != nil {
            log.Fatal(err)
        }

        resp, err := client.Do(req) // Send request to server
        if err != nil {
            log.Fatal(err)
        }

        defer resp.Body.Close()

        _, err = io.ReadAll(resp.Body) // Read response body
        if err != nil {
            log.Fatal(err)
        }

        c.getPacks(ctx) // Refresh packs after updating one
    })
}

// deletePack removes a pack from the server based on its ID.
func (c *calculator) deletePack(ctx app.Context, e app.Event) {
	id := ctx.JSSrc().Get("id").String() // Get ID from event source
	ctx.Async(func() {
        client := &http.Client{}
        url := "http://localhost:8080/packs/" + id

        req, err := http.NewRequest(http.MethodDelete, url, nil) // Create DELETE request
        req.Header.Set("Content-Type", "application/json")
        if err != nil {
            log.Fatal(err)
        }

        resp, err := client.Do(req) // Send request to server
        if err != nil {
            log.Fatal(err)
        }

        defer resp.Body.Close()

        _, err = io.ReadAll(resp.Body) // Read response body
        if err != nil {
            log.Fatal(err)
        }

        c.getPacks(ctx) // Refresh packs after deletion
    })
}

// setPack sets the current pack based on user input.
func (c *calculator) setPack(ctx app.Context, e app.Event) {
	id := ctx.JSSrc().Get("id").String() 
	c.currentPack.ID = id 
	sizeStr := ctx.JSSrc().Get("value").String() 
	size, err := strconv.Atoi(sizeStr) 
	if err != nil { 
	    log.Fatalf("Unable to convert size to int %s", err)
    } 
	c.currentPack.Size = size 
}

// setNewPack sets a new pack size based on user input.
func (c *calculator) setNewPack(ctx app.Context, e app.Event) { 
	sizeStr := ctx.JSSrc().Get("value").String() 
	size, err := strconv.Atoi(sizeStr) 
	if err != nil { 
	    log.Fatalf("Unable to convert size to int %s", err)
    } 
	c.currentPack.Size = size 
}

// setItems sets the number of items based on user input.
func (c *calculator) setItems(ctx app.Context, e app.Event) { 
	itemsStr := ctx.JSSrc().Get("value").String() 
	items, err := strconv.Atoi(itemsStr) 
	if err != nil { 
	    log.Fatalf("Unable to convert size to int %s", err)
    } 
	c.items = items 
}

// calculatePacks calculates how many packs are needed for the given number of items.
func (c *calculator) calculatePacks(ctx app.Context, e app.Event) { 
	c.packQuantities = nil 
	sort.Slice(c.packs, func(i, j int) bool { 
	    return c.packs[i].Size > c.packs[j].Size 
    })

	c.calculatePacksRecursive(c.items, 0)
}

// calculatePacksRecursive is a helper function that performs the actual calculation recursively.
func (c *calculator) calculatePacksRecursive(items int, packIndex int) { 
	if items <= 0 || packIndex >= len(c.packs) { 
	    return 
    }

	pack := c.packs[packIndex]

	packCount := items / pack.Size 

	if packIndex == (len(c.packs)-1) && items-pack.Size > 0 { 
	    pack.Size = c.packs[packIndex-1].Size 
    }

	if packCount > 0 { 
	    c.packQuantities = append(c.packQuantities, PackQuantity{ 
	        Pack: pack.Size,
	        Quantity: packCount,
	    }) 

	    items -= packCount * pack.Size 
    }

	if items > 0 { 
	    if packIndex < len(c.packs)-1 { 
	        c.calculatePacksRecursive(items, packIndex+1)
	    } else { 
	        nextPackSize := c.packs[packIndex].Size 

	        c.packQuantities = append(c.packQuantities, PackQuantity{ 
	            Pack: nextPackSize,
	            Quantity: 1,
	        }) 

	        items = 0  
	    }
    }
}

// updatePack updates the current selected pack.
func (c *calculator) updatePack(ctx app.Context, e app.Event) { 
	c.putPack(ctx, c.currentPack)
}

// createPack creates a new pack based on current input.
func (c *calculator) createPack(ctx app.Context, e app.Event) { 
	c.postPack(ctx, c.currentPack)
}

// Render defines how the component appears in the UI.
func (c *calculator) Render() app.UI { 
	return app.Div().Class("container text-center").Body( 
	    app.Div().Class("row align-items-start").Body( 
	        app.Div().Class("col").Body(  
	            app.H1().Class("w-auto p-3").Text("Order Packs Calculator"),  
	            app.Table().Class("table").Body(  
	                app.THead().Body(  
	                    app.Tr().Body(  
	                        app.Th().Class("text-start").Scope("col").Text("Pack Sizes"),  
	                    ),  
	                ),  
	                app.TBody().Body(  
	                    app.Range(c.packs).Slice(func(n int) app.UI {  
	                        return app.Tr().Body(  
                                app.Th().Scope("row").Body(  
                                    app.Div().Class("input-group flex-nowrap").Body(  
                                        app.Input().Type("number").ID(c.packs[n].ID).Class("form-control").Placeholder(strconv.Itoa(c.packs[n].Size)).OnChange(c.setPack),  
                                        app.Button().Class("btn btn-primary").Text("Update").OnClick(c.updatePack),  
                                        app.Button().ID(c.packs[n].ID).Class("btn btn-danger").Text("Delete").OnClick(c.deletePack),  
                                    ),  
                                ),  
                            )  
                        }),  
                        app.Th().Scope("row").Body(  
                            app.Div().Class("input-group flex-nowrap").Body(  
                                app.Input().Type("number").Class("form-control").OnChange(c.setNewPack),  
                                app.Button().Class("btn btn-success").Text("Add").OnClick(c.createPack),  
                            ),  
                        ),  
                    ),  
                ),  
            ),  
            app.Div().Class("col").Body(  
                app.H1().Class("w-auto p-3").Text("Calculate packs for order"),  
                app.Div().Class("input-group flex-nowrap").Body(  
                    app.Span().Class("input-group-text").Text("Items: "),  
                    app.Input().Type("number").Class("form-control").OnChange(c.setItems),  
                    app.Button().Class("btn btn-success").Text("Calculate").OnClick(c.calculatePacks),  
                ),  
                app.Table().Class("table").Body(  
                    app.THead().Body(  
                        app.Tr().Body(  
                            app.Th().Class("text-start").Scope("col").Text("Pack"),  
                            app.Th().Class("text-start").Scope("col").Text("Quantity"),  
                        ),  
                    ),   
                    app.TBody().Body(   
                        app.Range(c.packQuantities).Slice(func(n int) app.UI {   
                            return app.Tr().Body(   
                                app.Th().Scope("row").Body(   
                                    app.Div().Class("input-group flex-nowrap").Body(   
                                        app.Input().Type("number").Class("form-control").Placeholder(strconv.Itoa(c.packQuantities[n].Pack)),   
                                    ),   
                                ),   
                                app.Th().Scope("row").Body(   
                                    app.Div().Class("input-group flex-nowrap").Body(   
                                        app.Input().Type("number").Class("form-control").Placeholder(strconv.Itoa(c.packQuantities[n].Quantity)),   
                                    ),   
                                ),   
                            )   
                        }),   
                    ),   
                ),   
            ),   
        ),   
    )   
} 

// The main function is the entry point where the application is configured and started.
// It is executed in two different environments: a client (the web browser)
// and a server.
func main() {    
	app.Route("/", func() app.Composer { return &calculator{} }) 

	app.RunWhenOnBrowser() 

	http.Handle("/", &app.Handler{    
    	Name: "Order Packs Calculator",    
    	Description: "Display packs and calculate packs for orders",    
    	Styles: []string{    
        	"https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css",    
    	},    
    	Scripts: []string{    
        	"https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/js/bootstrap.bundle.min.js",    
    	},    
    })    

	if err := http.ListenAndServe(":5000", nil); err != nil {    
    	log.Fatal(err)    
    }    
}