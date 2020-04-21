package main

import (
    "strings"
)
// named types that will be used through out the code
type rawDef [][]string                  // raw data read from .cfg files
type def map[string]*set                // nagios object definition
type defs []def                         // nagios object definitons
type offset map[int]string              // nagios object attribute offset

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
    templateHostgroupName       offset      // hostgroup located in service template via hostgroup_name attr
    templateHostgroupNameExcl   offset      // excluded hostgroup located in service template via hostgroup_name attr
    templateHostName            offset      // hostname located in service template via host_name attr
    templateHostNameExcl        offset      // excluded hostname located in service template via host_name attr
    svcEnabled                  offset      // active services
    svcDisabled                 offset      // excluded services
    svcEnabledName              []string    // active services names that belong to a specific host
    svcDisabledName             []string    // excluded services names that belong to a specific host
}

// nagios hostgroup obj struct
type hostgroupOffset struct {
    members              offset   // hostgroup that the host member of via members attr
    membersExcl          offset   // excluded hostgroup that the host not member of via members attr
    hostgroupMembers     offset   // hostgroup_name that the host member of via hostgroup_name attr
    hostgroupMembersExcl offset   // excluded hostgroups_name that the host not member of via hostgroup_name attr
    templateHostgroups   offset   // hostgroup (acvite and excluded) defined in host template via hostgroups attr
    hgrpEnabled          *set     // active hostgroups
    hgrpDisabled         *set     // excluded hostgroups
    hgrpEnabledName      []string // hostgroups names that associated with a specific host
    hgrpDisabledName     []string // excluded hostgroups names htat associated with a specific host
}

// nagios host obj struct
type hostOffset struct {
    host               offset   // host location in hostDefs
    hostDef            def      // host definition
    hostIndex          int      // host index
    hostName           string   // hostname
    templateHostgroups []string // hostgroups defined in host templates via 'use' attr
    hostgroups         []string // hostgroups defined in host object definition
}

// hostOffset constructor
func newHostOffset() *hostOffset {
	o := &hostOffset{}
	o.host = make(offset)
	o.hostIndex = -1                    // -1 indicate host index does not exist
	o.hostName = ""                     // "" indicate host name does not exist
    return o
}

// serviceOffset constructor
func newServiceOffset() *serviceOffset {
    o := &serviceOffset{}
    o.hostName = make(offset)
    o.hostNameExcl = make(offset)
    o.hostgroupName = make(offset)
    o.hostgroupNameExcl = make(offset)
    o.templateHostgroupName = make(offset)
    o.templateHostgroupNameExcl = make(offset)
    o.templateHostName = make(offset)
    o.templateHostName = make(offset)
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
    o.hgrpEnabled =  NewSet()
    o.hgrpDisabled = NewSet()
    return o
}

// Getter and Setter for obj type
func (o *obj) ContactTempDefs() defs {
    return o.contactTempDefs
}

func (o *obj) ServiceTempDefs() defs {
    return o.serviceTempDefs
}

func (o *obj) HostTempDefs() defs {
    return o.hostTempDefs
}

