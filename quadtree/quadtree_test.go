package quadtree // Declares that this file is part of the "quadtree" package

import "testing" // Imports Go's standard testing framework

// TestNewQuadTree (You have a typo here, it should be TestNewQuadTree)
// This function tests the NewQuadTree "constructor".
func TestNewQaudTree(t *testing.T) {
	// We define a Boundary and capacity for our test
	boundary := Boundary{X: 0, Y: 0, Width: 100, Height: 100}
	capacity := 4

	// Call the function we want to test
	qt := NewQuadTree(boundary, capacity)

	// Test 1: The tree must not be 'nil' (it must have been created)
	if qt == nil {
		// t.Fatal immediately stops this test if it fails
		t.Fatal("NewQuadTree() returned nil")
	}

	// Test 2: The capacity must be what we set
	if qt.capacity != capacity {
		// t.Errorf reports an error but continues the test
		t.Errorf("Incorrect capacity: expected %d, got %d", capacity, qt.capacity)
	}

	// Test 3: The 'points' slice must be initialized (not 'nil')...
	if qt.points == nil {
		t.Error("qt.points is nil, but it should be an initialized slice")
	}

	// Test 4: ...but it must be empty (length 0)
	if len(qt.points) != 0 {
		t.Errorf("Wrong 'points' length: expected 0, got %d", len(qt.points))
	}
}

// TestQuadTreeInsert verifies the insertion logic,
// specifically the subdivision and redistribution of points.
func TestQuadTreeInsert(t *testing.T) {
	// We create a tree with capacity 2. This allows us to force
	// a subdivision on the third insert.
	qt := NewQuadTree(Boundary{X: 0, Y: 0, Width: 100, Height: 100}, 2)

	// Create two points that will go into different quadrants
	p1 := &Point{X: -50, Y: 50, Data: "p1 (NW)"} // North-West
	p2 := &Point{X: 50, Y: 50, Data: "p2 (NE)"}  // North-East

	// --- Test 1: Insertion below capacity ---
	qt.Insert(p1)
	qt.Insert(p2)

	// At this point, the tree must have 2 points in the root node...
	if len(qt.points) != 2 {
		t.Fatalf("Expected 2 points before splitting, found %d", len(qt.points))
	}
	// ...and it must NOT have subdivided yet (children must be 'nil')
	if qt.northWest != nil {
		t.Fatal("Subdivision occurred too soon")
	}

	// --- Test 2: Insertion that forces subdivision ---
	// We insert the third point (capacity is 2)
	p3 := &Point{X: -50, Y: -50, Data: "p3 (SW)"} // South-West
	qt.Insert(p3)

	// --- Test 3: Verification of redistribution ---
	// The parent node must now be empty (points were redistributed)
	if len(qt.points) != 0 {
		t.Errorf("Parent node not emptied after splitting. Found %d points", len(qt.points))
	}
	// The children must have been created
	if qt.northWest == nil {
		t.Fatal("Subdivision did not occur")
	}

	// Check that the 3 points (p1, p2, p3) ended up in the correct children
	if len(qt.northWest.points) != 1 || qt.northWest.points[0].Data != "p1 (NW)" {
		t.Error("Point p1 not found in the North-West quadrant")
	}
	if len(qt.northEast.points) != 1 || qt.northEast.points[0].Data != "p2 (NE)" {
		t.Error("Point p2 not found in the North-East quadrant")
	}
	if len(qt.southWest.points) != 1 || qt.southWest.points[0].Data != "p3 (SW)" {
		t.Error("Point p3 not found in the South-West quadrant")
	}
	// The South-East quadrant must still be empty
	if len(qt.southEast.points) != 0 {
		t.Error("Points found in the South-East quadrant, which should be empty")
	}

	// --- Test 4: Insertion into an already subdivided tree ---
	// We insert a fourth point. It must go directly into the correct child.
	p4 := &Point{X: 50, Y: -50, Data: "p4 (SE)"} // South-East
	qt.Insert(p4)
	if len(qt.southEast.points) != 1 || qt.southEast.points[0].Data != "p4 (SE)" {
		t.Error("Point p4 not inserted correctly in child South-East")
	}
}

// TestQuadTreeQuery verifies that the search function (Query)
// finds the correct points in different areas.
func TestQuadTreeQuery(t *testing.T) {
	// Create a test tree and populate it with 5 points
	// This will force a subdivision (capacity 2)
	qt := NewQuadTree(Boundary{X: 0, Y: 0, Width: 100, Height: 100}, 2)

	p1 := &Point{X: -50, Y: 50, Data: "p1 (NW)"}
	p2 := &Point{X: 50, Y: 50, Data: "p2 (NE)"}
	p3 := &Point{X: -50, Y: -50, Data: "p3 (SW)"}
	p4 := &Point{X: 50, Y: -50, Data: "p4 (SE)"}
	p5 := &Point{X: 60, Y: 60, Data: "p5 (NE, extra)"} // A second point in NE

	qt.Insert(p1)
	qt.Insert(p2)
	qt.Insert(p3) // Forces subdivision
	qt.Insert(p4)
	qt.Insert(p5)

	// --- Test 1: Search in a single quadrant (North-East) ---
	// This area only covers the NE quadrant (from (0,0) to (100,100))
	searchNE := &Boundary{X: 50, Y: 50, Width: 50, Height: 50}
	foundNE := qt.Query(searchNE)

	// We should find p2 and p5
	if len(foundNE) != 2 {
		t.Fatalf("NE query: 2 points expected, %d found", len(foundNE))
	}

	// --- Test 2: Search in an empty area ---
	// This area is in the center, where we didn't insert points
	searchEmpty := &Boundary{X: 0, Y: 0, Width: 10, Height: 10}
	foundEmpty := qt.Query(searchEmpty)

	// We should find nothing
	if len(foundEmpty) != 0 {
		t.Fatalf("Query Empty: 0 points expected, %d found", len(foundEmpty))
	}

	// --- Test 3: Search the entire map ---
	searchAll := &Boundary{X: 0, Y: 0, Width: 100, Height: 100}
	foundAll := qt.Query(searchAll)

	// We should find all 5 points
	if len(foundAll) != 5 {
		t.Fatalf("Query All: 5 points expected, %d found", len(foundAll))
	}

	// --- Test 4: Search covering two quadrants (South) ---
	// This area covers SW and SE
	searchSouth := &Boundary{X: 0, Y: -50, Width: 100, Height: 50}
	foundSouth := qt.Query(searchSouth)

	// We should find p3 and p4
	if len(foundSouth) != 2 {
		// Using the user's original Italian error string, translated
		t.Fatalf("Query South: 2 points expected (p3, p4), found %d", len(foundSouth))
	}
}
