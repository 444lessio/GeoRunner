package quadtree // Declares that this file belongs to the "quadtree" package

import (
	"sync" //Import concurrency package (Mutex)
)

type Point struct { // Points represents a single point in 2D space with associated data
	X    float64     // Longitude
	Y    float64     // Latitude
	Data interface{} //Generic Data (e.g ID Driver)
}

type Boundary struct { // Boundary defines a rectangular area using a center and "halves"
	X      float64 //Center X (Longitude)
	Y      float64 //Center Y (Latitude)
	Width  float64 //Half the width (from X on board)
	Height float64 // Half the Height (from Y on board)
}

// QuadTree is the primary data structure
// Contains a pointer to a Mutex to handle concurrency
type QuadTree struct {
	boundary Boundary // The area that this node covers
	capacity int      // Max number of points before splitting
	points   []*Point // Slice of pointers to points in this node

	// Pointer to the 4 children (initially nil)
	northWest *QuadTree
	northEast *QuadTree
	southWest *QuadTree
	southEast *QuadTree

	//Mutex to make the structure thread-safe
	//RWMutex is optimal: it allows multiple readings or a single writing
	mu sync.RWMutex
}

// NewQuadTree is the constructor for a QuadTree
func NewQuadTree(boundary Boundary, capacity int) *QuadTree {

	// Ensure the capacity is at least 1 to avoid logical errors
	if capacity < 1 {
		capacity = 1
	}

	// Initialize the QuadTree struct
	qt := &QuadTree{
		boundary: boundary,
		capacity: capacity,
		// Initialize the 'points' slice with a length of 0,
		// but with a pre-allocated capacity for efficiency.
		points: make([]*Point, 0, capacity),
	}

	return qt
}

// Contains checks if a point is within the boundary of this node
func (b *Boundary) Contains(p *Point) bool {
	// The logic uses a "semi-open" interval [min, max)
	// This means the 'min' boundary (West, South) is inclusive (>=)
	// and the 'max' boundary (East, North) is exclusive (<).
	// This prevents double-counting points that lie exactly on a shared border.
	return p.X >= (b.X-b.Width) && // West boundary (inclusive)
		p.X < (b.X+b.Width) && // East boundary (exclusive)
		p.Y >= (b.Y-b.Height) && // South boundary (inclusive)
		p.Y < (b.Y+b.Height) // North boundary (exclusive)
}

// subdivide creates four child quadrants for this node
func (qt *QuadTree) subdivide() {
	// Calculate the dimensions for the new children
	childWidth := qt.boundary.Width / 2
	childHeight := qt.boundary.Height / 2
	centerX := qt.boundary.X
	centerY := qt.boundary.Y

	// Create the boundary for the North-West child and initialize it
	nwBoundary := Boundary{X: centerX - childWidth, Y: centerY + childHeight, Width: childWidth, Height: childHeight}
	qt.northWest = NewQuadTree(nwBoundary, qt.capacity)

	// Create the boundary for the North-East child and initialize it
	neBoundary := Boundary{X: centerX + childWidth, Y: centerY + childHeight, Width: childWidth, Height: childHeight}
	qt.northEast = NewQuadTree(neBoundary, qt.capacity)

	// Create the boundary for the South-West child and initialize it
	swBoundary := Boundary{X: centerX - childWidth, Y: centerY - childHeight, Width: childWidth, Height: childHeight}
	qt.southWest = NewQuadTree(swBoundary, qt.capacity)

	// Create the boundary for the South-East child and initialize it
	seBoundary := Boundary{X: centerX + childWidth, Y: centerY - childHeight, Width: childWidth, Height: childHeight}
	qt.southEast = NewQuadTree(seBoundary, qt.capacity)
}

// Insert adds a point to the QuadTree
func (qt *QuadTree) Insert(p *Point) bool {

	// Acquire a Write Lock because we are modifying the tree
	qt.mu.Lock()
	// 'defer' ensures the lock is released when the function exits
	defer qt.mu.Unlock()

	// If the point is not within this node's boundary, reject it
	if !qt.boundary.Contains(p) {
		return false
	}

	// If this node is already subdivided (it's a "parent" node)...
	if qt.northWest != nil {
		// ...try to insert the point into one of its children recursively
		if qt.northWest.Insert(p) {
			return true
		}
		if qt.northEast.Insert(p) {
			return true
		}
		if qt.southWest.Insert(p) {
			return true
		}
		if qt.southEast.Insert(p) {
			return true
		}
		// If it fails to insert in all children (e.g., boundary issue), return failure
		return false
	}

	// If this is a "leaf" node (not subdivided), add the point to its list
	qt.points = append(qt.points, p)

	// Check if this node is now "full" and needs to be subdivided
	if len(qt.points) > qt.capacity {
		// Create the four children
		qt.subdivide()

		// --- Redistribution ---
		// Now that we have children, we must move all points
		// from this parent node down into the new children.
		oldPoints := qt.points
		// Clear the parent's point list
		qt.points = make([]*Point, 0, qt.capacity)

		// Loop over the old points and insert them into the children
		for _, pt := range oldPoints {
			// This recursive call will find the correct child
			if qt.northWest.Insert(pt) {
				continue
			}
			if qt.northEast.Insert(pt) {
				continue
			}
			if qt.southWest.Insert(pt) {
				continue
			}
			if qt.southEast.Insert(pt) {
				continue
			}
		}
	}
	// If we reached here, the point was successfully added to this leaf node
	return true
}

