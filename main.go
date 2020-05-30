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

// Find duplicate attribute names
func (d def) FindDuplicateAttrName(attrName *string, rdef rawDef, objType string) {
    if _, exist :=  d[*attrName]; exist {
        dupDef := def{}
        dupDef[*attrName] = d[*attrName]
        err := errors.New("duplicate attribute found")
        fmt.Println(&duplicateAttributeError{err,objType,*attrName,rdef.rawParseObjAttr(),dupDef})
    }
}

// Parse object attributes without modifying the original data (except empty attrVal)
func (a rawDef) rawParseObjAttr()  def {
    objDef := def{}
    for _,attr := range a {
        oAttr := attrVal{}
        oAttrVal := strings.Split(attr[2], ",")
        for _,val := range oAttrVal {
            oAttr.Add(strings.TrimSpace(val))
        }
        oAttr.Remove("")                                    // remove empty attr val silently
        objDef[attr[1]] = &oAttr                            // add attr to the def
    }
    return objDef
}

// parse Nagios object attributes; attr[1]-> attrName, attr[2]->attrVal
func parseObjAttr( rawObjDef []string, reAttr *regexp.Regexp, objType string )  def {
    objDef := def{}
    mAttr := reAttr.FindAllStringSubmatch(rawObjDef[2], -1)
    for _,attr := range mAttr {
        oAttr := attrVal{}
        oAttrVal := strings.Split(attr[2], ",")
        for _,val := range oAttrVal {
            oAttr.Add(strings.TrimSpace(val))
        }
        oAttr.Remove("")                                            // remove empty attr val silently
        objDef.FindDuplicateAttrName(&attr[1], mAttr, objType)      // check for duplicate attr name
        objDef[attr[1]] = &oAttr                                     // add attr to the def
    }
    return objDef
}

// Get Nagios objects definitions
func getObjDefs(data string) (*obj, error) {
    objDefs := newObj()
    reAttr := regexp.MustCompile(`\s*(?P<attr>.*?)\s+(?P<value>.*)\n`)
    reObjDef := regexp.MustCompile(`(?sm)(^\s*define\s+[a-z]+?\s*{)(.*?\n)(\s*})`)
    rawObjDefs := reObjDef.FindAllStringSubmatch(data, -1)
    c1,c2 := 0, 0           // hostdependency and servicedependency does not have a unique identifier, will use index instead
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
                c1 += 1
                objDefs.SetHostDependencyDefs(objAttrs, c1)
            case "defineservicedependency{":
                c2 += 1
                objDefs.SetServiceDependencyDefs(objAttrs, c2)
            case "definecontact{":
                if objAttrs.attrExist("name"){
                    objDefs.SetContactTempDefs(objAttrs)
                } else {
                    objDefs.SetContactDefs(objAttrs)
                }
            case "definecontactgroup{":
                objDefs.SetContactGroupDefs(objAttrs)
            case "definecommand{":
                if objAttrs.attrExist("command_name") && objAttrs.attrExist("command_line"){
                    objDefs.SetcommandDefs(objAttrs)
                }else {
                    fmt.Println("here",objAttrs)
                }
            default:
                err := errors.New("unknown naigos object type")
                fmt.Println(&unknownObjectError{objAttrs,objType,err})
        }
    }
    } else {
        err := errors.New("no nagios object definition found")
        return  nil,&NotFoundError{err, "Fatal", ""} 
    }
    return objDefs, nil
}

