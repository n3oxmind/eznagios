package main
import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "errors"
    "bytes"
)

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
                if _,exist := find(excludeDir,fileInfo.Name()); !exist{
                    configFiles = append(configFiles, path)
                }
            }
            return  nil
        }
        return nil
    })
    if exitErr != nil {
        panic(fmt.Sprintf("%v",exitErr))
    }
    // exit if nothing found
    if !(len(configFiles) > 0) {
        panic(fmt.Sprintf("no config files found in '%v'", path))
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
        fmt.Println(&duplicateAttributeError{err,objType,*attrName,rdef.rawParseObjAttr(),dupDef})
    }
}

// parse object attributes without removing/deleting anything (except empty attrVal)
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
        oAttr.Remove("")                                            // remove empty attr val silently
        objDef.FindDuplicateAttrName(&attr[1], mAttr, objType)      // check for duplicate attr name
        objDef[attr[1]] = oAttr                                     // add attr to the def
    }
    return objDef
}

// Get nagios objects definitions
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
                if objAttrs.attrExist("name"){
                    objDefs.SetHostTempDefs(objAttrs)
                } else {
                    objDefs.SetHostDefs(objAttrs)
                }
            case "defineservice{":
                if objAttrs.attrExist("name"){
                    objDefs.SetServiceTempDefs(objAttrs)
                } else {
                    objDefs.SetServiceDefs(objAttrs)
                }
            case "definehostgroup{":
                objDefs.SetHostGroupDefs(objAttrs)
            case "definehostdependency{":
                objDefs.SetHostDependencyDefs(objAttrs)
            case "defineservicedependency{":
                objDefs.SetServiceDependencyDefs(objAttrs)
            case "definecontact{":
                if objAttrs.attrExist("name"){
                    objDefs.SetContactTempDefs(objAttrs)
                } else {
                    objDefs.SetContactDefs(objAttrs)
                }
            case "definecontactgroup{":
                objDefs.SetContactGroupDefs(objAttrs)
            case "definecommand{":
                objDefs.SetcommandDefs(objAttrs)
                objDefs.commandDefs = append(objDefs.commandDefs, objAttrs)
            default:
                err := errors.New("unknown naigos object type")
                fmt.Println(&unknownObjectError{objAttrs,objType,err})
        }
    }
    } else {
        err := errors.New("no nagios object definition found")
        return  nil,&objectNotFoundError{err}
    }
    return &objDefs, nil
}

// Find hostgroup association (hostgroups that belong to a specific host)
func findHostGroups(hg *defs, td *defs, hOffset hostOffset) hostgroupOffset {
    hgrpOffset := newHostGroupOffset()
    hostnameExcl := "!"+hOffset.GetHostName()
    for idx,def := range *hg {
        if def.attrExist("members"){
            if def["members"].RegexHas(hOffset.GetHostName()) && !hgrpOffset.GetMembersOffset().OffsetExist(idx) {
                hostgroupName := def["hostgroup_name"].joinAttrVal()
                hgrpOffset.SetMembersOffset(idx, hostgroupName)
                findHostGroupMembership(hg, hostgroupName, *hgrpOffset)
            } else if def["members"].Has(hostnameExcl) && !hgrpOffset.GetMembersExclOffset().OffsetExist(idx) {
                hostgroupName := def["hostgroup_name"].joinAttrVal()
                hgrpOffset.SetMembersExclOffset(idx, hostgroupName)
            }
        }
    }
    // search host template for hostgroups
    if hOffset.templateHostgroups != nil {
        for _, tempName := range hOffset.templateHostgroups {
            findTemplateHostGroups(td, hg,tempName, *hgrpOffset)
        }
    }
    // save hostgroup name and location
    (*hgrpOffset).SetEnabledHostgroup()
    (*hgrpOffset).SetDisabledHostgroup()

    return *hgrpOffset
}

