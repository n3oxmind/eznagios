package main

import (
    "strings"
    "regexp"
//    "fmt"
)
// named types that will be used through out the code
type rawDef [][]string                    // raw data read from .cfg files
type attrVal []string                     // nagois object attribute value
type def map[string]*attrVal              // nagios object definition
type defs map[string]def                  // nagios object definitons
type offset map[string][]string           // nagios object attribute offset

// nagios object definition struct
type obj struct {
    hostDefs                defs        // nagios host object definitions
    serviceDefs             defs        // nagios service object definitions
    hostgroupDefs           defs        // nagios hostgroup object definitions
    hostdependencyDefs      defs        // nagios hostdependency object definition
    servicedependencyDefs   defs        // nagios servicedependency object definition
    contactDefs             defs        // nagios contact object definition
    contactgroupDefs        defs        // nagios contactgroup object definition
    commandDefs             defs        // nagios command object definition
    hostTempDefs            defs        // nagios host template object definition
    serviceTempDefs         defs        // nagios service template object definition
    contactTempDefs         defs        // nagios contact template object definition
}

// nagios service obj struct
type serviceOffset struct {
    hostName                    offset      // service that the host is associated with via host_name attr
    hostNameExcl                offset      // excluded service that the host is not associated with via host_name attr
    hostgroupName               offset      // service that the host is associated with via hostgroup_name attr
    hostgroupNameExcl           offset      // excluded service that the host is not associated with via hostgroup_name attr
    use                         offset      // service with only 'use' attr that have association via template attributes
    templateHostgroupName       offset      // hostgroup located in service template via hostgroup_name attr
    templateHostgroupNameExcl   offset      // excluded hostgroup located in service template via hostgroup_name attr
    templateHostName            offset      // hostname located in service template via host_name attr
    templateHostNameExcl        offset      // excluded hostname located in service template via host_name attr
    svcEnabled                  offset      // active services
    svcDisabled                 offset      // excluded services
    tmplEnabledName             attrVal     // active service template name
    tmplDisabledName            attrVal     // disabled service template name
    svcEnabledName              attrVal     // active service name that belong to a specific host
    svcDisabledName             attrVal     // excluded service name from host association
    svcEnabledDisabledName      attrVal     // active and exlcuded service name
    tmplEnabledDisabledName     attrVal     // active and excluded service template name
    tmplDeleted                 attrVal     // deleted service template object definition
    svcDeleted                  attrVal     // deleted service object definition
    svcOthers                    attrVal     // service definition that has name and service_description attributes ( useful for printHostInfo )
}

// nagios hostgroup obj struct
//TODO: use set instead of offset map[string]{}
type hostgroupOffset struct {
    members                     offset      // hostgroup that the host member of via members attr
    membersExcl                 offset      // excluded hostgroup that the host not member of via members attr
    hostgroupMembers            offset      // hostgroup_name that the host member of via hostgroup_name attr
    hostgroupMembersExcl        offset      // excluded hostgroups_name that the host not member of via hostgroup_name attr
    templateHostgroups          offset      // hostgroup (acvite and excluded) defined in host template via hostgroups attr
    hgrpEnabled                 *attrVal    // active hostgroups
    hgrpDisabled                *attrVal    // excluded hostgroups
    hgrpEnabledName             attrVal     // hostgroups names that associated with a specific host
    hgrpDisabledName            attrVal     // excluded hostgroups names htat associated with a specific host
    hgrpEnabledDisabledName     attrVal     // active and excluded hgroups
    hgrpDeleted                 attrVal     // deleted hostgroup object definition
}

// nagios host obj struct
type hostOffset struct {
    hostgroups                  offset      // hostgroups defined in host object definition
    hostgroupsExcl              offset      // hostgroups defined in host object definition
    templateHostgroups          offset      // hostgroups defined in host templates via 'use' attr
    templateHostgroupsExcl      offset      // hostgroups defined in host templates via 'use' attr
    host                        offset      // host location in hostDefs
    hostDef                     def         // host definition
    hostIndex                   string      // host index
    hostName                    string      // hostname
    hostAddr                    string      // host ip address
    hgrpEnabledName             []string    // 
    hgrpDisabledName            []string    //
    templateOrder               []string       // host template index order (so inheritance dont get messed up)
}

