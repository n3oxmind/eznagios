package main


import (
    "fmt"
    "sort"
    "strings"
)

// Convert *set (attr values) into string
func (s attrVal) joinAttrVal() string {
    (s).SortAttrVal()
    return strings.Join(s, ",")
}

// Format object attribute
func formatAttr(od def) string {
    colGap := 2
    oDefFormat := ""
    maxAttrLen := 0
    // find max attr name
    for aName := range od {
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
func (a *attrVal) SortAttrVal(){
    sort.Strings(*a)
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
func formatObjDef (od def, objType string, maxAttrLen int) string {
    objDefFormat := ""
    attrNames := od.sortAttrNames()                                                         // sort map keys
    for _,attrName := range attrNames { 
        attrValue := od[attrName].joinAttrVal()
        objDefFormat += fmt.Sprintf("\t%*v% v\n",-(maxAttrLen+4), attrName,attrValue)         //formated attr
    }
    return objType+"{\n"+objDefFormat+"}\n"
}

// Get maximum attribute of an obj
func getMaxAttr(s string) (string, int) {
    // default value for attr length
    maxAttrLength := 30
    objType := ""
    switch s {
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
    return objType, maxAttrLength
}

// Print a pretty format of object definition
func (d defs)  printObjDef(h string) {
    objType, maxAttr := getMaxAttr(h)
    for _,def := range d {
        formatDef := formatObjDef(def, objType, maxAttr)
        fmt.Printf(formatDef)
    }
}

//Print object definitions in Go format
func (o defs) printObj(ftype string) {
    if ftype != "go" {
        for _,s := range o {
            for k, v := range s {
                fmt.Printf("%*v\t%v: %v\n",-10, k, v, v.length())
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
func (a *attrVal) printAttr() {
    for _,item := range *a{
        fmt.Printf("%v,",item)
    }
    fmt.Println()

}
// get max length of items in slice
func MaxLen(a *attrVal) int {
    maxLen := 10
    for _, v := range *a {
        if len(v) > maxLen {
            maxLen = len(v)
        }
    }
    return maxLen
}

func FillEmpty(s *attrVal, hg *attrVal, h *string) ([]string,[]string,[]string) {
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
    svc := append(services.svcEnabledName, services.svcOthers...)
    fmt.Println(services.svcOthers)
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

// print a specifi object definition
func (d *defs) printDef(objType string, ids ...string) {
    otype, alen := getMaxAttr(objType)
    for _, id := range ids {
        formatDef := formatObjDef((*d)[id], otype, alen)
        fmt.Println(formatDef)
    } 
}