// recursive lookup for hostgroup delcared in the template definition (support inheritance)
func findTemplateHostGroups(t *defs, hg *defs, tempName string, hgrpOffset hostgroupOffset)  {
    for idx, def := range *t {
        if def["name"].Has(tempName) {
            if def.attrExist("hostgroups") {
                hgrpName := def["hostgroups"].joinAttrVal()
                hgrpOffset.SetTemplateHostgroupsOffset(idx,hgrpName)
                findHostGroupMembership(hg, hgrpName, hgrpOffset)
            }
            // nested template, 'use' can have multiple vals
            if def.attrExist("use") {
                for item,_ := range def["use"].m {
                    findTemplateHostGroups(t, hg,item.(string), hgrpOffset)
                }
            }
        break
        }
    }
}

// Perform recursive lookup for hostgroup membership (where a hostgroup could be a member of another hostgroup)
func findHostGroupMembership(d *defs, hgName string, hgrpOffset hostgroupOffset) {
    hostgroupNameExcl := fmt.Sprintf("!%v",hgName)
    for idx, def := range *d {
        if def.attrExist("hostgroup_members"){
            if def["hostgroup_members"].Has(hgName) && !hgrpOffset.GetHostgroupMembersOffset().OffsetExist(idx){
                hostgroupName := def["hostgroup_name"].joinAttrVal()
                hgrpOffset.SetHostgroupMembersOffset(idx,hostgroupName)
                findHostGroupMembership(d, hostgroupName, hgrpOffset)
                // I dont think you can exclude hostgroup in hostgroup object definition
                // this could be removed if the above is true 100%
            } else if def["hostgroup_members"].Has(hostgroupNameExcl) && !hgrpOffset.GetHostgroupMembersExclOffset().OffsetExist(idx){
                hostgroupName := def["hostgroup_name"].joinAttrVal()
                hgrpOffset.SetHostgroupMembersExclOffset(idx,hostgroupName)
            }
        }
    }
}

// check if []string has a specific item
func find(s []string, pattern string) (int, bool) {
    for i, item := range s {
        if item == pattern {
            return i, true
        }
    }
    return -1, false
}

// Find services association
func findServices(d *defs, t *defs, hostgroups hostgroupOffset, hostname string) serviceOffset {
    svcOffset := newServiceOffset()
    hostnameExcl := "!"+hostname
    hgEnabled := hostgroups.GetEnabledHostgroupName()
    hgNameExcl := AddEP(hgEnabled)
    // this could be enhanced by create separate function to perform search instead of append
    // since the service file size if not that big this wont make difference in performance
    serviceDefinitions := append(*d, *t...)
    for idx, def := range serviceDefinitions {
        // check if hostname exist in host_name attr
        if def.attrExist("host_name") {
            if def["host_name"].RegexHas(hostname) {
                svcOffset.SetHostNameOffset(idx, def["service_description"].joinAttrVal())
            } else if def["host_name"].Has(hostnameExcl) && !svcOffset.GetHostNameExclOffset().OffsetExist(idx){
                svcOffset.SetHostNameExclOffset(idx, def["service_description"].joinAttrVal())
            }
        }
        // check if any enabled/excluded hostgroup exist in hostgroup_name attr
        if def.attrExist("hostgroup_name") {
            if def["hostgroup_name"].HasAny(hgEnabled) {
                svcOffset.SetHostgroupNameOffset(idx,def["service_description"].joinAttrVal())
            }else if def["hostgroup_name"].HasAny(hgNameExcl){
                svcOffset.SetHostgroupNameExclOffset(idx, def["service_description"].joinAttrVal())
            }
        }
        // service definition attributes take precedence over service template attributes
        // check if service template contain hostgroups 
        if def.attrExist("use") && (!def.attrExist("hostgroup_name") || !def.attrExist("host_name"))  {
            findServiceTemplateHostgroupName(svcOffset, t, def["use"].StringSlice(), def, hostname, hgEnabled)
        }
    }
    // Filter enabled and excluded/disabled services
    (*svcOffset).SetEnabledService()
    (*svcOffset).SetDisabledService()
    fmt.Println(svcOffset.GetTemplateHostgroupNameOffset())
    fmt.Println(svcOffset.GetTemplateHostNameOffset())
    return *svcOffset
}

