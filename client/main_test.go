package main

import (
	"testing"
	"sort"
)

type calculatorTest struct {
	packs          []Pack
	currentPack    Pack
	items          int
	packQuantities []PackQuantity
}

// calculatePacks calculates how many packs are needed for the given number of items.
func (c *calculatorTest) calculatePacksTest() {
	c.packQuantities = nil
	// Sort packs in descending order by size
	sort.Slice(c.packs, func(i, j int) bool {
		return c.packs[i].Size > c.packs[j].Size
	})
	c.calculatePacksRecursiveTest(c.items, 0, 0)
}

// calculatePacksRecursive is a helper function that performs the actual calculation recursively.
func (c *calculatorTest) calculatePacksRecursiveTest(items int, packIndex int, totalItems int) {
	if items <= 0 || packIndex >= len(c.packs) { 
	    return 
    }

	pack := c.packs[packIndex]

	packCount := items / pack.Size

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
	        c.calculatePacksRecursiveTest(items, packIndex+1, totalItems)
	    } else {
	        nextPackSize := c.packs[packIndex].Size
            pq := []PackQuantity{}
            var n int

            if items > 0 {
                for i := 1; i * nextPackSize <= totalItems + nextPackSize; i++ {
                    if i * nextPackSize >= c.items {
                        if c.items - i * nextPackSize == 0 || c.items - i * nextPackSize < c.items - totalItems + nextPackSize {
                            n = i
                        }
                    }
                }
            }

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
                cm := c.canMerge(c.packQuantities)
                if cm {
                    c.merge(c.packQuantities)
                }

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

func (c *calculatorTest) canMerge(pq []PackQuantity) bool {
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

func (c *calculatorTest) merge(pq []PackQuantity) {
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

func TestCalculatePacks(t *testing.T) {
	tests := []struct {
		name     string
		packs    []Pack
		items    int
		expected []PackQuantity
	}{
		{
			name: "Order 1 item",
			packs: []Pack{
				{ID: "1", Size: 250},
				{ID: "2", Size: 500},
				{ID: "3", Size: 1000},
			},
			items:    1,
			expected: []PackQuantity{{Pack: 250, Quantity: 1}},
		},
		{
			name: "Order 250 items",
			packs: []Pack{
				{ID: "1", Size: 250},
				{ID: "2", Size: 500},
			},
			items:    250,
			expected: []PackQuantity{{Pack: 250, Quantity: 1}},
		},
		{
			name: "Order 251 items",
			packs: []Pack{
				{ID: "1", Size: 250},
				{ID: "2", Size: 500},
			},
			items:    251,
			expected: []PackQuantity{{Pack: 500, Quantity: 1}},
		},
		{
			name: "Order 501 items",
			packs: []Pack{
				{ID: "1", Size: 250},
				{ID: "2", Size: 500},
			},
			items:    501,
			expected: []PackQuantity{{Pack: 500, Quantity: 1}, {Pack: 250, Quantity: 1}},
		},
        {
            name:"Order with 12001 items",
            packs :[]Pack{
                {ID:"1",Size :5000},
                {ID:"2",Size :2000},
                {ID:"3",Size :250},
            },
            items :12001,
            expected :[]PackQuantity{{Pack :5000, Quantity :2},{Pack :2000, Quantity :1},{Pack :250, Quantity :1}},
        },
		{
            name:"Order of 14 items with packs of size 5 and 12",
            packs :[]Pack{
                {ID:"1", Size :5},
                {ID:"2", Size :12},
            },
            items :14,
            expected :[]PackQuantity{{Pack :5, Quantity :3}},
        },
    }

	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := calculatorTest{
                packs: tt.packs,
                items: tt.items,
            }
            c.calculatePacksTest()

            if len(c.packQuantities) != len(tt.expected) {
                t.Errorf("expected %v pack quantities, got %v", len(tt.expected), len(c.packQuantities))
                return
            }

            for i := range c.packQuantities {
                if c.packQuantities[i] != tt.expected[i] {
                    t.Errorf("expected %v, got %v", tt.expected[i], c.packQuantities[i])
                }
            }
        })
    }
}