package main

// Check if attr (type string) exist in  def/(map[string]*set)
func (d def) attrExist(attrName string) bool {
    if _, exist := d[attrName]; exist {
        return true
    }
    return false
}

// Check if offset (type int) already exist
func (o offset) OffsetExist(i int) bool {
    if _, exist := o[i]; exist {
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

// Copy offset ( map[int]string)
func CopyMapInt(m offset) *offset {
    newMap := make(offset)
    for k, v := range m {
        newMap[k] = v
    }
    return &newMap
}

// Prefix item with '!' for execlution
func AddEP(s []string) []string {
    EPSlice := make([]string,len(s))
    for i,item := range s {
        EPSlice[i] = "!"+item
    } 
    return EPSlice
}

// Remove item from a slice based on index
func RemoveByIdx(s *[]string, i int) {
    // handle index out of range, return same slice if index out of range
    length := len(*s)
    if i > length {
        return
    }
    (*s)[i] = (*s)[len(*s)-1]
    (*s)[len(*s)-1] = ""
    *s = (*s)[:len(*s)-1]
}

// Remove item from a slice based on val
func RemoveByVal(s []string, v string) {
    // handle index out of range, return same slice if index out of range
    for i, val := range s {
        if val == v {
            RemoveByIdx(&s, i)
        }
    }
}

// Convert offset to []int -> holds indices and []string -> holds attr value
func (o *offset) ToSlice() (idx []int, val []string) {
    for k, v := range *o {
        idx = append(idx, k)
        val = append(val, v)
    }
    return idx, val
}
