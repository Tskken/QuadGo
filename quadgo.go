package quadgo

// quadrant type for iota child quadrants.
type quadrant uint8

// constant values for child quadrants.
const (
	bottomLeft quadrant = iota
	bottomRight
	topLeft
	topRight
)

// Option function type for setting the Options of a new tree.
type Option func(*Options)

// Options struct to old the new trees options that will be set by the Option functions.
type Options struct {
	Width, Height         float64
	MaxEntities, MaxDepth int
}

// defaultOptions for QuadGo
var defaultOption = &Options{
	Width:       1024,
	Height:      768,
	MaxEntities: 10,
	MaxDepth:    2,
}

// SetBounds sets the bounds of the new tree.
func SetBounds(width, height float64) Option {
	return func(o *Options) {
		o.Width = width
		o.Height = height
	}
}

// SetMaxEntities sets the max number of entities per node for the new tree.
func SetMaxEntities(maxEntities int) Option {
	return func(o *Options) {
		o.MaxEntities = maxEntities
	}
}

// SetMaxDepth sets the max depth that the tree can split to.
func SetMaxDepth(maxDepth int) Option {
	return func(o *Options) {
		o.MaxDepth = maxDepth
	}
}

// QuadGo - Base Quadtree data structure.
type QuadGo struct {
	*node

	maxDepth int
}

// New creates the basic QuadGo instance.
//
// You can give New() any number of Option functions to change the desired settings of the tree.
// The main function to set would be SetBound(width, height). This function will set the new trees root bounds to be
// the given width and height. To see other Options check out the Godoc's.
//
// If no Options are given the default is a bounds of 1024x768, max entities per node of 10, and max depth of 2.
func New(ops ...Option) *QuadGo {
	// copy defaults
	o := defaultOption

	// update for any given options
	for _, op := range ops {
		op(o)
	}

	// Return new QuadGo instance
	return &QuadGo{
		node: &node{
			parent:   nil,
			bounds:   NewBound(0, 0, o.Width, o.Height),
			entities: make(Entities, 0, o.MaxEntities),
			children: make(nodes, 0, 4),
			depth:    0,
		},
		maxDepth: o.MaxDepth,
	}
}

// Insert takes the new entities Min and Max xy coordinates and inserts it in to the quadtree.
// It also takes any number of objects of any type as extra data to store with in the entity with the given Bound.
//
// The Object is any data type you may want to store in the entity.
// When searching the tree it will return an entity which holds the objects provided.
func (q *QuadGo) Insert(minX, minY, maxX, maxY float64, objs ...interface{}) {
	// insert in to quadtree
	q.insert(NewEntity(minX, minY, maxX, maxY, objs), q.maxDepth)
}

// InsertEntity inserts an entity in to the quadtree.
//
// This can be used as a second Option over Insert if you want to create your Entity before adding it to the quadtree.
func (q *QuadGo) InsertEntity(entity *Entity) {
	q.insert(entity, q.maxDepth)
}

// Remove removes the given Entity from the quadtree.
func (q *QuadGo) Remove(entity *Entity) {
	// remove from quadtree
	q.remove(entity)
}

// RetrieveFromPoint returns a list of entities that are stored in the node that the given point can be contained within.
func (q *QuadGo) RetrieveFromPoint(point Point) Entities {
	// retrieve entities for quadtree
	return q.retrieve(point)
}

// RetrieveFromBound returns a list of entities that are stored in a node that the given bound's center point can be contained within.
func (q *QuadGo) RetrieveFromBound(bound Bound) Entities {
	return q.retrieve(bound.Center)
}

// IsEntity checks if a given entity exists within the tree.
func (q *QuadGo) IsEntity(entity *Entity) bool {
	return q.isEntity(entity)
}

// IsIntersectPoint takes a point and returns if that point intersects any entity within the tree.
func (q *QuadGo) IsIntersectPoint(point Point) bool {
	entities := q.retrieve(point)
	// check all entities returned from retrieve for if they intersect
	for i := range entities {
		// check for intersect
		if entities[i].IsIntersectPoint(point) {
			return true
		}
	}
	return false
}

// IsIntersectBound take a bound and returns if that bound intersects any entity within the tree.
func (q *QuadGo) IsIntersectBound(bound Bound) bool {
	// get entities from a node that bound.Center can fit in
	entities := q.retrieve(bound.Center)

	// check all entities returned from retrieve for if they intersect
	for i := range entities {
		// check for intersect
		if entities[i].IsIntersectBound(bound) {
			return true
		}
	}
	return false
}

// IntersectsPoint takes a point and returns all entities that that point intersects with within the tree.
func (q *QuadGo) IntersectsPoint(point Point) (intersects Entities) {
	// get entities from a node that the point can fit in
	entities := q.retrieve(point)

	// check all entities returned from retrieve for if they intersect
	for i := range entities {
		// add to list if they intersect
		if entities[i].IsIntersectPoint(point) {
			intersects = append(intersects, entities[i])
		}
	}
	return
}

// IntersectsBound takes a bound and returns all entities that that bound intersects with within the tree.
func (q *QuadGo) IntersectsBound(bound Bound) (intersects Entities) {
	// get entities from a node that the bound.Center can fit in
	entities := q.retrieve(bound.Center)

	// check all entities returned from retrieve for if they intersect
	for i := range entities {
		// add to list if they intersect
		if entities[i].IsIntersectBound(bound) {
			intersects = append(intersects, entities[i])
		}
	}
	return
}

// list of node
type nodes []*node

// node is the container that holds the branch and leaf data for the tree.
type node struct {
	parent   *node
	bounds   Bound
	entities []*Entity
	children []*node
	depth    int
}

