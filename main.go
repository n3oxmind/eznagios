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
    // hostgroups from host obj definition(include host template)
    for _, hgrp := range hOffset.GetEnabledHostgroupsName(){
        findHostGroupMembership(hg, hgrp, *hgrpOffset)
    }
    // set enabled hostgroups
    (*hgrpOffset).SetEnabledHostgroup()
    (*hgrpOffset).SetDisabledHostgroup()
    // add hostgroups extracted from host obj definition to hostgroup list in the hostgroupOffset
    hOffset.SetEnabledHostgroups(hgrpOffset)

    return *hgrpOffset
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
    hgExcluded := AddEP(hgEnabled)
    // search template inheritance (recursively) for association
    findServiceTemplate(t, svcOffset,hostname, &hgEnabled, &hgExcluded)
    tmplEnabled := svcOffset.GetEnabledTemplateName()
    for idx, def := range *d {
        // check if service definition contain host_name attribute
        if  def.attrExist("host_name"){
            if def["host_name"].RegexHas(hostname) {
                svcOffset.SetHostNameOffset(idx, def["service_description"].joinAttrVal())
            }
            if def["host_name"].Has(hostnameExcl){
                svcOffset.SetHostNameExclOffset(idx, def["service_description"].joinAttrVal())
            }
        }
        // check if service definition contains hostgroup_name attribute
        if def.attrExist("hostgroup_name"){
            if def["hostgroup_name"].HasAny(hgEnabled) {
                svcOffset.SetHostgroupNameOffset(idx,def["service_description"].joinAttrVal())
            }
            if def["hostgroup_name"].HasAny(hgExcluded){
                svcOffset.SetHostgroupNameExclOffset(idx, def["service_description"].joinAttrVal())
            }
        }
        // service definition that does not have hostname/hostgroup_name attr just 'use'
        if def.attrExist("use"){
            if def["use"].HasAny(tmplEnabled){
                svcOffset.SetUseOffset(idx, def["service_description"].joinAttrVal())
            }
        }
    }
    // Filter enabled and excluded/disabled services
    (*svcOffset).SetEnabledService()
    (*svcOffset).SetDisabledService()
    return *svcOffset
}

// find service template association
func findServiceTemplate(t *defs, svcOffset *serviceOffset, hostname string,  hgEnabled *[]string , hgExcluded *[]string) {
    hasAssociation := false
    for idx, def := range *t {
        if def.attrExist("host_name") {
            if def["host_name"].RegexHas(hostname){
                if def.attrExist("service_description"){
                    svcOffset.SetHostNameOffset(idx, def["service_description"].joinAttrVal())
                }else {
                    svcOffset.SetTemplateHostNameOffset(idx,def["name"].joinAttrVal())
                }
                hasAssociation = true
            }
            if def["host_name"].RegexHas("!"+hostname){
                if def.attrExist("service_description"){
                    svcOffset.SetHostNameExclOffset(idx, def["service_description"].joinAttrVal())
                }else{
                    svcOffset.SetTemplateHostNameExclOffset(idx,def["name"].joinAttrVal())
                }
            }
        }
        if def.attrExist("hostgroup_name"){
            if def["hostgroup_name"].HasAny(*hgEnabled){
                if def.attrExist("service_description"){
                    svcOffset.SetHostgroupNameOffset(idx, def["service_description"].joinAttrVal())
                }else{
                    svcOffset.SetTemplateHostgroupNameOffset(idx,def["name"].joinAttrVal())
                }
                hasAssociation = true
            }
            if def["hostgroup_name"].HasAny(*hgExcluded){
                if def.attrExist("service_description"){
                    svcOffset.SetHostgroupNameExclOffset(idx, def["service_description"].joinAttrVal())
                }else{
                    svcOffset.SetTemplateHostgroupNameExclOffset(idx,def["name"].joinAttrVal())
                }
            }
        }
        if hasAssociation && def.attrExist("use") {
            GetInheritanceDepth(t , svcOffset , def["use"].joinAttrVal(), hostname , hgEnabled, hgExcluded,idx, def["name"].joinAttrVal())
        }
    }
// remove duplicate and return enabled template only
svcOffset.SetEnabledTemplate()
}

func GetInheritanceDepth(t *defs, svcOffset *serviceOffset, tmplName string, hostname string,  hgEnabled *[]string , hgExcluded *[]string, idx int, name string) {
    // speed up lookup for the same inheritance chain
    for _, def := range *t {
        if tmplName == def["name"].joinAttrVal() {
            if def.attrExist("host_name") {
                if def["host_name"].RegexHas(hostname){
                    if def.attrExist("service_description"){
                        svcOffset.SetHostNameOffset(idx, def["service_description"].joinAttrVal())
                    }else {
                        svcOffset.SetTemplateHostNameOffset(idx,def["name"].joinAttrVal())
                    }
                }
                if def["host_name"].RegexHas("!"+hostname){
                    if def.attrExist("service_description"){
                        svcOffset.SetHostNameExclOffset(idx, def["service_description"].joinAttrVal())
                    }else{
                        svcOffset.SetTemplateHostNameExclOffset(idx,def["name"].joinAttrVal())
                    }
                }
            }
            if def.attrExist("hostgroup_name"){
                if def["hostgroup_name"].HasAny(*hgEnabled){
                    if def.attrExist("service_description"){
                        svcOffset.SetHostgroupNameOffset(idx, def["service_description"].joinAttrVal())
                    }else{
                        svcOffset.SetTemplateHostgroupNameOffset(idx,def["name"].joinAttrVal())
                    }
                }
                if def["hostgroup_name"].HasAny(*hgExcluded){
                    if def.attrExist("service_description"){
                        svcOffset.SetHostgroupNameExclOffset(idx, def["service_description"].joinAttrVal())
                    }else{
                        svcOffset.SetTemplateHostgroupNameExclOffset(idx,def["name"].joinAttrVal())
                    }
                }
            }
            if def.attrExist("use") {
                GetInheritanceDepth(t , svcOffset , def["use"].joinAttrVal(), hostname , hgEnabled, hgExcluded, idx, name)
            }
        break
        }
    }
}


// Find for hostname
func findHost(d *defs ,t *defs, hostname string) hostOffset {
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
                    for tmpl := range def["use"].m{
                        findHostTemplate(t, hOffset, tmpl.(string))
                    }
                }
                if def.attrExist("hostgroups") {
                    hOffset.SetHostgroupsOffset(idx, def["hostgroups"])
                }
                break
            }
        }
    }
    hOffset.SetEnabledHostgroupsName()
    return *hOffset
}

// recursive lookup for hostgroup delcared in the template definition (support inheritance)
func findHostTemplate(t *defs, hOffset *hostOffset, tmplName string ){
    for idx, def := range *t {
        if def["name"].Has(tmplName) {
            if def.attrExist("hostgroups") {
                hOffset.SetTemplateHostgroupsOffset(idx, def["hostgroups"])
            }
            if def.attrExist("use"){
                for tmpl := range def["use"].m{
                    findHostTemplate(t, hOffset, tmpl.(string))
                }
            }
            break
        }
    }
}

func main() {
//    path := "/home/afathi/nagios-configs"
    path := "test/"
//    hostname := "sdk-jenkins.sea.bigfishgames.com"
//    hostname := "java03-mongo-db05.sea.bigfishgames.com"
//    hostname := "sdk-jenkins.sea.bigfishgames.com"
    hostname := "host3.bigfishgames.com"
//    hostname := "casino-game210.sea.bigfishgames.com"
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
    host :=  findHost(&objDefs.hostDefs, &objDefs.hostTempDefs, hostname)
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
