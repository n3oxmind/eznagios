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

// ResizeSlice will resize and return an ordered slice
func (a *attrVal) ResizeSlice(s int) {
    emptySlice := make([]string, s-len(*a))
    sort.Strings(*a)
    *a = append(*a, emptySlice...)
}

// Print host info (services and hostgroups association)
func printHostInfoPretty(dictList []objDict, termWidth int) {
    // max lenght of every object attributes
    hostAttrMaxLen, svcAttrMaxLeng, hgrpAttrMaxLeng := 0, 0, 0
    // create a new slice if the capacity if more than 20
    svcs := make([]attrVal, len(dictList))
    hgrps := make([]attrVal, len(dictList))
    hosts := make([]attrVal, len(dictList))
    for i, dict := range dictList {
        svc  := append(dict.services.enabled.ToSlice(), dict.services.others...)
        hgrp := dict.hostgroups.enabled

        // max length of an object attribute
        hostAttrLen := len(dict.hosts.GetHostName())
        svcAttrLen := MaxLen(&svc)
        hgrpAttrLen := MaxLen(&hgrp)
        if hostAttrLen > hostAttrMaxLen {
            hostAttrMaxLen = hostAttrLen
        }
        if svcAttrLen > svcAttrMaxLeng {
            svcAttrMaxLeng = svcAttrLen
        }
        if hgrpAttrLen > hgrpAttrMaxLeng {
            hgrpAttrMaxLeng = hgrpAttrLen
        }
        // resize and sort
        hgrpSize := len(hgrp)
        svcSize := len(svc)
        if svcSize > hgrpSize {
            hgrp.ResizeSlice(svcSize)
            sort.Strings(svc)
        }else {
            svc.ResizeSlice(hgrpSize)
            sort.Strings(hgrp)
        }
        svcs[i] =  svc
        hgrps[i] = hgrp
        // hosts a slice to hold hostname, hostaddress, and some stats about obj association
        hosts[i] = []string{dict.hosts.hostName, dict.hosts.hostAddr, fmt.Sprintf("num of svcs: %v",svcSize), fmt.Sprintf("num of hgrps: %v", hgrpSize)}
    }
    //header
    line := strings.Repeat("-",hostAttrMaxLen+svcAttrMaxLeng+hgrpAttrMaxLeng+8)
    fmt.Println(line)
    tableHeader := fmt.Sprintf("| %-*v | %-*v| %-*v|\n%v",hostAttrMaxLen,"Hostname", svcAttrMaxLeng,"Service", hgrpAttrMaxLeng, "Hostgroup",line)
    fmt.Println(tableHeader)

    for i:=0; i < len(dictList); i++{
        numRows := len(svcs[i])
        //rows
        for j:=0; j < numRows; j++ {
            hostname := ""
            // hostname
            if j <= 1 {
                hostname = hosts[i][j]
            }else if j < len(hosts[i]) {
                hostname = hosts[i][j]
            }
            row := fmt.Sprintf("| %-*v | %-*v| %-*v|",hostAttrMaxLen,hostname, svcAttrMaxLeng,svcs[i][j], hgrpAttrMaxLeng, hgrps[i][j])
            fmt.Println(row)
        }
        fmt.Println(line)
    }
}

// print a specifi object definition
func (d *defs) printDef(objType string, ids ...string) {
    otype, alen := getMaxAttr(objType)
    for _, id := range ids {
        formatDef := formatObjDef((*d)[id], otype, alen)
        fmt.Println(formatDef)
    } 
}
// print host info not pretty but live ( show host as you find it )
func printHostInfo(hostname string, hostAddr string, hostgroups hostgroupOffset, services serviceOffset) {
    svc := append(services.enabled.ToSlice(), services.others...)
    hgrp := hostgroups.enabled
    svcSize := len(svc)
    hgrpSize := len(hgrp)
    svcAttrMaxLeng := MaxLen(&svc)

    // make svc and hgrp indices equal. this will make the print operation simple and more effecient
    maxSize := svcSize
    if svcSize > hgrpSize {
        for i:=0 ; i < svcSize-hgrpSize; i++ {
            hgrp = append(hgrp,"")
        }
    }else if svcSize < hgrpSize {
        for i:=0 ; i < hgrpSize-svcSize; i++ {
            svc = append(svc,"")
        }
        maxSize = hgrpSize
    }
    // print one host with its association at a time
    fmt.Printf("%v%v (%v)%v\n", Green, hostname, hostAddr, RST)
    for i:=0 ; i < maxSize; i++ {
        fmt.Printf("\t%-*v\t%v\n", svcAttrMaxLeng,svc[i], hgrp[i])
    }
}
