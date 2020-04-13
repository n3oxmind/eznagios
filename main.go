package main
import (
    "bytes"
    "errors"
    _ "eznagios/math"
    . "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "regexp"
    "strings"
)
attrVal := _.NewSet()
type rawDef [][]string
type def map[string]*Set
type defs []def
type obj struct {
    hostDefs                defs
    serviceDefs             defs
    hostgroupDefs           defs
    hostdependencyDefs      defs
    servicedependencyDefs   defs
    contactDefs             defs
    contactgroupDefs        defs
    commandDefs             defs
}
// Find Nagios config files
func findConfFiles(path string, fileExtention string, excludeDir []string) (configFiles []string) {
    exitErr := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
        fileInfo, err := os.Stat(path)
        if err != nil {
            return  err
        }
        //exclude directories
        if info.IsDir() && len(excludeDir) > 0  {
            for _, dirname := range excludeDir {
                if dirname == info.Name() {
                    return filepath.SkipDir
                }
            }
        }
        //filter configFiles based on file extention
        mode := fileInfo.Mode()
        if mode.IsRegular() {
            if filepath.Ext(path) == fileExtention {
                configFiles = append(configFiles, path)
            }
            return  nil
        }
        return nil
    })
    if exitErr != nil {
        panic(Sprintf("%v",exitErr))
    }
    // exit if nothing found
    if !(len(configFiles) > 0) {
        panic(Sprintf("no config files found in '%v'", path))
    }
    return configFiles
}

// Read the contents of Nagios config files
func readConfFile(filename []string ) (data string, err error) {
    var buffer bytes.Buffer
    for _, cfile := range filename {
        data, err := ioutil.ReadFile(cfile); if err != nil {
            return "", err
        }
        buffer.Write(data)
    }
    return buffer.String(), nil
}


// find duplicate attribute name
func (d def) FindDuplicateAttrName(attrName *string, rdef rawDef, objType string) {
    if _, exist :=  d[*attrName]; exist {
        dupDef := def{}
        dupDef[*attrName] = d[*attrName]
        err := errors.New("duplicate attribute found")
        Println(&duplicateAttributeError{err,objType,*attrName,rdef.rawParseObjAttr(),dupDef})
    }
}
// parse object attributes without removing/deleting anything
func (a rawDef) rawParseObjAttr()  def {
    objDef := def{}
    for _,attr := range a {
        oAttr := NewSet()
        oAttrVal := strings.Split(attr[2], ",")
        for _,val := range oAttrVal {
            oAttr.Add(strings.TrimSpace(val))
        }
        oAttr.Remove("")                                    // remove empty attr val silently
        objDef[attr[1]] = oAttr                             // add attr to the def
    }
    return objDef
}

// parse Nagios object attributes
// param: attr[1] -- attrName
// param: attr[2] -- attrVal
func parseObjAttr( rawObjDef []string, reAttr *regexp.Regexp, objType string )  def {
    objDef := def{}
    mAttr := reAttr.FindAllStringSubmatch(rawObjDef[2], -1)
    for _,attr := range mAttr {
        oAttr := NewSet()
        oAttrVal := strings.Split(attr[2], ",")
        for _,val := range oAttrVal {
            oAttr.Add(strings.TrimSpace(val))
        }
        oAttr.Remove("")                                    // remove empty attr val silently
        objDef.FindDuplicateAttrName(&attr[1], mAttr, objType)       // check for duplicate attr name
        objDef[attr[1]] = oAttr                             // add attr to the def
        //objDef[attr[1]].SortAttrVal()
    }
    return objDef
}

// Get Nagios objects definitions
func getObjDefs(data string) (*obj, error) {
    objDefs := obj{}
    reAttr := regexp.MustCompile(`\s*(?P<attr>.*?)\s+(?P<value>.*)\n`)
    reObjDef := regexp.MustCompile(`(?sm)(^\s*define\s+[a-z]+?\s*{)(.*?)(})`)
    rawObjDefs := reObjDef.FindAllStringSubmatch(data, -1)
    if rawObjDefs != nil {
        for _,oDef:= range rawObjDefs {
            defStart := strings.Join(strings.Fields(oDef[1]),"")
            objType := strings.TrimSpace(oDef[1])
            objAttrs := parseObjAttr(oDef, reAttr, objType)
            switch defStart {
            case "definehost{":
                objDefs.hostDefs = append(objDefs.hostDefs, objAttrs)
            case "defineservice{":
                objDefs.serviceDefs = append(objDefs.serviceDefs, objAttrs)
            case "definehostgroup{":
                objDefs.hostgroupDefs = append(objDefs.hostgroupDefs, objAttrs)
            case "definehostdependency{":
                objDefs.hostdependencyDefs = append(objDefs.hostdependencyDefs, objAttrs)
            case "defineservicedependency{":
                objDefs.servicedependencyDefs = append(objDefs.servicedependencyDefs, objAttrs)
            case "definecontact{":
                objDefs.contactDefs = append(objDefs.contactDefs, objAttrs)
            case "definecontactgroup{":
                objDefs.contactgroupDefs = append(objDefs.contactgroupDefs, objAttrs)
            case "definecommand{":
                objDefs.commandDefs = append(objDefs.commandDefs, objAttrs)
            default:
                err := errors.New("unknown naigos object type")
                Println(&unknownObjectError{objAttrs,objType,err})
        }
    }
    } else {
        err := errors.New("no nagios object definition found")
        return  nil,&objectNotFoundError{err}
    }
    return &objDefs, nil
}