// retrieve finds all of the entities with in a the nodes that the given point can fit within.
func (n *node) retrieve(point Point) Entities {
	// check if you are at a leaf node
	if len(n.children) > 0 {
		// get quadrant the point fits in and go to that next node
		return n.getQuadrant(point).retrieve(point)
	} else {
		// return entities from leaf
		return n.entities
	}
}

// insert inserts a given Entity in to the quadtree.
func (n *node) insert(entity *Entity, maxDepth int) {
	// Check if you are on a leaf node
	if len(n.children) > 0 && n.depth <= maxDepth {
		// get the next node that the given entity fits in and attempt to insert it
		n.getQuadrant(entity.Center).insert(entity, maxDepth)
	} else {
		// Check if a split is needed
		if len(n.entities)+1 > cap(n.entities) && n.depth < maxDepth {
			// split node in to child nodes and add this nodes entities in to the appropriate child nodes
			n.split(append(n.entities, entity), maxDepth)
		} else {
			// Add Entity to node
			n.entities = append(n.entities, entity)
		}
	}
}

// remove removes the given Entity from the quadtree.
func (n *node) remove(entity *Entity) {
	// check if we are on a leaf node
	if len(n.children) > 0 {
		// get the next node that the given entity fits in and attempt to remove it
		n.getQuadrant(entity.Center).remove(entity)
	} else {
		// check the entities in leaf for given entity
		for i := range n.entities {
			// check if given Entity is the same as node Entity
			if n.entities[i] == entity {
				// check if removal would make the leaf have no entities
				if len(n.entities) == 1 {
					// set node entities to an empty slice
					n.entities = make(Entities, 0, cap(n.entities))
				} else {
					// remove Entity from node
					n.entities = append(n.entities[:i], n.entities[i+1:]...)
				}

				// check if children can be collapsed in to parent node
				n.parent.collapse()
			}
		}
	}
}

// collapse checks if a parent's children hold less entities then the set maxEntities count.
// if the count is less then maxEntities it collapses all children in to the parent node, copying
// all of there entities to the parent node and setting the children to new empty slices.
func (n *node) collapse() {
	// create base counter for children entity count
	eCount := 0

	// count up total entities in children
	for i := range n.children {
		eCount += len(n.children[i].entities)
	}

	// check if the total number of entities in the nodes children is less then the
	// Max number of entities allowed in a node
	if eCount < cap(n.entities) {
		// move children entities to parent node
		for i := range n.children {
			n.entities = append(n.entities, n.children[i].entities...)
		}

		// reset children
		n.children = make(nodes, 0, 4)
	}
}

// isEntity returns if a given entity exists in the tree.
func (n *node) isEntity(entity *Entity) bool {
	// get entities from a node that the entity.Center can fit in
	entities := n.retrieve(entity.Center)

	// check each entity for if it is equal to given entity
	for i := range entities {
		// check if given Entity equals given entity
		if entities[i] == entity {
			return true
		}
	}

	return false
}

// split creates the children for a node by subdividing the nodes boundaries in to 4 even quadrants. It then
// adds the nodes entities to the new child nodes.
func (n *node) split(entities Entities, maxDepth int) {
	// Bottom Left child node
	n.children = append(n.children, &node{
		parent:   n,
		bounds:   NewBound(n.bounds.Min.X, n.bounds.Min.Y, n.bounds.Center.X, n.bounds.Center.Y),
		entities: make([]*Entity, 0, cap(n.entities)),
		children: make([]*node, 0, 4),
		depth:    n.depth + 1,
	})

	// Bottom Right child node
	n.children = append(n.children, &node{
		parent:   n,
		bounds:   NewBound(n.bounds.Center.X, n.bounds.Min.Y, n.bounds.Max.X, n.bounds.Center.Y),
		entities: make([]*Entity, 0, cap(n.entities)),
		children: make([]*node, 0, 4),
		depth:    n.depth + 1,
	})

	// Top Left child node
	n.children = append(n.children, &node{
		parent:   n,
		bounds:   NewBound(n.bounds.Min.X, n.bounds.Center.Y, n.bounds.Center.X, n.bounds.Max.Y),
		entities: make([]*Entity, 0, cap(n.entities)),
		children: make([]*node, 0, 4),
		depth:    n.depth + 1,
	})

	// Top Right child node
	n.children = append(n.children, &node{
		parent:   n,
		bounds:   NewBound(n.bounds.Center.X, n.bounds.Center.Y, n.bounds.Max.X, n.bounds.Max.Y),
		entities: make([]*Entity, 0, cap(n.entities)),
		children: make([]*node, 0, 4),
		depth:    n.depth + 1,
	})

	// loop through all entities to add them to there appropriate child node
	for i := range entities {
		// get the next node that the given entity fits in and insert it
		n.getQuadrant(entities[i].Center).insert(entities[i], maxDepth)
	}

	// clear entities for branch node
	n.entities = make(Entities, 0, cap(n.entities))
}

// getQuadrant returns the nodes child node that the given point fits within.
func (n *node) getQuadrant(point Point) *node {
	switch {
	// bottom left node check
	case point.X <= n.bounds.Center.X && point.Y <= n.bounds.Center.Y:
		return n.children[bottomLeft]
	// bottom right node check
	case point.X > n.bounds.Center.X && point.Y <= n.bounds.Center.Y:
		return n.children[bottomRight]
	// top left node check
	case point.X <= n.bounds.Center.X && point.Y > n.bounds.Center.Y:
		return n.children[topLeft]
	// top right node check
	case point.X > n.bounds.Center.X && point.Y > n.bounds.Center.Y:
		return n.children[topRight]
	// default should never trigger as there should never be a point were the given point can not fit in any child node
	default:
		return nil
	}
}
