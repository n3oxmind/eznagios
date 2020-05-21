package main

import (
	//	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Check if attr (type string) exist in  def/(map[string]*set)
func (d def) attrExist(attrName string) bool {
    if _, exist := d[attrName]; exist {
        return true
    }
    return false
}

// Check if offset already exist
func (o offset) OffsetExist(id string) bool {
    if _, exist := o[id]; exist {
        return true
    }
    return false
}

// Check if offset (type int) already exist
func (o offset) tmplExist(id string) bool {
    if _, exist := o[id]; exist {
        return true
    }
    return false
}

// Check if map is empty
func (o *offset) isEmpty() bool {
    if len(*o) == 0 {
        return true
    }
    return false
}

// Check if slice (o) contains a list of elements (flags)
func (o *boolFlag) HasAll(flags ...string) bool {
    if *o == nil {
        return false
    }
    flen := len(flags)
    matchCounter := 0
    for _, f := range flags {
        for _, v := range *o {
            if v == f {
                matchCounter += 1
                break
            }
        }
    }
    if flen == matchCounter {
        return true
    }else {
        return false
    }
}
// check if all items exist in a slice
func (o *attrVal) HasAll(flags ...string) bool {
    if *o == nil {
        return false
    }
    flen := len(flags)
    matchCounter := 0
    for _, f := range flags {
        for _, v := range *o {
            if v == f {
                matchCounter += 1
                break
            }
        }
    }
    if flen == matchCounter {
        return true
    }else {
        return false
    }
}

// check if []string contain and items other than the provided item, or the item does not exist at all
func (o *attrVal) HasOnly(item string) bool {
    for _, v := range *o {
        if v != item {
            return false
        }
    }
    return true
}

// Copy offset ( map[int]string)
func CopyMapInt(m offset) *offset {
    newMap := make(offset)
    for k, v := range m {
        newMap[k] = v
    }
    return &newMap
}

// Prefix item with '!' for execlution
func AddEP(s attrVal) *attrVal {
    EPSlice := make(attrVal,len(s))
    for i,item := range s {
        EPSlice[i] = "!"+item
    } 
    return &EPSlice
}

// Remove item from a slice based on index
func RemoveItemByIndex(s *attrVal, i int) {
    // handle index out of range, return same slice if index out of range
    length := len(*s)
    if i > length {
        return
    }
    (*s)[i] = (*s)[len(*s)-1]
    (*s)[len(*s)-1] = ""
    *s = (*s)[:len(*s)-1]
}

// find items and !items of an object attribute  and return their indices
func (a *attrVal) FindItemIndex(AttrVals ...string) *[]int {
    idx := []int{}
    for _, v := range AttrVals {
        for i, e := range *a {
            if has, _ := regexp.MatchString("^"+e+"$", v); has {
                idx = append(idx, i)
            }
            if has, _ := regexp.MatchString("^"+e+"$", "!"+v); has {
                idx = append(idx, i)
            }
        }
    }
    // return reverse order index
    sort.Sort(sort.Reverse(sort.IntSlice(idx)))
    return &idx
}

// Remove item from a slice based on val
func RemoveItemByName(s attrVal, v string) {
    // handle index out of range, return same slice if index out of range
    for i, val := range s {
        if val == v {
            RemoveItemByIndex(&s, i)
        }
    }
}

// Convert offset to []int -> holds indices and []string -> holds attr value
func (o *offset) ToSlice() (id []string, val attrVal) {
    for k, v := range *o {
        id = append(id, k)
        for _, item := range v {
            val = append(val, item)
        }
    }
    return id,val
}

// Convert []string to string
func (o attrVal) ToString() string {
    str := ""
    str = strings.Join(o, ",")
    return str
}
// Convert []string to flag
func ToFlag(slc *[]string) *boolFlag {
    if len(*slc) == 0 {
        return nil
    }

    flags := boolFlag{}
    for _, v := range *slc {
        flags = append(flags, v)
    }
    return &flags
}