// objDict is used to hold a collection of hosts with their association
type objDict struct{
    hosts        hostOffset
    services     serviceOffset
    hostgroups   hostgroupOffset
}

// initialize objDict
func newObjDict() *objDict {
    o := &objDict{}
    o.hosts = *newHostOffset()
    o.services = *newServiceOffset()
    o.hostgroups = *newHostGroupOffset()
    return o
}

// obj constructor
func newObj() *obj {
    o := &obj{}
    o.hostDefs = make(defs)
    o.hostTempDefs = make(defs)
    o.hostgroupDefs = make(defs)
    o.hostdependencyDefs = make(defs)
    o.serviceDefs = make(defs)
    o.serviceTempDefs = make(defs)
    o.servicedependencyDefs = make(defs)
    o.commandDefs  = make(defs)
    o.contactDefs  = make(defs)
    o.contactTempDefs  = make(defs)
    o.contactgroupDefs  = make(defs)
    return o
}

// hostOffset constructor
func newHostOffset() *hostOffset {
	o := &hostOffset{}
	o.hostIndex = ""
	o.hostName = ""              // "" indicate host name does not exist
    o.templateHostgroups =        make(offset)
    o.templateHostgroupsExcl =    make(offset)
    o.hostgroups =                make(offset)
    o.hostgroupsExcl =            make(offset)
	o.host =                      make(offset)
    return o
}

// serviceOffset constructor
func newServiceOffset() *serviceOffset {
    o := &serviceOffset{
        hostName:                   make(offset),
        hostNameExcl:               make(offset),
        hostgroupName:              make(offset),
        hostgroupNameExcl:          make(offset),
        use:                        make(offset),
        templateHostName:           make(offset),
        templateHostNameExcl:       make(offset),
        templateHostgroupName:      make(offset),
        templateHostgroupNameExcl:  make(offset),
    }
    return o
}

// hostgroupOffset constructor
func newHostGroupOffset() *hostgroupOffset {
    o := &hostgroupOffset{}
    o.members = make(offset)
    o.membersExcl = make(offset)
    o.hostgroupMembers = make(offset)
    o.hostgroupMembersExcl = make(offset)
    o.templateHostgroups = make(offset)
    o.hgrpEnabled = &attrVal{}
    o.hgrpDisabled = &attrVal{}
    o.hgrpEnabledName = attrVal{}
    o.hgrpEnabledName = attrVal{}
    return o
}

func (o *obj) SetContactTempDefs(contactTempDef def) {
    ID := contactTempDef["name"].ToString()
    o.contactDefs[ID] = contactTempDef
}

func (o *obj) SetHostTempDefs(hostTempDef def) {
    ID := hostTempDef["name"].ToString()
    o.hostTempDefs[ID] = hostTempDef
}

func (o *obj) SetServiceTempDefs(serviceTempDef def) {
    ID := serviceTempDef["name"].ToString()
    o.serviceTempDefs[ID] = serviceTempDef
}

func (o *obj) SetHostDefs(hostDef def) {
    ID := hostDef["host_name"].ToString()
    o.hostDefs[ID] = hostDef
}

func (o *obj) SetHostGroupDefs(hostgroupDef def) {
    ID := hostgroupDef["hostgroup_name"].ToString()
    o.hostgroupDefs[ID] = hostgroupDef
}

func (o *obj) SetServiceDefs(serviceDef def) {
    ID := serviceDef["service_description"].ToString()
    o.serviceDefs[ID] = serviceDef
}

func (o *obj) SetContactDefs(contactDef def) {
    ID := contactDef["contact_name"].ToString()
    o.contactDefs[ID] = contactDef
}