// recursive search the service template for possible hostgroup_name
func findServiceTemplateHostgroupName(svcOffset *serviceOffset, svcTempDefs *defs, tempNames []string, svcDef def, hostname string, hgNames []string) {
    for _, tempName := range tempNames {
        for i, def := range *svcTempDefs {
            if def.attrExist("name"){
                if def["name"].Has(tempName){// template enabled disabled __< here left
                    if def.attrExist("hostgroup_name") && def["hostgroup_name"].HasAny(hgNames){
                        svcOffset.SetTemplateHostgroupNameOffset(i, def["hostgroup_name"].joinAttrVal())
                    }
                    if def.attrExist("host_name") && def["host_name"].RegexHas(hostname) {
                        svcOffset.SetTemplateHostNameOffset(i, def["host_name"].joinAttrVal())
                    }
                    if def.attrExist("use") {
                        findServiceTemplateHostgroupName(svcOffset, svcTempDefs, def["use"].StringSlice(), def, hostname, hgNames)
                    }
                    // break if template does not have nested use attribute
                    break
                }
            }
        }
    }
}


// Find for hostname
func findHost(d *defs , hostname string) hostOffset {
    hOffset := newHostOffset()
    for idx, def := range *d {
        if def.attrExist("host_name") {
            if def["host_name"].Has(hostname) {
                hOffset.SetHostIndex(idx, hostname)
                hOffset.SetHostName(hostname)
                hOffset.SetHostOffset(idx)
                hOffset.SetHostDefinition(def)
                // TODO:need to find away to preserve order *set unordered need to use [] instead for 'use' only
                if def.attrExist("use"){
                    hOffset.SetTemplateHostgroups(def["use"])
                }
                if def.attrExist("hostgroups") {
                    hOffset.SetHostgroups(def["hostgroups"])
                }
                break
            }
        }
    }
    return *hOffset
}

func main() {
    path := "/home/afathi/nagios-configs"
//    path := "test/"
//    hostname := "sdk-jenkins.sea.bigfishgames.com"
    hostname := "java03-mongo-db05.sea.bigfishgames.com"
//    hostname := "host3.bigfishgames.com"
    excludedDir := []string {"automated", ".git", "libexec", "timeperiods.cfg", "servicegroups.cfg"}
    configFiles := findConfFiles(path, ".cfg", excludedDir)
    data, err := readConfFile(configFiles)
    if err != nil {
        panic(fmt.Sprintf("%v", err))
    }
    objDefs,err := getObjDefs(data); if err != nil {
        fmt.Println(err, path)
        os.Exit(1)
    }
    // search for the host
    host :=  findHost(&objDefs.hostDefs, hostname)
    if host.GetHostName() == "" {
        fmt.Println("Warning: host does not exist")
        os.Exit(1)
    }
    // find hostgroup association 
    hostgroups := findHostGroups(&objDefs.hostgroupDefs, &objDefs.hostTempDefs, host)
    // find service association
    services := findServices(&objDefs.serviceDefs, &objDefs.serviceTempDefs, hostgroups, host.GetHostName())
    if services.GetEnabledServiceName() == nil {
        services.svcEnabledName = append(services.svcEnabledName, "Not Found")
        fmt.Println(hostgroups)
    }
    printHostInfo(host.GetHostName(), hostgroups, services)
//    arr := strSlice{"a", "b", "c", "d"}
//    fmt.Println(arr)
//    arr.RemoveByVal("c")
//    fmt.Println(arr)
//    objDefs.hostTempDefs.printObjDef("host")
//    objDefs.hostgroupDefs.printObjDef("hostgroup")
//    objDefs.serviceTempDefs.printObjDef("service")
//    objDefs.contactTempDefs.printObjDef("service")
}
