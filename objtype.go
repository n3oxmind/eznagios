package main

import (
    "strings"
    "regexp"
//    "fmt"
)
// named types that will be used through out the code
type rawDef [][]string                    // raw data read from .cfg files
type attrVal []string
type def map[string]*attrVal              // nagios object definition
type defs []def                           // nagios object definitons
type offset map[int][]string              // nagios object attribute offset in defs

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
    svcEnabledName              attrVal    // active services names that belong to a specific host
    svcDisabledName             attrVal    // excluded services names that belong to a specific host
}
// nagios hostgroup obj struct
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
}
// nagios host obj struct
type hostOffset struct {
    hostgroups                  offset      // hostgroups defined in host object definition
    hostgroupsExcl              offset      // hostgroups defined in host object definition
    templateHostgroups          offset      // hostgroups defined in host templates via 'use' attr
    templateHostgroupsExcl      offset      // hostgroups defined in host templates via 'use' attr
    host                        offset      // host location in hostDefs
    hostDef                     def         // host definition
    hostIndex                   int         // host index
    hostName                    string      // hostname
    hgrpEnabledName             []string    // 
    hgrpDisabledName            []string    //
    templateOrder               []int       // host template index order (so inheritance dont get messed up)
}
// hostOffset constructor
func newHostOffset() *hostOffset {
	o := &hostOffset{}
	o.hostIndex = -1             // -1 indicate host index does not exist
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
    o.contactTempDefs = append(o.contactTempDefs,contactTempDef)
}

func (o *obj) SetHostTempDefs(hostTempDef def) {
    o.hostTempDefs = append(o.hostTempDefs,hostTempDef)
}

func (o *obj) SetServiceTempDefs(serviceTempDef def) {
    o.serviceTempDefs = append(o.serviceTempDefs,serviceTempDef)
}

func (o *obj) SetHostDefs(hostDef def) {
    o.hostDefs = append(o.hostDefs, hostDef)
}

func (o *obj) SetHostGroupDefs(hostgroupDef def) {
    o.hostgroupDefs = append(o.hostgroupDefs, hostgroupDef)
}

func (o *obj) SetServiceDefs(serviceDef def) {
    o.serviceDefs = append(o.serviceDefs, serviceDef)
}

func (o *obj) SetContactDefs(contactDef def) {
    o.contactDefs = append(o.contactDefs, contactDef)
}

func (o *obj) SetContactGroupDefs(contactgoupDef def) {
    o.contactgroupDefs = append(o.contactgroupDefs, contactgoupDef)
}

func (o *obj) SetcommandDefs(commandDef def) {
    o.commandDefs = append(o.commandDefs, commandDef)
}

func (o *obj) SetHostDependencyDefs(hostdependencyDef def) {
    o.hostdependencyDefs = append(o.hostdependencyDefs, hostdependencyDef)
}