func (o *obj) SetContactGroupDefs(contactgroupDef def) {
    ID := contactgroupDef["contactgroup_name"].ToString()
    o.contactgroupDefs[ID] = contactgroupDef
}

func (o *obj) SetcommandDefs(commandDef def) {
    ID := commandDef["command_name"].ToString()
    o.commandDefs[ID] = commandDef
}

func (o *obj) SetHostDependencyDefs(hostdependencyDef def, idx int) {
    o.hostdependencyDefs[string(idx)] = hostdependencyDef
}

// TODO:obj dependency does not have a unique ID need to hcange the data surct to []defs
func (o *obj) SetServiceDependencyDefs(servicedependencyDef def, idx int) {
    o.servicedependencyDefs[string(idx)] = servicedependencyDef
}

// hostgroupOffset Getters
func (o *hostgroupOffset) GetMembersOffset() offset {
    return o.members
}

func (o *hostgroupOffset) GetMembersExclOffset() offset {
    return o.membersExcl
}

func (o *hostgroupOffset) GetHostgroupMembersOffset() offset {
    return o.hostgroupMembers
}

func (o *hostgroupOffset) GetHostgroupMembersExclOffset() offset {
    return o.hostgroupMembersExcl
}

func (o *hostgroupOffset) GetEnabledHostgroup() *attrVal {
    return o.hgrpEnabled
}

func (o *hostgroupOffset) GetDisabledHostgroup() *attrVal {
    return o.hgrpDisabled
}

func (o *hostgroupOffset) GetEnabledHostgroupName() attrVal {
    return o.hgrpEnabledName
}

func (o *hostgroupOffset) GetDisabledHostgroupName() attrVal {
    return o.hgrpDisabledName
}
func (o *hostgroupOffset) GetTemplateHostgroupsOffset() offset {
    return o.templateHostgroups
}
// hostgroupOffset Setters
func (o *hostgroupOffset) SetMembersOffset(id string, member string) {
    o.members[id] = append(o.members[id], member)
}

func (o *hostgroupOffset) SetMembersExclOffset(id string, member string) {
    o.membersExcl[id] = append(o.membersExcl[id], member)
}

func (o *hostgroupOffset) SetHostgroupMembersOffset(id string, member string) {
    o.hostgroupMembers[id] = append(o.hostgroupMembers[id], member)
}

func (o *hostgroupOffset) SetHostgroupMembersExclOffset(id string, member string) {
    o.hostgroupMembersExcl[id] = append(o.hostgroupMembersExcl[id], member)
}

func (o *hostgroupOffset) SetTemplateHostgroupsOffset(id string, hostgroup string) {
    o.templateHostgroups[id] = append(o.templateHostgroups[id], hostgroup)
}

func (o *hostgroupOffset) SetDeletedHostgroup(hostgroup string) {
    o.hgrpDeleted = append(o.hgrpDeleted, hostgroup)
}
func (o *hostgroupOffset) GetEnabledDisabledHostgroup() attrVal {
    return o.hgrpEnabledDisabledName
}

func (o *hostgroupOffset) GetDeletedHostgroup() attrVal {
    return o.hgrpDeleted
}

func (o *hostgroupOffset) SetEnabledHostgroup() {
    for _, m := range o.GetMembersOffset() {
        for _, v := range m{
            if !o.hgrpEnabled.Has(v){
                o.hgrpEnabled.Add(v)
            }
        }
    }
    for _, m := range o.GetHostgroupMembersOffset() {
        for _, v := range m{
            if !o.hgrpEnabled.Has(v){
                o.hgrpEnabled.Add(v)
            }
        }
    }
    // Remove excluded hostgroups
    for _, m := range o.GetHostgroupMembersExclOffset() {
        for _, v := range m{
            o.hgrpEnabled.Remove(v)
        }
    }
    for _, m := range o.GetMembersExclOffset() {
        for _, v := range m{
            o.hgrpEnabled.Remove(v)
        }
    }
    o.hgrpEnabledName = *o.hgrpEnabled
}

