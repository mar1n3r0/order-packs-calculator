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

	c.calculatePacksRecursive(c.items, 0, 0)
}

// calculatePacksRecursive is a helper function that performs the actual calculation recursively.
func (c *calculator) calculatePacksRecursive(items int, packIndex int, totalItems int) {
	if items <= 0 || packIndex >= len(c.packs) { 
	    return 
    }

	pack := c.packs[packIndex]

	packCount := items / pack.Size

    // choose larger pack when overshooting
    if len(c.packs) - 1 > packIndex {
        if (items + 1) / pack.Size > 0 && items % c.packs[packIndex+1].Size != 0 {            
            packCount = (items + 1) / pack.Size
        }   
    } 

	if packCount > 0 { 
        totalItems += packCount * pack.Size
        
	    c.packQuantities = append(c.packQuantities, PackQuantity{ 
	        Pack: pack.Size,
	        Quantity: packCount,
	    }) 

	    items -= packCount * pack.Size
        
    }

	if items > 0 { 
        
	    if packIndex < len(c.packs)-1 { 
            
	        c.calculatePacksRecursive(items, packIndex+1, totalItems)
	    } else {
            
	        nextPackSize := c.packs[packIndex].Size
            pq := []PackQuantity{}
            var n int

            // check if smaller packs can match closer desired amount of items
            if items > 0 {
                for i := 1; i * nextPackSize <= totalItems + nextPackSize; i++ {
                    if i * nextPackSize >= c.items {
                        if c.items - i * nextPackSize == 0 || c.items - i * nextPackSize < c.items - totalItems + nextPackSize {
                            n = i
                        }
                    }
                }
            }

            // compare which pack combination is closer to requirements
            if n > 0 && n * nextPackSize - c.items < (totalItems + nextPackSize) - c.items {
                var multiplier int
                if nextPackSize * n > c.packs[packIndex-1].Size && c.packs[packIndex-1].Size % nextPackSize == 0 {
                    multiplier =  c.packs[packIndex-1].Size / nextPackSize
                    
                    pq = append(pq, PackQuantity{ 
                        Pack: c.packs[packIndex-1].Size,
                        Quantity: 1,
                    })
                }
                if multiplier > 0 {
                    n = n - multiplier
                }
                pq = append(pq, PackQuantity{ 
                    Pack: nextPackSize,
                    Quantity: n,
                })

                // check if we can merge duplicate packs into a bigger one
                cm := c.canMerge(pq)
                
                if cm {
                    c.merge(pq)
                } else {
                    c.packQuantities = pq
                }
            } else {
                for i, pq := range c.packQuantities {
                    if pq.Pack == nextPackSize {
                        c.packQuantities[i].Quantity++
                        items = 0
                    }
                }
                if items > 0 {
                    c.packQuantities = append(c.packQuantities, PackQuantity{ 
                        Pack: nextPackSize,
                        Quantity: 1,
                    }) 
                }

                // check if we can merge duplicate packs into a bigger one
                cm := c.canMerge(c.packQuantities)
                if cm {
                    c.merge(c.packQuantities)
                }

                // do a final result check with difference and remainder
                var finalResultCheck int
                for _,p := range c.packQuantities {
                    finalResultCheck += p.Pack * p.Quantity
                }

                difference := finalResultCheck - c.items
                index := finalResultCheck / c.packs[packIndex].Size
                remainder := finalResultCheck % c.packs[packIndex].Size

                
                

                var pqs PackQuantity
                for i := 1; i <= index; i++ {
                    pqs = PackQuantity{
                        Pack: c.packs[packIndex].Size,
                        Quantity: i,
                    }
                }
                if remainder > 0 && remainder < difference {
                    c.packQuantities = nil
                    c.packQuantities = append(c.packQuantities, pqs)
                }
            }

	        items = 0  
	    }
    }
}

// canMerge checks for duplicate packs and if they can be merged into a bigger one available
func (c *calculator) canMerge(pq []PackQuantity) bool {
    var aggregate int
    var index int
    
    for n, p := range pq {
        if n > 0 && p.Pack == pq[n-1].Pack {
            aggregate = p.Pack + p.Pack
            index = p.Quantity + pq[n].Quantity
        } else if p.Quantity > 1 {
            for _, pp := range c.packs {
                if pp.Size / p.Pack * p.Quantity > 0 {
                    return true
                }
            }
        }
    }

    if aggregate > 0 && index > 0 {
        for _, pp := range c.packs {
            if pp.Size / aggregate > 0 {
                return true
            }
        }
    }

    return false
}

// merge does the actual merging of duplicate packs into a bigger one available
func (c *calculator) merge(pq []PackQuantity) {
    var aggregate int
    var remainder int

    for n, p := range pq {
        if n > 0 && p.Pack == pq[n-1].Pack {
            
            aggregate = p.Pack + p.Pack
        } else if p.Quantity > 1 {
            
            for _, pp := range c.packs {
                if pp.Size > p.Pack && pp.Size / p.Pack > 0 && pp.Size <= p.Pack * p.Quantity {
                    
                    aggregate = pp.Size
                    remainder = p.Pack * p.Quantity - pp.Size 
                    
                    
                }
            }
        }
    }

    for n, p := range pq {
        
        
        if pq[n].Pack == aggregate {
            
            pq[n].Quantity++
        } else if pq[n].Pack * pq[n].Quantity == aggregate {
            for _, pp := range c.packs {
                if pp.Size == aggregate {
                    
                    pq = append(pq[:n], pq[n+1:]...)
                    pq = append(pq, PackQuantity{
                        Pack: pp.Size,
                        Quantity: 1,
                    })
                }
            }
        }
        if pq[n].Pack == remainder {
            pq[n].Quantity = remainder / p.Pack
            
        }
    }

    c.packQuantities = pq
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