func (o *obj) SetServiceDependencyDefs(servicedependencyDef def) {
    o.servicedependencyDefs = append(o.servicedependencyDefs, servicedependencyDef)
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
func (o *hostgroupOffset) SetMembersOffset(i int, member string) {
    o.members[i] = append(o.members[i], member)
}

func (o *hostgroupOffset) SetMembersExclOffset(i int, member string) {
    o.membersExcl[i] = append(o.membersExcl[i], member)
}

func (o *hostgroupOffset) SetHostgroupMembersOffset(i int, member string) {
    o.hostgroupMembers[i] = append(o.hostgroupMembers[i], member)
}

func (o *hostgroupOffset) SetHostgroupMembersExclOffset(i int, member string) {
    o.hostgroupMembersExcl[i] = append(o.hostgroupMembersExcl[i], member)
}

func (o *hostgroupOffset) SetTemplateHostgroupsOffset(i int, hostgroup string) {
    o.templateHostgroups[i] = append(o.templateHostgroups[i], hostgroup)
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

func (o *serviceOffset) GetDisabledServiceName() attrVal {
    return o.svcDisabledName
}

// serviceOffset Setters
func (o *serviceOffset) SetHostNameOffset(i int, member string) {
    o.hostName[i] = append(o.hostName[i], member)
}

func (o *serviceOffset) SetHostNameExclOffset(i int, member string) {
    o.hostNameExcl[i] = append(o.hostNameExcl[i], member)
}

func (o *serviceOffset) SetHostgroupNameOffset(i int, member string) {
    o.hostgroupName[i] = append(o.hostgroupName[i], member)
}

func (o *serviceOffset) SetHostgroupNameExclOffset(i int, member string) {
    o.hostgroupNameExcl[i] = append(o.hostgroupNameExcl[i], member)
}

func (o *serviceOffset) SetUseOffset(i int, tmpl string) {
    o.use[i] = append(o.use[i], tmpl)
}

func (o *serviceOffset) SetTemplateHostgroupNameOffset(i int, hostgroup string) {
    o.templateHostgroupName[i] = append(o.templateHostgroupName[i], hostgroup)
}

func (o *serviceOffset) SetTemplateHostgroupNameExclOffset(i int, hostgroup string) {
    o.templateHostgroupNameExcl[i] = append(o.templateHostgroupNameExcl[i], hostgroup)
}

func (o *serviceOffset) SetTemplateHostNameOffset(i int, hostname string) {
    o.templateHostName[i] = append(o.templateHostName[i], hostname)
}

func (o *serviceOffset) SetTemplateHostNameExclOffset(i int, hostname string) {
    o.templateHostNameExcl[i] = append(o.templateHostNameExcl[i], hostname)
}

func (o *serviceOffset) SetEnabledServiceTemplate() {
    enabledTemplate := attrVal{}
    for _, v := range o.templateHostgroupName {
        for _,t := range v{
            enabledTemplate.Add(t)
        }
    }
    for _, v := range o.templateHostName {
        for _,t := range v{
            enabledTemplate.Add(t)
        }
    }
    o.tmplEnabledName = enabledTemplate
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
    
}

func (o *serviceOffset) SetDisabledService() {
    svcDisabled := CopyMapInt(o.GetHostNameExclOffset())
    for k, v := range o.GetHostgroupNameExclOffset() {
        (*svcDisabled)[k] = v
    }
    o.svcDisabled = *svcDisabled
    _, o.svcDisabledName = o.svcDisabled.ToSlice()
}

// hostOffset Getters
func (h *hostOffset) GetHostOffset() int {
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
func (h *hostOffset) SetHostOffset(hostIndex int) {
    h.hostIndex = hostIndex
}

func (h *hostOffset) SetHostName(hostName string) {
    h.hostName = hostName
}

func (h *hostOffset) SetTemplateHostgroupsOffset(idx int, hgrp *attrVal) {
    h.templateHostgroups[idx] = append(h.templateHostgroups[idx], *hgrp...)
}

func (h *hostOffset) SetTemplateOrder(idx int) {
    h.templateOrder = append(h.templateOrder, idx)
}

func (h *hostOffset) SetTemplateHostgroupsExclOffset(idx int, hgrp *attrVal) {
    h.templateHostgroupsExcl[idx] = append(h.templateHostgroupsExcl[idx], *hgrp...)
}

func (h *hostOffset) SetHostgroupsOffset(idx int, hgrp *attrVal) {
    h.hostgroups[idx] = append(h.hostgroups[idx], *hgrp...)
}

func (h *hostOffset) SetHostgroupsExclOffset(idx int, hgrp *attrVal) {
    h.hostgroupsExcl[idx] = append(h.hostgroupsExcl[idx], *hgrp...)
}

func (h *hostOffset) SetHostDefinition(hDef def) {
    h.hostDef = hDef
}

func (h *hostOffset) GetHostIndex() offset {
    return h.host
}

func (h *hostOffset) SetHostIndex(idx int, hostName string) {
    h.host[idx] = append(h.host[idx], hostName)
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
func (s *attrVal) HasAny(items []string) bool{
    for _,v := range *s {
        for _, item := range items{
            if v == item{
                return true
            }
        }
    }
    return false
}