func (o *hostgroupOffset) SetEnabledDisabledHostgroup() {
    o.hgrpEnabledDisabledName = append(o.hgrpEnabledDisabledName, o.hgrpEnabledName...)
    o.hgrpEnabledDisabledName = append(o.hgrpEnabledDisabledName, o.hgrpDisabledName...)
}

func (o *hostOffset) SetEnabledHostgroups(hg *hostgroupOffset) {
    for _, v := range o.GetEnabledHostgroupsName(){
        if !hg.hgrpEnabled.Has(v){
            hg.hgrpEnabled.Add(v)
        }
    }
    for _, v := range o.GetDisabledHostgroupsName(){
        if hg.hgrpEnabled.Has(strings.TrimLeft(v, "!")){
            hg.hgrpEnabled.Remove(strings.TrimLeft(v, "!"))
        }
    }
    hg.hgrpEnabledName = *hg.hgrpEnabled
}

func (o *hostgroupOffset) SetDisabledHostgroup() {
    for _,m := range o.GetMembersExclOffset(){
        for _, v := range m{
            o.hgrpDisabled.Add(v)
        }
    }
    for _,m := range o.GetHostgroupMembersExclOffset() {
        for _, v := range m{
            o.hgrpDisabled.Add(v)
        }
    }
    o.hgrpDisabledName = *o.hgrpDisabled
}

// serviceOffset Getters
func (o *serviceOffset) GetTemplateHostgroupNameOffset() offset {
    return o.templateHostgroupName
}
func (o *serviceOffset) GetTemplateHostgroupNameExclOffset() offset {
    return o.templateHostgroupNameExcl
}
func (o *serviceOffset) GetTemplateHostNameOffset() offset {
    return o.templateHostName
}
func (o *serviceOffset) GetTemplateHostNameExclOffset() offset {
    return o.templateHostNameExcl
}

func (o *serviceOffset) GetHostNameOffset() offset {
    return o.hostName
}

func (o *serviceOffset) GetHostNameExclOffset() offset {
    return o.hostNameExcl
}

func (o *serviceOffset) GetHostgroupNameOffset() offset {
    return o.hostgroupName
}

func (o *serviceOffset) GetHostgroupNameExclOffset() offset {
    return o.hostgroupNameExcl
}

func (o *serviceOffset) GetEnabledServiceOffset() offset {
    return o.svcEnabled
}

func (o *serviceOffset) GetDisabledServiceOffset() offset {
    return o.svcDisabled
}

func (o *serviceOffset) GetUseOffset() offset {
    return o.use
}

func (o *serviceOffset) GetEnabledServiceName() attrVal {
    return o.svcEnabledName
}

func (o *serviceOffset) GetEnabledTemplateName() attrVal {
    return o.tmplEnabledName
}

func (o *serviceOffset) GetDisabledTemplateName() attrVal {
    return o.tmplDisabledName
}

func (o *serviceOffset) GetEnabledDisabledTemplateName() attrVal {
    return o.tmplEnabledDisabledName
}

func (o *serviceOffset) GetEnabledDisabledServiceName() *attrVal {
    return &o.svcEnabledDisabledName
}

func (o *serviceOffset) GetDisabledServiceName() attrVal {
    return o.svcDisabledName
}

// serviceOffset Setters
func (o *serviceOffset) SetHostNameOffset(id string, member string) {
    o.hostName[id] = append(o.hostName[id], member)
}

func (o *serviceOffset) SetHostNameExclOffset(id string, member string) {
    o.hostNameExcl[id] = append(o.hostNameExcl[id], member)
}

func (o *serviceOffset) SetHostgroupNameOffset(id string, member string) {
    o.hostgroupName[id] = append(o.hostgroupName[id], member)
}

func (o *serviceOffset) SetHostgroupNameExclOffset(id string, member string) {
    o.hostgroupNameExcl[id] = append(o.hostgroupNameExcl[id], member)
}

