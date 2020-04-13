package main

// interface as set
var exists = struct{}{} // take 0 byte

func NewSet() *set {
    s := &set{}
    s.m = make(map[interface{}]struct{})
    return s
}

// emulate set with struct of a map
type set struct {
    m map[interface{}]struct{}
}

// Add items to the set
func (s *set) Add(items ...interface{}){
    for _, item := range items {
        s.m[item] = exists
    }
}

// Remove item from set
func (s *set) Remove(items ...interface{}){
    for _, item := range items {
        delete(s.m, item)
    }
}

// Check if item exist
func (s *set) Has(items ...interface{}) bool{
    has := true
    for _, item := range items {
        if _,has = s.m[item]; !has{
            break
        }
    }
    return has
}

// Remove all items from set
func (s *set) Clear() {
    s.m = make(map[interface{}]struct{})
}

// Calculate set size
func (s *set) Size() int {
    return len(s.m)
}

// check set is empty
func (s *set) IsEmpty() bool {
    return s.Size() == 0
}

// create a new copy of set s
func (s *set) Copy() *set {
    n := NewSet()
    for item := range s.m {
        n.Add(item)
    }
    return n
}

// iterate through set items
func (s *set) Each(f func(item interface{}) bool) {
    for item := range s.m{
        if !f(item){
            break
        }
    }
}

// set union operation
func Union(sets ...*set) *set {
    u := NewSet()
    for _, set := range sets {
        for item := range set.m {
            u.Add(item)
        }
    }
    return u
}

// set intersection operation
func Intersect(sets ...*set) *set {
    all := Union(sets...)
    for item := range all.m {
        for _, set := range sets {
            if !set.Has(item) {
                all.Remove(item)
            }
        }

    }
    return all
}