// Find hostgroup association (hostgroups that belong to a specific host)
func findHostGroups(hg *defs, td *defs, hOffset hostOffset) hostgroupOffset {
    hgrpOffset := newHostGroupOffset()
    hostnameExcl := "!"+hOffset.GetHostName()
    for idx,def := range *hg {
        if def.attrExist("members"){
            if def["members"].RegexHas(hOffset.GetHostName()) && !hgrpOffset.members.Has(idx) {
                hgrpOffset.members.Add(idx)
                findHostGroupMembership(hg, idx, *hgrpOffset)
            } else if def["members"].Has(hostnameExcl) && !hgrpOffset.membersExcl.Has(idx) {
                hgrpOffset.membersExcl.Add(idx)
            }
        }
    }
    // hostgroups from host obj definition(include host template)
    for _, hgrp := range hOffset.GetEnabledHostgroupsName(){
        hgrp := strings.TrimLeft(hgrp,"+")
        findHostGroupMembership(hg, hgrp, *hgrpOffset)
    }
    // set enabled hostgroups
    (*hgrpOffset).SetEnabledHostgroup()
    (*hgrpOffset).SetDisabledHostgroup()
    // add hostgroups extracted from host obj definition to hostgroups list in the hostgroupOffset
    hOffset.SetEnabledHostgroups(hgrpOffset)
    return *hgrpOffset
}