func (o *obj) ServiceDefs() defs {
    return o.serviceDefs
}
func (o *obj) HostDefs() defs {
    return o.hostDefs
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

// hostgroupOffset Getters and Setters
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

func (o *hostgroupOffset) GetEnabledHostgroup() *set {
    return o.hgrpEnabled
}

func (o *hostgroupOffset) GetDisabledHostgroup() *set {
    return o.hgrpDisabled
}

func (o *hostgroupOffset) GetEnabledHostgroupName() []string {
    return o.hgrpEnabledName
}

func (o *hostgroupOffset) GetDisabledHostgroupName() []string {
    return o.hgrpDisabledName
}
func (o *hostgroupOffset) GetTemplateHostgroupsOffset() offset {
    return o.templateHostgroups
}

func (o *hostgroupOffset) SetMembersOffset(i int, member string) {
    o.members[i] = member
}

func (o *hostgroupOffset) SetMembersExclOffset(i int, member string) {
    o.membersExcl[i] = member
}

func (o *hostgroupOffset) SetHostgroupMembersOffset(i int, member string) {
    o.hostgroupMembers[i] = member
}

func (o *hostgroupOffset) SetHostgroupMembersExclOffset(i int, member string) {
    o.hostgroupMembersExcl[i] = member
}

func (o *hostgroupOffset) SetTemplateHostgroupsOffset(i int, hostgroup string) {
    o.templateHostgroups[i] = hostgroup
}

func (o *serviceOffset) SetTemplateHostgroupNameOffset(i int, hostgroup string) {
    o.templateHostgroupName[i] = hostgroup
}
func (o *serviceOffset) SetTemplateHostgroupNameExclOffset(i int, hostgroup string) {
    o.templateHostgroupNameExcl[i] = hostgroup
}
func (o *serviceOffset) SetTemplateHostNameOffset(i int, hostname string) {
    o.templateHostName[i] = hostname
}
func (o *serviceOffset) SetTemplateHostNameExclOffset(i int, hostname string) {
    o.templateHostNameExcl[i] = hostname
}


func (o *hostgroupOffset) SetEnabledHostgroup() {
    for _, v := range o.GetMembersOffset() {
        o.hgrpEnabled.Add(v)
    }
    for _, v := range o.GetHostgroupMembersOffset() {
        o.hgrpEnabled.Add(v)
    }
    // Remove excluded hostgroups
    for _, v := range o.GetHostgroupMembersExclOffset() {
        o.hgrpEnabled.Remove(v)
    }
    for _, v := range o.GetMembersExclOffset() {
        o.hgrpEnabled.Remove(v)
    }
    // add hostgroups extracted from host template nad hostgroups attr
    // templateHostgroups might includes multiple hostgroups separated with comman
    // templateHostgroups might includes repeated hostgroup name due to misconfiguration
    // set will take care of the duplicate
    for _, v := range o.GetTemplateHostgroupsOffset() {
        hgrpName := strings.Split(v, ",")
        for _,item := range hgrpName {
            if strings.HasPrefix(item,"!"){
                o.hgrpDisabled.Add(item)
            }
            o.hgrpEnabled.Add(item)
        }
    }
    o.hgrpEnabledName = o.hgrpEnabled.StringSlice()
}

func (o *hostgroupOffset) SetDisabledHostgroup() {
    for _,v := range o.GetMembersExclOffset(){
        o.hgrpDisabled.Add(v)
    }
    for _,v := range o.GetHostgroupMembersExclOffset() {
        o.hgrpDisabled.Add(v)
    }
    o.hgrpDisabledName = o.hgrpDisabled.StringSlice()
}

// serviceOffset Getters and Setters
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


func (o *serviceOffset) GetEnabledServiceName() []string {
    return o.svcEnabledName
}

func (o *serviceOffset) GetDisabledServiceName() []string {
    return o.svcDisabledName
}

func (o *serviceOffset) SetHostNameOffset(i int, member string) {
    o.hostName[i] = member
}

func (o *serviceOffset) SetHostNameExclOffset(i int, member string) {
    o.hostNameExcl[i] = member
}

func (o *serviceOffset) SetHostgroupNameOffset(i int, member string) {
    o.hostgroupName[i] = member
}

func (o *serviceOffset) SetHostgroupNameExclOffset(i int, member string) {
    o.hostgroupNameExcl[i] = member
}

// set active service checks
func (o *serviceOffset) SetEnabledService() {
    // copy service assocaition via host_name
    svcEnabled := CopyMapInt(o.GetHostNameOffset())
    // add service association via hostgroup_name
    for k, v := range o.GetHostgroupNameOffset() {
        (*svcEnabled)[k] = v
    }
    // add service association via template host_name
    for k, v := range o.GetTemplateHostNameOffset() {
        (*svcEnabled)[k] = v
    }
    // add service association via template hostgroup_name
    for k, v := range o.GetTemplateHostgroupNameOffset() {
        (*svcEnabled)[k] = v
    }
    // Remove excluded service declared in host_name
    for k, _ := range o.GetHostNameExclOffset() {
        if _,exist := (*svcEnabled)[k]; exist {
            delete(*svcEnabled, k)
        }
    }
    // Remove excluded service declared in hostgroup_name
    for k, _ := range o.GetHostgroupNameExclOffset() {
        if _,exist := (*svcEnabled)[k]; exist {
            delete(*svcEnabled, k)
        }
    }
    // Remove excluded service declared in template host_name
    for k, _ := range o.GetTemplateHostNameExclOffset() {
        if _,exist := (*svcEnabled)[k]; exist {
            delete(*svcEnabled, k)
        }
    }
    // Remove excluded service declared in template hostgroup_name
    for k, _ := range o.GetTemplateHostgroupNameExclOffset() {
        if _,exist := (*svcEnabled)[k]; exist {
            delete(*svcEnabled, k)
        }
    }
    o.svcEnabled = *svcEnabled
    _,o.svcEnabledName = o.svcEnabled.ToSlice()
    
}

// set disabled service checks
func (o *serviceOffset) SetDisabledService() {
    svcDisabled := CopyMapInt(o.GetHostNameExclOffset())
    for k, v := range o.GetHostgroupNameExclOffset() {
        (*svcDisabled)[k] = v
    }
    o.svcDisabled = *svcDisabled
    _, o.svcDisabledName = o.svcDisabled.ToSlice()
}


func (h *hostOffset) GetHostOffset() int {
    return h.hostIndex
}

func (h *hostOffset) GetHostDefinition() def {
    return h.hostDef
}

func (h *hostOffset) GetTemplateHostgroups() []string {
    return h.templateHostgroups
}

func (h *hostOffset) GetHostgroups() []string {
    return h.hostgroups
}

func (h *hostOffset) SetHostOffset(hostIndex int) {
    h.hostIndex = hostIndex
}

func (h *hostOffset) GetHostName() string {
    return h.hostName
}

func (h *hostOffset) SetHostName(hostName string) {
    h.hostName = hostName
}

func (h *hostOffset) SetTemplateHostgroups(s *set) {
    for k,_ := range (*s).m {
        h.templateHostgroups = append(h.templateHostgroups, k.(string))

    }
}

func (h *hostOffset) SetHostgroups(s *set) {
    for k,_ := range (*s).m {
        h.hostgroups = append(h.hostgroups, k.(string))

    }
}

func (h *hostOffset) SetHostDefinition(hDef def) {
    h.hostDef = hDef
}

func (h *hostOffset) GetHostIndex() offset {
    return h.host
}

func (h *hostOffset) SetHostIndex(idx int, hostName string) {
    h.host[idx] = hostName
}