// Intersects checks if this boundary overlaps with another boundary
func (b *Boundary) Intersects(other *Boundary) bool {

	// Calculate the min/max coordinates for this boundary
	bMinX := b.X - b.Width
	bMaxX := b.X + b.Width
	bMinY := b.Y - b.Height
	bMaxY := b.Y + b.Height

	// Calculate the min/max coordinates for the other boundary
	otherMinX := other.X - other.Width
	otherMaxX := other.X + other.Width
	otherMinY := other.Y - other.Height
	otherMaxY := other.Y + other.Height

	// Use Axis-Aligned Bounding Box (AABB) intersection logic.
	// We check for all cases where they *do not* intersect.
	// If none of these are true, they must be intersecting.
	// The logic (<= and >=) must match the [min, max) logic from Contains().

	// if this boundary is entirely to the right of 'other'
	if bMinX >= otherMaxX {
		return false
	}
	// if this boundary is entirely to the left of 'other'
	if bMaxX <= otherMinX {
		return false
	}
	// if this boundary is entirely above 'other'
	if bMinY >= otherMaxY {
		return false
	}
	// if this boundary is entirely below 'other'
	if bMaxY <= otherMinY {
		return false
	}

	// If none of the "no-overlap" conditions are met, they must overlap
	return true
}

// Query is the public function to find points within a specific area
func (qt *QuadTree) Query(rangeRect *Boundary) []*Point {
	// Create an empty slice to store the results
	found := []*Point{}

	// Call the recursive helper function to populate the 'found' slice
	qt.queryRecursive(rangeRect, &found)

	// Return the populated slice
	return found

}

// queryRecursive is the internal helper that performs the recursive search
func (qt *QuadTree) queryRecursive(rangeRect *Boundary, found *[]*Point) {
	// Acquire a Read Lock (RLock).
	// This allows *multiple* queries to run at the same time,
	// but blocks if an Insert() is writing.
	qt.mu.RLock()
	// Release the Read Lock when the function exits
	defer qt.mu.RUnlock()

	// --- The Core Optimization ---
	// If the query area (rangeRect) doesn't even overlap
	// with this node's boundary, stop searching.
	// This "prunes" entire branches of the tree.
	if !qt.boundary.Intersects(rangeRect) {
		return
	}

	// If this is a "leaf" node (it has points, no children)...
	if qt.northWest == nil {
		// ...check every point in this node's list
		for _, p := range qt.points {
			// If the point is inside the query area...
			if rangeRect.Contains(p) {
				// ...add it to the results
				*found = append(*found, p)
			}
		}
		// We are a leaf, so we are done
		return
	}

	// If this is a "parent" node (it has children)...
	// ...recursively call queryRecursive on all four children.
	// They will each run the 'Intersects' check (step 1).
	qt.northWest.queryRecursive(rangeRect, found)
	qt.northEast.queryRecursive(rangeRect, found)
	qt.southWest.queryRecursive(rangeRect, found)
	qt.southEast.queryRecursive(rangeRect, found)
}

// Remove finds and removes a specific point from the tree
func (qt *QuadTree) Remove(p *Point) bool {

	// Acquire a Write Lock (we are modifying the tree)
	qt.mu.Lock()
	defer qt.mu.Unlock()

	// If the point can't exist in this boundary, return failure
	if !qt.boundary.Contains(p) {
		return false
	}

	// If this is a "parent" node (it has children)...
	if qt.northWest != nil {
		// ...recursively call Remove on the correct child
		if qt.northWest.Remove(p) {
			return true
		}
		if qt.northEast.Remove(p) {
			return true
		}
		if qt.southWest.Remove(p) {
			return true
		}
		if qt.southEast.Remove(p) {
			return true
		}
		// Point not found in any child
		return false
	}

	// If this is a "leaf" node...
	// ...find the exact index of the point in our list
	foundIndex := -1
	for i, pt := range qt.points {
		// We must check for an *exact* match (X, Y, and Data)
		if pt.Data == p.Data && pt.X == p.X && pt.Y == p.Y {
			foundIndex = i
			break
		}
	}

	// If the point was not found in our list
	if foundIndex == -1 {
		return false
	}

	// --- O(1) Slice Removal ---
	// "Swap and Pop" trick:
	// 1. Overwrite the element-to-remove with the *last* element in the slice
	qt.points[foundIndex] = qt.points[len(qt.points)-1]
	// 2. Reslice the slice to be one element shorter, dropping the (now duplicated) last element
	qt.points = qt.points[:len(qt.points)-1]

	return true
}