func (o *serviceOffset) SetUseOffset(id string, tmpl string) {
    o.use[id] = append(o.use[id], tmpl)
}

func (o *serviceOffset) SetTemplateHostgroupNameOffset(id string, hostgroup string) {
    o.templateHostgroupName[id] = append(o.templateHostgroupName[id], hostgroup)
}

func (o *serviceOffset) SetTemplateHostgroupNameExclOffset(id string, hostgroup string) {
    o.templateHostgroupNameExcl[id] = append(o.templateHostgroupNameExcl[id], hostgroup)
}

func (o *serviceOffset) SetTemplateHostNameOffset(id string, hostname string) {
    o.templateHostName[id] = append(o.templateHostName[id], hostname)
}

func (o *serviceOffset) SetTemplateHostNameExclOffset(id string, hostname string) {
    o.templateHostNameExcl[id] = append(o.templateHostNameExcl[id], hostname)
}

func (o *serviceOffset) SetDeletedTemplate(tmpl string) {
    o.tmplDeleted = append(o.tmplDeleted, tmpl)
}

func (o *serviceOffset) SetHybridService(svc string) {
    o.svcOthers = append(o.svcOthers, svc)
}

func (o *serviceOffset) SetDeletedService(svc string) {
    o.svcDeleted = append(o.svcDeleted, svc)
}

func (o *serviceOffset) SetEnabledServiceTemplate() {
    enabledTemplate := attrVal{}
    for _, v := range o.templateHostgroupName {
        for _,t := range v{
            if !enabledTemplate.Has(t) {
            enabledTemplate.Add(t)
            }
        }
    }
    for _, v := range o.templateHostName {
        for _,t := range v{
            if !enabledTemplate.Has(t) {
            enabledTemplate.Add(t)
            }
        }
    }
    o.tmplEnabledName = enabledTemplate
    o.tmplEnabledDisabledName = append(o.tmplEnabledDisabledName, enabledTemplate...)
}

func (o *serviceOffset) SetDisabledServiceTemplate() {
    disabledTemplate := attrVal{}
    for _, v := range o.templateHostgroupNameExcl {
        for _,t := range v{
            if !disabledTemplate.Has(t){
            disabledTemplate.Add(t)
            }
        }
    }
    for _, v := range o.templateHostNameExcl {
        for _,t := range v{
            if !disabledTemplate.Has(t){
            disabledTemplate.Add(t)
            }
        }
    }
    o.tmplDisabledName = disabledTemplate
    for _, t := range disabledTemplate {
        if !o.tmplEnabledDisabledName.Has(t){
            o.tmplEnabledDisabledName = append(o.tmplEnabledDisabledName, t)
        }
    }
}

func (o *serviceOffset) SetEnabledService() {
    // copy service assocaition via host_name
    svcEnabled := CopyMapInt(o.GetHostNameOffset())
    // add service association via hostgroup_name
    for k, v := range o.GetHostgroupNameOffset() {
        (*svcEnabled)[k] = v
    }
    // add service association via use directive (template)
    for k, v := range o.use {
        (*svcEnabled)[k] = v
    }
    // Remove excluded service declared in host_name
    for k := range o.GetHostNameExclOffset() {
        if _,exist := (*svcEnabled)[k]; exist {
            delete(*svcEnabled, k)
        }
    }
    // Remove excluded service declared in hostgroup_name
    for k := range o.GetHostgroupNameExclOffset() {
        if _,exist := (*svcEnabled)[k]; exist {
            delete(*svcEnabled, k)
        }
    }
    o.svcEnabled = *svcEnabled
    _,o.svcEnabledName = o.svcEnabled.ToSlice()
    o.svcEnabledDisabledName = o.svcEnabledName
    
}

func (o *serviceOffset) SetDisabledService() {
    svcDisabled := CopyMapInt(o.GetHostNameExclOffset())
    for k, v := range o.GetHostgroupNameExclOffset() {
        (*svcDisabled)[k] = v
    }
    o.svcDisabled = *svcDisabled
    _, o.svcDisabledName = o.svcDisabled.ToSlice()
    o.svcEnabledDisabledName = append(o.svcEnabledDisabledName, o.svcDisabledName...)
}