// Perform recursive lookup for hostgroup membership (where a hostgroup could be a member of another hostgroup)
func findHostGroupMembership(d *defs, hgName string, hgrpOffset hostgroupOffset) {
    hostgroupNameExcl := fmt.Sprintf("!%v",hgName)
    hgrpExist := false
    for idx, def := range *d {
        if def["hostgroup_name"].ToString() == hgName {
            hgrpExist = true
        }
        if def.attrExist("hostgroup_members"){
            if def["hostgroup_members"].Has(hgName) && !hgrpOffset.hostgroupMembers.Has(idx){
                hgrpOffset.hostgroupMembers.Add(idx)
                findHostGroupMembership(d, idx, hgrpOffset)
                // I dont think you can exclude hostgroup in hostgroup object definition
                // this could be removed if the above is true 100%
            } else if def["hostgroup_members"].Has(hostgroupNameExcl) && !hgrpOffset.hostgroupMembers.Has(idx){
                hgrpOffset.hostgroupMembersExcl.Add(idx)
            }
        }
    }
    if !hgrpExist {
        fmt.Println("warning, hostgroup does not exist", hgName)
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
    hgEnabled := hostgroups.enabled
    hgExcluded := AddEP(hgEnabled)
    // search template inheritance (recursively) for association
    findServiceTemplate(t, svcOffset,hostname, &hgEnabled, hgExcluded)
    tmplEnabled := svcOffset.tmpl.enabled
    for idx, def := range *d {
        hasAssociation := false
        // check if service definition contain host_name attribute
        if  def.attrExist("host_name"){
            if def["host_name"].RegexHas(hostname) {
                svcOffset.hostName.Add(idx)
                hasAssociation = true
            }
            if def["host_name"].Has(hostnameExcl){
                svcOffset.hostNameExcl.Add(idx)
                hasAssociation = true
            }
        }
        // check if service definition contains hostgroup_name attribute
        if def.attrExist("hostgroup_name"){
            if def["hostgroup_name"].HasAny(hgEnabled...) {
                svcOffset.hostgroupName.Add(idx)
                hasAssociation = true
            }
            if def["hostgroup_name"].HasAny(*hgExcluded...){
                svcOffset.hostgroupNameExcl.Add(idx)
                hasAssociation = true
            }
        }
        // service definition that does not have hostname/hostgroup_name attr just 'use'
        if def.attrExist("use") && !hasAssociation {
            if def["use"].HasAny(tmplEnabled...){
                svcOffset.use.Add(idx)
            }
        }
    }
    // Filter enabled and excluded/disabled services
    (*svcOffset).SetEnabledService()
    (*svcOffset).SetDisabledService()
    return *svcOffset
}

// find service template association
func findServiceTemplate(t *defs, svcOffset *serviceOffset, hostname string,  hgEnabled *attrVal , hgExcluded *attrVal) {
    vistedTemplate := attrVal{}
    hasAssociation := false
    for idx, def := range *t {
        hasAssociation = false
        if def.attrExist("host_name") {
            if def["host_name"].RegexHas(hostname){
                if def.attrExist("name"){
                    svcOffset.Add("tmplHostName", idx, idx )
                    if def.attrExist("service_description") {
                        svcOffset.others.Add(def["service_description"].ToString())
                    }
                }else {
                    svcOffset.hostName.Add(def["service_description"].ToString())
                }
                hasAssociation = true
            }
            if def["host_name"].RegexHas("!"+hostname){
                if def.attrExist("name"){
                    svcOffset.Add("tmplHostNameExcl", idx, idx )
                }else{
                    svcOffset.hostNameExcl.Add(def["service_description"].ToString())
                }
            }
        }
        if def.attrExist("hostgroup_name"){
            if def["hostgroup_name"].HasAny(*hgEnabled...){
                if def.attrExist("name"){
                    svcOffset.Add("tmplHostgroupName", idx, idx )
                    if def.attrExist("service_description") {
                        svcOffset.others.Add(def["service_description"].ToString())
                    }
                }else{
                    svcOffset.hostgroupName.Add(def["service_description"].ToString())
                }
                hasAssociation = true
            }
            if def["hostgroup_name"].HasAny(*hgExcluded...){
                if def.attrExist("name"){
                    svcOffset.hostgroupName.Add(idx,def["name"].joinAttrVal())
                }else{
                    svcOffset.hostgroupNameExcl.Add(def["service_description"].ToString())
                }
            }
        }
        if hasAssociation && def.attrExist("use") {
            findServiceInheritance(t , svcOffset , *def["use"], hostname , hgEnabled, hgExcluded,idx, def["name"].ToString(), &vistedTemplate)
        }
    }
//    create log/debug level for this
//    fmt.Println(vistedTemplate)
    // remove duplicate and return enabled template only
    svcOffset.SetEnabledServiceTemplate()
    svcOffset.SetDisabledServiceTemplate()
}

// find inherited service template [template_name][temp1 temp2 temp3..]
func findServiceInheritance(t *defs, svcOffset *serviceOffset, useAttr attrVal, hostname string,  hgEnabled *attrVal , hgExcluded *attrVal, ID string, name string, vistedTemplate *attrVal) {
    // speed up lookup for the same inheritance chain
    for _, tmpl := range useAttr {
    // check if the template already been lookup for inheritance
        if !vistedTemplate.Has(tmpl){
            for _, def := range *t {
                if tmpl == def["name"].ToString() {
                    *vistedTemplate = append(*vistedTemplate, tmpl)
                    if def.attrExist("host_name") {
                        if def["host_name"].RegexHas(hostname){
                            if def.attrExist("name"){
                                svcOffset.Add("tmplhostName", ID, def["name"].ToString())
                                if def.attrExist("service_description") {
                                    svcOffset.others.Add(def["service_description"].ToString())
                                }
                            }else {
                                svcOffset.Add("tmplhostName", ID, def["service_description"].ToString())
                            }
                        }
                        if def["host_name"].RegexHas("!"+hostname){
                            if def.attrExist("name"){
                                svcOffset.Add("tmplhostNameExcl", ID, def["name"].ToString())
                            }else{
                                svcOffset.Add("tmplhostNameExcl", ID, def["service_description"].ToString())
                            }
                        }
                    }
                    if def.attrExist("hostgroup_name"){
                        if def["hostgroup_name"].HasAny(*hgEnabled...){
                            if def.attrExist("name"){
                                svcOffset.Add("tmplhostgroupName", ID, def["name"].ToString())
                                if def.attrExist("service_description") {
                                    svcOffset.others.Add(def["service_description"].ToString())
                                }
                            }else{
                                svcOffset.Add("tmplhostgroupName", ID, def["service_description"].ToString())
                            }
                        }
                        if def["hostgroup_name"].HasAny(*hgExcluded...){
                            if def.attrExist("name"){
                                svcOffset.Add("tmplhostgroupNameExcl", ID, def["name"].ToString())
                            }else{
                                svcOffset.Add("tmplhostgroupNameExcl", ID, def["service_description"].ToString())
                            }
                        }
                    }
                    if def.attrExist("use") {
                        findServiceInheritance(t , svcOffset , *def["use"], hostname , hgEnabled, hgExcluded, ID, name, vistedTemplate)
                    }
                break
                }
            }
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
                if def.attrExist("use"){
                    for _,tmpl := range *def["use"]{
                        findHostTemplate(t, hOffset, tmpl)
                    }
                }
                if def.attrExist("hostgroups") {
                    hOffset.SetHostgroupsOffset(idx, def["hostgroups"])
                }
                if def.attrExist("address") {
                    hOffset.hostAddr = def["address"].ToString()
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
                hOffset.SetTemplateOrder(idx)
            }
            if def.attrExist("use"){
                for _,tmpl := range *def["use"]{
                    findHostTemplate(t, hOffset, tmpl)
                }
            }
            break
        }
    }
}

// delete host obj
func deleteHost(hd *defs, td *defs, h *hostOffset){
    if len(*(*hd)[h.hostIndex]["host_name"]) > 1 {
        (*hd)[h.hostIndex]["host_name"].deleteAttrVal(hd, td, h.hostIndex, "HOST HOST_NAME", "host_name", h.hostName, h.hostName)
    } else {
        printDeletion(h.hostIndex, "HOST", "", "", "def")
        delete(*hd, h.GetHostName())
    }
    // TODO: checkif host template is being used or not
}
// delete service association
func deleteService(objectDefs *obj, svc *serviceOffset,  hgrpDeleted attrVal, hostname string){
    sd := objectDefs.serviceDefs
    st := objectDefs.serviceTempDefs
    ht := objectDefs.hostTempDefs
    svcEnabledDisabled := Union(&svc.enabled, &svc.disabled)
    tmplEnabledDisabled := svc.tmpl.enabledDisabled
    deleteServiceTemplate(&sd, &st, &ht, svc, tmplEnabledDisabled, hgrpDeleted, hostname)
    unregisterTemplate := attrVal{"0"}
    for v := range svcEnabledDisabled.m{
        if sd[v].attrExist("host_name"){
            sd[v]["host_name"].deleteAttrVal(&sd , &ht, v, "SVC HOSTNAME", hostname, hostname)
            if len(*sd[v]["host_name"]) == 0 {
                printDeletion(v, "SVC HOSTNAME", "host_name", "", "attr")
                delete(sd[v], "host_name")
            }
        }
        if sd[v].attrExist("hostgroup_name"){
            sd[v]["hostgroup_name"].deleteAttrVal(&sd , &ht, v, "SVC HOSTGROUP_NAME", "hostgroup_name", hostname, hgrpDeleted...)
            if len(*sd[v]["hostgroup_name"]) == 0 {
                printDeletion(v, "SVC HOSTGROUP_NAME", "hostgroup_name", "", "attr")
                delete(sd[v], "hostgroup_name")
            }
        }
        if !sd[v].attrExist("host_name") && !sd[v].attrExist("hostgroup_name"){                                    // delete hostgroup obj definition
            if sd[v].attrExist("use") {
                sd[v]["use"].deleteAttrVal(&sd, &ht, v, "SVC USE", "use", hostname, svc.tmpl.deleted...)
                if len(*sd[v]["use"]) == 0 {
                    if !sd[v].attrExist("register") || sd[v]["register"].ToString() == "1" {
                        if sd[v].attrExist("name") && isTemplateBeingUsed(&sd, &st, sd[v]["name"].ToString()){
                            sd[v]["register"] = &unregisterTemplate
                        }else{
                            printDeletion(v, "SVC USE", "use", "", "attr")
                            delete(sd[v], "use")
                            printDeletion(v, "SVC", "", "", "def")
                            svc.deleted.Add(v)
                            delete(sd, v)
                        }
                    }else {
                        if !sd[v].attrExist("name") || (sd[v].attrExist("name") && !isTemplateBeingUsed(&sd, &st,sd[v]["name"].ToString())){
                            printDeletion(v, "SVC", "","", "def")
                            svc.deleted.Add(v)
                            delete(sd, v)
                        }
                    }
                }else if isSafeDeleteTemplate(&st, *sd[v]["use"], hgrpDeleted, hostname){
                    if !sd[v].attrExist("register") || sd[v]["register"].ToString() == "1" {
                        if sd[v].attrExist("name") && isTemplateBeingUsed(&sd, &st, sd[v]["name"].ToString()){
                            sd[v]["register"] = &unregisterTemplate
                        }else{
                            printDeletion(v, "SVC", "", "", "def")
                            svc.deleted.Add(v)
                            delete(sd, v)
                        }
                    } else if isSafeDeleteTemplate(&st, *sd[v]["use"], hgrpDeleted, hostname){
                        if !sd[v].attrExist("name") || (sd[v].attrExist("name") && !isTemplateBeingUsed(&sd, &st, v)){
                            printDeletion(v, "SVC", "","", "def")
                            svc.deleted.Add(v)
                            delete(sd, v)
                        }
                    }
                }
            } else {
                if !sd[v].attrExist("register") || sd[v]["register"].ToString() == "1" {
                    if sd[v].attrExist("name") && isTemplateBeingUsed(&sd, &st, sd[v]["name"].ToString()){
                        sd[v]["register"] = &unregisterTemplate
                    }else{
                        printDeletion(v, "SVC", "", "", "def")
                        svc.deleted.Add(v)
                        delete(sd, v)
                    }
                }else {
                    if !sd[v].attrExist("name") || (sd[v].attrExist("name") && !isTemplateBeingUsed(&sd, &st,sd[v]["name"].ToString())){
                        printDeletion(v, "SVC", "","", "def")
                        svc.deleted.Add(v)
                        delete(sd, v)
                    }
                }
            }
        }
    }
}

// delete service inheritance via templates
func deleteServiceTemplate(sd *defs, st *defs, ht *defs, svc *serviceOffset, tmplEnabledDisabled attrVal, hgrpDeleted attrVal, hostname string){
    for _, t := range tmplEnabledDisabled {
        if (*st)[t].attrExist("host_name"){
            (*st)[t]["host_name"].deleteAttrVal(st , ht, t, "SVCTMPL HOSTNAME", "host_name", hostname, hostname)
            if len(*(*st)[t]["host_name"]) == 0 {
                printDeletion(t, "SVCTMPL HOSTNAME", "host_name", "", "attr")
                delete((*st)[t], "host_name")
            }
        }
        if (*st)[t].attrExist("hostgroup_name"){
            (*st)[t]["hostgroup_name"].deleteAttrVal(st , ht, t, "SVCTMPL HOSTGROUP_NAME", "hostgroup_name", hostname, hgrpDeleted...)
            if len(*(*st)[t]["hostgroup_name"]) == 0 {
                printDeletion(t, "SVCTMPL HOSTGROUP_NAME", "hostgroup_name", "", "attr")
                delete((*st)[t], "hostgroup_name")
            }
        }
        if !(*st)[t].attrExist("host_name") && !(*st)[t].attrExist("hostgroup_name"){                                    // delete hostgroup obj definition
            if !(*st)[t].attrExist("register") || (*st)[t]["register"].ToString() == "1" {
                if (*st)[t].attrExist("use") && !isTemplateBeingUsed(sd, st, t){
                    if isSafeDeleteTemplate(st, *(*st)[t]["use"], hgrpDeleted, hostname) {
                        printDeletion(t, "SVCTMPL", "", "", "def")
                        svc.tmpl.deleted.Add(t)
                        delete(*st, t)
                    }
                } else if isTemplateBeingUsed(sd, st, t){
                    unregisterTemplate := attrVal{"0"}
                    (*st)[t]["register"] = &unregisterTemplate
                    fmt.Printf("%vRegister%v:%v[SVCTMPL EDIT]%v: Unregister service template %v\n", Yellow, RST, Blue, RST, t)
                } else {
                    printDeletion(t, "SVCTMPL", "", "", "def")
                    svc.tmpl.deleted.Add(t)
                    delete(*st, t)
                }
            } else if (*st)[t].attrExist("use") && !isTemplateBeingUsed(sd, st, t){
                if isSafeDeleteTemplate(st, *(*st)[t]["use"], hgrpDeleted, hostname) {
                    fmt.Printf("%vWarning%v:%v[SVCTMPL]%v: found template not being used '%v'", Yellow, RST,Blue,RST, t)
                    // TODO: flag to allow not used template deletion
//                        printDeletion(t, "SVCTMPL", "", "", "def")
//                        svc.SetDeletedTemplate(t)
//                        delete(*st, t)
                }
            } else if !isTemplateBeingUsed(sd, st, t){
                fmt.Printf("%vWarning%v:%v[SVCTMPL]%v: found template not being used '%v'", Yellow, RST,Blue,RST, t)
            }
        }
    }
}

// check if a template is used by service object
func isTemplateBeingUsed(sd *defs, st *defs, tmplName string) bool {
    for _, def := range *st {
            if def.attrExist("use") && def["use"].Has(tmplName){
                return true
            }
    }
    for _, def := range *sd {
        if def.attrExist("use") && def["use"].Has(tmplName){
            return true
        }
    }
    return false
}

// check if its safe to delete template
func isSafeDeleteTemplate(st *defs, use attrVal, hgrpDeleted attrVal, hostname string) bool {
    for _, v := range use {
        if (*st)[v].attrExist("host_name") {
            // check if host_name contain any values other than hostname, if so, dont delete
            if !(*st)[v]["host_name"].HasOnly(hostname){
                return false
            }
        }
        if (*st)[v].attrExist("hostgroup_name") {
            // check if hostgroup_name contain any values other than the deleted hostgroups, if so, dont delete
            if !hgrpDeleted.HasAll(*(*st)[v]["hostgroup_name"]...) {
                return false
            }
        }
        if (*st)[v].attrExist("use") {
            isSafeDeleteTemplate(st, *(*st)[v]["use"], hgrpDeleted, hostname)
        }
    }
    return true
}

// check if any other host using the same regex before deletion
func isSafeDeleteRegex(hd *defs, td *defs, reStr string, id string, hostname string) bool {
    for _, def := range *hd {
        if def.attrExist("host_name") {
            for _, v := range *def["host_name"]{
                if has, _ := regexp.MatchString("^"+reStr+"$", v); has {
                    if v != hostname{
                        return false
                    }
                }
            }
        }
    }
    for _, def := range *td {
        if def.attrExist("host_name") {
            for _, v := range *def["host_name"]{
                if has, _ := regexp.MatchString("^"+reStr+"$", v); has {
                    if v != hostname{
                        return false
                    }
                }
            }
        }
    }
    return true
}
// check if its safe to delete the hostgroup delfinition.
func isSafeDeleteHostgroup(d *defs, t *defs, hostgroupName string) bool {
    // search host template for hostgroup assiciation
    for _, def := range *t {
        if def.attrExist("hostgroups") {
            if def["hostgroups"].Has(hostgroupName){
                return false
            }
        }
    }
    // search host definition for hostgroup assiciation
    for _, def := range *d {
        if def.attrExist("hostgroups") {
            if def["hostgroups"].Has(hostgroupName){
                return false
            }
        }
    }
    return true
}

// helper function to print hostgroup deletion
func printDeletion(id string, codeName string, attrName string, attrVal string, delType string, flags... string){
    switch delType {
    // print deleted object definition, attribute and value
    case "val":
        if len(flags) != 0 {
            if ToFlag(&flags).HasAll("color") {
                fmt.Printf("%vRemove%v:%v[%v]%v: Removed %v from %v\n",Red, RST, Blue, codeName, RST, attrVal, id)
            }

        }else {
            fmt.Printf("Remove:[%v]: removed %v from %v\n", codeName, attrVal,id)
        }
    // print deleted object attribute
    case "attr":
        if len(flags) != 0 {
            if ToFlag(&flags).HasAll("color") {
                fmt.Printf("%vDelete%v:%v[%v]%v: Deleted %v attribute from %v\n",Red, RST, Blue, codeName, RST, attrName, id)
            }

        }else {
            fmt.Printf("Delete:[%v]: deleted %v attribute from %v\n",codeName, attrName, id)
        }
    // print deleted object definition
    case "def":
        if len(flags) != 0 {
            if ToFlag(&flags).HasAll("color") {
                fmt.Printf("%vDelete%v:%v[%v DEFINITION]%v: Deleted object definition %v\n",Red, RST, Blue, codeName, RST, id)
            }

        }else {
            fmt.Printf("Delete:[%v]: deleted object definition %v%v%v\n", attrName, Italic, id, RST)
        }

    }
}

// Handle attribute value deletion
func (a *attrVal) deleteAttrVal(hd *defs, td *defs, id string, codeName string, attrName string, hostname string, attrVals ...string) {
    idx := (*a).FindItemIndex(attrVals...)
    for _, i := range *idx {
        if strings.HasPrefix((*a)[i], "^"){
            if isSafeDeleteRegex(hd, td, (*a)[i], id, hostname){
                printDeletion(id, codeName, attrName, (*a)[i], "val")
                RemoveItemByIndex(a, i)
            }
        }else{
            printDeletion(id, codeName, attrName, (*a)[i], "val")
            RemoveItemByIndex(a, i)

        }
    }
}

// delete hostgroup association
func deleteHostgroup(objectDefs *obj, hg *hostgroupOffset, hostname string){
    hgd := objectDefs.hostgroupDefs
    hd := objectDefs.hostDefs
    td := objectDefs.hostTempDefs
    for _, v := range hg.enabledDisabled {
        if hgd[v].attrExist("members") {
            hgd[v]["members"].deleteAttrVal(&hd, &td, v, "HGRP MEMBERS", "members", hostname, hostname)
            if len(*hgd[v]["members"]) == 0 {
                printDeletion(v, "HGRP MEMBERS", "members", "members", "attr")
                delete((hgd)[v], "members" )
                if !hgd[v].attrExist("hostgroup_members") {
                    printDeletion(v,"HGRP", "", "", "def")
                    hg.deleted.Add(v)
                    delete(hgd, v)
                    //recursive deletion for hostgroups inherited from this hostgroup
                    deleteHostgroupMembership(&hgd,&td, hg, v, hostname)
                }
            }
        }
    }
}

// deleted inhereted hostgroup
func deleteHostgroupMembership(hgd *defs, td *defs, hgrp *hostgroupOffset, hgrpName string, hostname string) {
    for _, v := range hgrp.enabledDisabled{
        if (*hgd)[v].attrExist("hostgroup_members") {
            (*hgd)[v]["hostgroup_members"].deleteAttrVal(hgd, td, v, "hostgroup_members", hostname, hgrpName)
            if len(*(*hgd)[v]["hostgroup_members"]) == 0 {
                printDeletion(v, "HGRP HOSTGROUP_MEMBERS", "hostgroup_members", "", "attr")
                delete((*hgd)[v], "hostgroup_members")
                if !(*hgd)[v].attrExist("members"){
                    printDeletion(v,"HGRP", "", "", "def")
                    hgrp.deleted.Add(v)
                    delete(*hgd, v)
                    //recursive deletion for hostgroups inherited from this hostgroup
                    deleteHostgroupMembership(hgd, td, hgrp, v, hostname)
                }
            }
        }
    }
}

// write chagnes to a file
func WriteFile(d *defs, fileName string, objName string, flags... string) {
    objType, attrLen := getMaxAttr(objName)
    f, err := os.Create(fileName); if err != nil {
        fmt.Println(err)
        f.Close()
        return
    }
    defer func(){
        err := f.Close(); if err != nil {
            fmt.Println("Failed to close file ", err)
        }
        if len(flags) != 0 && ToFlag(&flags).HasAll("color") {
            fmt.Printf("%vWrite%v: Created a new file with the changes applied '%v'\n",Green,RST,fileName)
        }else {
            fmt.Printf("Write: Created a new file with the changes applied '%v'\n",fileName)
        }
    }()
    // write defs to a file
    for _, def := range *d {
        formatDef := formatObjDef(def, objType, attrLen)
        f.WriteString(formatDef)
    }
}

//func main() {
//    flags =  append(flags, "color")
////    path := "/home/afathi/nagios-configs"
//    path := "/home/afathi/work/nagios-configs"
////    path := "test/"
////    hostname := "sdk-jenkins.sea.bigfishgames.com"
////    hostname := "java03-mongo-db05.sea.bigfishgames.com"
////    hostname := "sdk-jenkins.sea.bigfishgames.com"
////    hostname := "host3.bigfishgames.com"
////    hostname := "casino-game210.sea.bigfishgames.com"
//    hostname := "casino-elasticsearch04.sea.bigfishgames.com"
////    hostname := "f2p.bigfishgames.com"
////    hostname := "adash01.bigfish.lan"
//    excludedDir := []string {".git", "libexec", "timeperiods.cfg", "servicegroups.cfg"}
//    configFiles := findConfFiles(path, ".cfg", excludedDir)
//    data, err := readConfFile(configFiles)
//    if err != nil {
//        panic(fmt.Sprintf("%v", err))
//    }
//    objDefs,err := getObjDefs(data); if err != nil {
//        fmt.Println(err, path)
//        os.Exit(1)
//    }
//    // search for the host
//    host :=  findHost(&objDefs.hostDefs, &objDefs.hostTempDefs, hostname)
//    if host.GetHostName() == "" {
//        fmt.Println("Warning: host does not exist")
//        os.Exit(1)
//    }
//    // find hostgroup association
//    hostgroups := findHostGroups(&objDefs.hostgroupDefs, &objDefs.hostTempDefs, host)
//    // find service association
//    services := findServices(&objDefs.serviceDefs, &objDefs.serviceTempDefs, hostgroups, host.GetHostName())
//    if services.GetEnabledServiceName() == nil {
//        services.svcEnabledName = append(services.svcEnabledName, "Not Found")
//        fmt.Println(hostgroups)
//    }
//    deleteHost(&objDefs.hostDefs, &objDefs.hostTempDefs, &host)
//    fmt.Println("------------------------")
//    deleteHostgroup(objDefs, &hostgroups, host.hostName)
//    fmt.Println("------------------------")
//    deleteService(objDefs, &services, hostgroups.hgrpDeleted, host.hostName)
////    fmt.Println(*services.GetEnabledDisabledServiceName())
////    objDefs.serviceDefs.printDef("service", services.svcEnabledName...)
////      fmt.Println(len(services.svcEnabled))
////    WriteFile(&objDefs.hostDefs, "myhost.cfg", "host", "color")
////    objDefs.hostDefs.printObjDef("host")
//
//
//    printHostInfo(host.GetHostName(), hostgroups, services)
////    arr := strSlice{"a", "b", "c", "d"}
////    fmt.Println(arr)
////    arr.RemoveByVal("c")
////    fmt.Println(arr)
////    objDefs.hostTempDefs.printObjDef("host")
////    objDefs.hostgroupDefs.printObjDef("hostgroup")
////    objDefs.serviceTempDefs.printObjDef("service")
////    objDefs.contactTempDefs.printObjDef("service")
//}
