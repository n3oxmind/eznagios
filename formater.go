package main


import (
    "fmt"
    "sort"
    "strings"
)

// Convert *set (attr values) into string
func (s *set) joinAttrVal() string {
    attrVals := s.SortAttrVal()
    return strings.Join(attrVals, ",")
}

// Format object attribute
func formatAttr(od def) string {
    colGap := 2
    oDefFormat := ""
    maxAttrLen := 0
    // find max attr name
    for aName,_ := range od {
        if len(aName) > maxAttrLen {
           maxAttrLen = len(aName)
        }
    }
    // join attr values
    for aName,aVal := range od {
        attrValue := aVal.joinAttrVal()
        oDefFormat += fmt.Sprintf("\t%*v%v\n",-(maxAttrLen+colGap), aName,attrValue)
    }
    return oDefFormat
}

// Sort attribute values of type *set
func (a *set) SortAttrVal() []string{
    attrVals := make([]string, 0, a.Size())
    for attrVal := range a.m {
        attrVals = append(attrVals, attrVal.(string))

    }
    sort.Strings(attrVals)
    return attrVals
}
// Sort attribute names of type def
func (d def) sortAttrNames() []string {
    // attrNames are the keys for the map
    attrNames := make([]string, 0, len(d))
    for attrName := range d {
        attrNames = append(attrNames, attrName)
    }
    sort.Strings(attrNames)
    return attrNames
}

// Format Nagios object Definition before printing it
func formatObjDef (od def, maxAttrLen int) string {
    objDefFormat := ""
    attrNames := od.sortAttrNames()                                                         // sort map keys
    for _,attrName := range attrNames { 
        attrValue := od[attrName].joinAttrVal()                                             //sort and join attr value (*set)
        objDefFormat += fmt.Sprintf("\t%*v% v\n",-(maxAttrLen+4), attrName,attrValue)         //formated attr
    }
    return objDefFormat
}

// Print a pretty format of object definition
func (d defs)  printObjDef(h string) {
    maxAttrLength := 30
    objType := ""
    switch h {
    case "host":
        maxAttrLength = maxHostAttrLen
        objType       = "define host"
    case "service":
        maxAttrLength = maxSvcAttrLen
        objType       = "define service"
    case "servicegroup":
        maxAttrLength = maxSvcGrpAttrLen
        objType       = "define servicegroup"
    case "hostgroup":
        maxAttrLength = maxHgrpAttrLen
        objType       = "define hostgroup"
    case "contact":
        maxAttrLength = maxContactAttrLen
        objType       = "define contact"
    case "contactgroup":
        maxAttrLength = maxCGrpAttrLen
        objType       = "define contactgroup"
    case "hostdpendency":
        maxAttrLength = maxHostDpndAttrLen
        objType       = "define hostdependency"
    case "servicedpendency":
        maxAttrLength = maxSvcDpndlAttrLen
        objType       = "define servicedependency"
    case "serviceescalation":
        maxAttrLength = maxSvcEsclAttrLen
        objType       = "define serviceescalation"
    case "Hostescalation":
        maxAttrLength = maxHostEsclAttrLen
        objType       = "define hostescalation"
    default:
        //warning
        maxAttrLength = 30
    }
    for _,def := range d {
        oAttrs := formatObjDef(def, maxAttrLength)
        fmt.Println(objType+"{\n",oAttrs,"}")
    }
}

//Print object definitions in Go format
func (o defs) printObj(ftype string) {
    if ftype != "go" {
        for _,s := range o {
            for k, v := range s {
                fmt.Printf("%*v\t%v: %v\n",-10, k, v, v.Size())
            }
            fmt.Println()
        }
        //fmt.Printf("objNum: %v", len(o))
        fmt.Println()
    } else {
        fmt.Printf("%v\n", o)
        fmt.Println()
    }        
}

// Print *set (object attribute value)
func (a *set) printAttr() {
    for item := range a.m{
        fmt.Printf("%v,",item)
    }
    fmt.Println()

}
// get max length of items in slice
func MaxLen(a *[]string) int {
    maxLen := 10
    for _, v := range *a {
        if len(v) > maxLen {
            maxLen = len(v)
        }
    }
    return maxLen
}

func FillEmpty(s *[]string, hg *[]string, h *string) ([]string,[]string,[]string) {
    maxSize := len(*s)
    if len(*hg) > maxSize {
        maxSize = len(*hg)
    }
    sort.Strings(*s)
    sort.Strings(*hg)
    svc  := make([]string, maxSize)     // new svc
    copy(svc,*s)
    hgrp := make([]string, maxSize)    // new hgrp
    copy(hgrp,*hg)
    host := make([]string, maxSize)    // new host
    host[0] = *h

    return host,svc,hgrp

}

// Print host info (services and hostgroup association)
func printHostInfo(hostname string, hostgroups hostgroupOffset, services serviceOffset) {
    svc := services.GetEnabledServiceName()
    hgrp := hostgroups.GetEnabledHostgroupName()
    hostLen := len(hostname)
    svcMaxLen := MaxLen(&svc)
    hgrpMaxLen := MaxLen(&hgrp)
    // max index will be use for looping
    h, s, hg := FillEmpty(&svc,&hgrp,&hostname)

    line := strings.Repeat("-",hostLen+svcMaxLen+hgrpMaxLen+8)
    //header
    fmt.Println(line)
    tableHeader := fmt.Sprintf("| %-*v | %-*v| %-*v|\n%v",hostLen,"Hostname", svcMaxLen,"Service", hgrpMaxLen, "Hostgroup",line)
    fmt.Println(tableHeader)
    //rows
    for i,v := range h {
        row := fmt.Sprintf("| %-*v | %-*v| %-*v|",hostLen,v, svcMaxLen,s[i], hgrpMaxLen, hg[i])
        fmt.Println(row)
    }
    fmt.Println(line)
}