// hostOffset Getters
func (h *hostOffset) GetHostOffset() string {
    return h.hostIndex
}

func (h *hostOffset) GetHostDefinition() def {
    return h.hostDef
}

func (h *hostOffset) GetTemplateHostgroups() offset {
    return h.templateHostgroups
}

func (h *hostOffset) GetHostgroups() offset {
    return h.hostgroups
}

func (h *hostOffset) GetHostName() string {
    return h.hostName
}
func (h *hostOffset) GetEnabledHostgroupsName() attrVal {
    return h.hgrpEnabledName
}

func (h *hostOffset) GetDisabledHostgroupsName() attrVal {
    return h.hgrpDisabledName
}

// hostOffset Setters
func (h *hostOffset) SetHostOffset(hostIndex string) {
    h.hostIndex = hostIndex
}

func (h *hostOffset) SetHostName(hostName string) {
    h.hostName = hostName
}

func (h *hostOffset) SetTemplateHostgroupsOffset(id string, hgrp *attrVal) {
    h.templateHostgroups[id] = append(h.templateHostgroups[id], *hgrp...)
}

func (h *hostOffset) SetTemplateOrder(id string) {
    h.templateOrder = append(h.templateOrder, id)
}

func (h *hostOffset) SetTemplateHostgroupsExclOffset(id string, hgrp *attrVal) {
    h.templateHostgroupsExcl[id] = append(h.templateHostgroupsExcl[id], *hgrp...)
}

func (h *hostOffset) SetHostgroupsOffset(id string, hgrp *attrVal) {
    h.hostgroups[id] = append(h.hostgroups[id], *hgrp...)
}

func (h *hostOffset) SetHostgroupsExclOffset(id string, hgrp *attrVal) {
    h.hostgroupsExcl[id] = append(h.hostgroupsExcl[id], *hgrp...)
}

func (h *hostOffset) SetHostDefinition(hDef def) {
    h.hostDef = hDef
}

func (h *hostOffset) GetHostIndex() offset {
    return h.host
}

func (h *hostOffset) SetHostIndex(id string, hostName string) {
    h.host[id] = append(h.host[id], hostName)
}

func (o *hostOffset) SetEnabledHostgroupsName() {
    // hostgroups declared in host definition
    for _, m := range o.hostgroups {
        for _, v := range m {
            if !strings.HasPrefix(v, "!"){
                o.hgrpEnabledName = append(o.hgrpEnabledName, v)
            }else {
                o.hgrpDisabledName = append(o.hgrpDisabledName, v)
            }
        }
    }

    // hostgroups delcared in host template support multiple inheritance
    // check for addative iheritance start with '+' e.g. +hostgroup
    if (o.hgrpEnabledName == nil && o.hgrpDisabledName == nil) || strings.HasPrefix(o.hgrpEnabledName[0], "+"){
        for _,i := range o.templateOrder {
            if !strings.HasPrefix(o.templateHostgroups[i][0], "+"){
                for _, v := range o.templateHostgroups[i] {
                    if !strings.HasPrefix(v, "!"){
                        o.hgrpEnabledName = append(o.hgrpEnabledName, v)
                    }else {
                        o.hgrpDisabledName = append(o.hgrpDisabledName, v)
                    }
                }
                break
            }
            for _, v := range o.templateHostgroups[i] {
                if !strings.HasPrefix(v, "!"){
                    o.hgrpEnabledName = append(o.hgrpEnabledName, v)
                }else {
                    o.hgrpDisabledName = append(o.hgrpDisabledName, v)
                }
            }
        }
    }
}

// Add items to a slice
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

func (s *attrVal) length() int {
    return len(*s)
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
func (s *attrVal) HasAny(items ...string) bool{
    for _,v := range *s {
        for _, item := range items{
            if v == item{
                return true
            }
        }
    }
    return false
}