// check if attr (type string) exist in map[string]*set/def
func (d def) attrExist(attrName string) bool {
    if _, exist := d[attrName]; exist {
        return true
    }
    return false
}


type offset map[int]string
type objectOffset struct {
    index     offset
    indexExcl offset
}

// objectOffset constructor
func newObjectOffset() *objectOffset {
    o := &objectOffset{}
    o.index = make(offset)
    o.indexExcl = make(offset)
    return o
}

// objectOffset getters and setters
func (o *objectOffset) offset() offset {
    return o.index
}
func (o *objectOffset) offsetExcl() offset {
    return o.indexExcl
}

func (o *objectOffset) setOffset(i int,s string) {
    o.index[i] = s
}

func (o *objectOffset) setOffsetExcl(offsetExcl offset) {
    o.indexExcl = offsetExcl
}

// check if index (type int) already exist
func (o offset) offsetExist(i int) bool {
    if _, exist := o[i]; exist {
        return true
    }
    return false
}

// Find hostgroup association
// hostgroup_name, members, hostgroup_members
func findHostGroups(d *defs, hostname string) objectOffset {
    hgrpOffset := newObjectOffset()
    hostnameExcluded := Sprintf("!%v",hostname)
    for idx,def := range *d {
        if def.attrExist("members"){
            if def["members"].Has(hostname) && !hgrpOffset.offset().offsetExist(idx) {
                hostgroupName := def["hostgroup_name"].joinAttrVal()
                hgrpOffset.setOffset(idx, hostgroupName)
                findHostGroupMembership(d, hostgroupName, hgrpOffset.offset(), hgrpOffset.offsetExcl())
            } else if def["members"].Has(hostnameExcluded) && !hgrpOffset.offset().offsetExist(idx) {
                hostgroupName := def["hostgroup_name"].joinAttrVal()
                hgrpOffset.setOffset(idx, hostgroupName)
            }
        }
    }
    return *hgrpOffset
}

func findHostGroupMembership(d *defs, hostgroupName string, hostgroupOffset,hostgroupOffsetExcluded offset) {
    hostgroupNameExcluded := Sprintf("!%v",hostgroupName)
    for idx, def := range *d {
        if def.attrExist("hostgroup_members"){
            if def["hostgroup_members"].Has(hostgroupName) && !hostgroupOffset.offsetExist(idx){
                hostgroupName2 := def["hostgroup_name"].joinAttrVal()
                hostgroupOffset[idx] = hostgroupName2
                findHostGroupMembership(d, hostgroupName2, hostgroupOffset, hostgroupOffsetExcluded)
            } else if def["hostgroup_members"].Has(hostgroupNameExcluded) && !hostgroupOffset.offsetExist(idx){
                hostgroupName2 := def["hostgroup_name"].joinAttrVal()
                hostgroupOffsetExcluded[idx] = hostgroupName2
            }
                
        }
    }
}

func findServices(d *defs, hostgroups objectOffset, hostname string) offset {
    serviceOffset := offset{}
    for idx, def := range *d {
        if def.attrExist("host_name") {
            Println(idx,"here")
        }
    }
    return serviceOffset
}
func main() {
    //path := "/home/afathi/nagios-configs"
    path := "test/"
    hostname := "host3.bigfishgames.com"
    excludedDir := []string {"automated", ".git", "libexec"}
    configFiles := findConfFiles(path, ".cfg", excludedDir)
    data, err := readConfFile(configFiles)
    if err != nil {
        panic(Sprintf("%v", err))
    }
    objDefs,err := getObjDefs(data); if err != nil {
        Println(err, path)
        os.Exit(1)
    }
    hostgroups := findHostGroups(&objDefs.hostgroupDefs, hostname)
    services := findServices(&objDefs.serviceDefs, hostgroups, hostname)
    Println(hostgroups)
    Println(services)
    //objDefs.hostDefs.printObjDef("host")
    //objDefs.hostgroupDefs.printObjDef("hostgroup")
    //objDefs.serviceDefs.printObjDef("service")
}
