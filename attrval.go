package main

import (
    "regexp"
)
type attrVal []string

// Add items to the set
func (s *attrVal) Add(items ...string){
    for _, item := range items {
        *s = append(*s, item)
    }
}

// Remove items from slice if exist
func (s *attrVal) Remove(items ...string){
    for i, v := range *s {
        for _, item := range items{
            if v == item {
                (*s)[i] = (*s)[len(*s)-1]
                (*s) = (*s)[:len(*s)-1]
            }
        }
    }
}

// Check if item exist in a slice
func (s *attrVal) Has(item string) bool{
    for _, v := range *s {
        if item ==  v {
            return true
        }
    }
    return false
}

// Check if item exist using regex
func (s *attrVal) RegexHas(item string) bool{
    for _,v := range *s {
        if has, _ := regexp.MatchString(v, item); has {
            return true
        }
    }
    return false
}

// Check if an item exist in both slices
func (s *attrVal) HasAny(items []string) bool{
    for _,v := range *s {
        for _, item := range items{
            if v == item{
                return true
            }
        }
    }
    return false
}
