package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
	//    "reflect"
)

type multiValues []string
type boolFlagsList []string

func (s *multiValues) String() string {
    return fmt.Sprintf("%v", *s) }

func (s *multiValues) Set(value string) error {
    *s = strings.Split(value, " ")
    return nil
}


// parse os.Args, handle multiplevalues option
func parseArgs(s []string) *[]string {
    newArgs := []string{}
    previousIndex := -1
    for i, arg := range s{
        // remove item as delimiter [, , value]
        if arg == "," {
            continue
        }
        // trim ","
        if strings.HasSuffix(arg, ","){
            arg = strings.TrimRight(arg, ",")
        }
        // hold previous arg index to be used later
        previousIndex = -1
        if len(newArgs) > 1 {
            previousIndex = len(newArgs)-1
        }
        if strings.HasPrefix(arg, "-") || i == 0 {
            newArgs = append(newArgs, arg)
        } else if previousIndex != -1 && !strings.HasPrefix(newArgs[previousIndex], "-"){
            newArgs[len(newArgs)-1] = newArgs[len(newArgs)-1]+","+arg
        }else {
            newArgs = append(newArgs, arg)
        }
    }
    return &newArgs
}

type commandState struct {
    manditoryArgs           []flag.Flag
    optionalArgs            []flag.Flag
    flags                   []flag.Flag
    synopsis                string
    usageExamples           string
    maxManditoryArgLenght   int
    maxOptionalArgLenght    int
    maxFlagArgLenght        int
}

// breakLongLine divides long line of usage into multiple lines of width = n
func breakLongLine(s string, n int) (usage *[]string) {
    parts := []string{}
    slen := len(s)
    if len(s) > n {
        for i := 0 ; i <= len(s); i=i+n{
            parts = append(parts, s[i:i+n])
            slen -= n
            if slen < n {
                parts = append(parts, s[i+n:])
                break
            }
        }
    }else {
        parts = append(parts, s)
    }
    return &parts
}

// prettyUsage will dynamically fit usage to the current terminal size
func prettyUsage(opt flag.Flag, maxFlagLen int, prefix string) (optHelp *string) {
    const (
        defaultWidth = 80      // wrap text if exceeds 80 chars
        optPadding  = 2        // padding flag
        optGap = 4             // gap width between flag name and flag usage
        rightEdgeGap = 6      // wrap text 6 cols before terminal right edge
    )
    width, _, _ := terminal.GetSize(0)  // get current terminal size
    if width > defaultWidth {
        width = defaultWidth
    }
    actionWidth := optPadding + maxFlagLen + optGap
    usageText := ""
    if opt.DefValue != "false" && opt.DefValue != "" && opt.DefValue != "[]" {
        usageText = fmt.Sprintf("%v. Default(%v)", opt.Usage, opt.DefValue)
    }else {
        usageText = fmt.Sprintf("%v", opt.Usage)
    }
    parts := breakLongLine(usageText, width-actionWidth-rightEdgeGap)
    // join parts together
    joinParts := ""
    for i, part := range *parts {
        if i == 0 {
            joinParts = fmt.Sprintf("%*v%v%-*v%-*v%v\n", optPadding, "", prefix, maxFlagLen, opt.Name, optGap,"", part)
        }else {
            joinParts += fmt.Sprintf("%*v%v\n",actionWidth+len(prefix),"", part)
        }
    }
    return &joinParts
}

// custom command usage
func formatUsage(cmd *flag.FlagSet){
        cmdState := commandState{}
        cmd.VisitAll(func(f *flag.Flag){
            if f.Name == "host" || f.Name == "file"{
                cmdState.manditoryArgs = append(cmdState.manditoryArgs, *f)
                if len(f.Name) > cmdState.maxManditoryArgLenght {
                    cmdState.maxManditoryArgLenght = len(f.Name)
                }
            }else if f.Name == "verbose" || f.Name == "warn" || f.Name == "pretty" {
                cmdState.flags = append(cmdState.flags, *f)
                if len(f.Name) > cmdState.maxFlagArgLenght {
                    cmdState.maxFlagArgLenght = len(f.Name)
                }
            } else {
                cmdState.optionalArgs = append(cmdState.optionalArgs, *f)
                if len(f.Name) > cmdState.maxOptionalArgLenght {
                    cmdState.maxOptionalArgLenght = len(f.Name)
                }
            }
        })
        // usage header
        if cmd.Name() == "search" {
            fmt.Fprintf(cmd.Output(), "Usage: %v search -h <hostname> [flags...]\n", os.Args[0])
        }else if cmd.Name() == "show" {
            fmt.Fprintf(cmd.Output(), "Usage: %v show <optional arguments> [flags...] \n", os.Args[0])
        }else if cmd.Name() == "delete" {
            fmt.Fprintf(cmd.Output(), "Usage: %v delete <optional argument> [flags...] \n", os.Args[0])
        }

        // required arguments
        if cmdState.maxManditoryArgLenght > 0 {
            fmt.Fprintf(cmd.Output(), "\nmanditory arguments:\n")
            for _, opt := range cmdState.manditoryArgs {
                fmt.Fprintf(cmd.Output(), "%v", *prettyUsage(opt, cmdState.maxManditoryArgLenght, "--"))
            }
        }

        // optional arguments
        if cmdState.maxOptionalArgLenght > 0 {
        fmt.Fprintf(cmd.Output(), "\noptional arguments:\n")
            for _, opt := range cmdState.optionalArgs {
                fmt.Fprintf(cmd.Output(), "%v", *prettyUsage(opt, cmdState.maxOptionalArgLenght, "--"))
            }
        }

        // flags
        if cmdState.maxFlagArgLenght > 0 {
            fmt.Fprintf(cmd.Output(), "\nflags:\n")
            for _, opt := range cmdState.flags {
                fmt.Fprintf(cmd.Output(), "%v", *prettyUsage(opt, cmdState.maxFlagArgLenght, "--"))
            }
        }
}

// check if file exist
func isFileExist (filename string) bool {
    if _, err := os.Stat(filename); err == nil {
        return true
    } else {
        return false
    }
}

func topLevelUsage() {
    maxFlagLen  := 6
    cmdSearch   := flag.Flag{Name:"search", Usage:"find services and hostgroups that belong to a specific host"}
    cmdShow     := flag.Flag{Name:"show", Usage:"show Nagios object definition"}
    cmdDelete   := flag.Flag{Name:"delete", Usage:"delete Nagios object definition/association"}
    fmt.Fprintf(os.Stderr, "EzNagios is a tool for managing Nagios config files\n\n")
    fmt.Fprintf(os.Stderr, "Usage: %v <command> [arguments]\n", os.Args[0])
    fmt.Fprintf(os.Stderr, "\ncommands:\n")
    fmt.Fprintf(os.Stderr, "%v", *prettyUsage(cmdSearch, maxFlagLen, ""))
    fmt.Fprintf(os.Stderr, "%v", *prettyUsage(cmdShow, maxFlagLen, ""))
    fmt.Fprintf(os.Stderr, "%v", *prettyUsage(cmdDelete, maxFlagLen, ""))
    fmt.Fprintf(os.Stderr, "\nUse \"eznagios <command>\" for more information about a command.\n")
}

// set eznagios config file
func setConfigFile() string  {
    usr, err := user.Current(); if err != nil {
        panic("Failed to optain user info")
    }
    configDir := path.Join(usr.HomeDir,".config","gonag")
    configFile := path.Join(usr.HomeDir, ".config","gonag","gonag.json")
    // check if eznagios config file exist  
    if !isFileExist(configFile) {
        fmt.Printf("%vGoNagConfig:%v Created a new gonag config\n", Green, RST)
        os.MkdirAll(configDir, os.ModePerm)
        f, err := os.OpenFile(configFile, os.O_RDONLY|os.O_CREATE, 0755);
        if err != nil {
            panic("Failed to setup eznagios config file")
        }
        f.Close()
    }
    return configFile
}

// read eznagios config file
func readConfigFile(f string) *os.File {
    config, err := os.Open(f); if err != nil {
        panic(err)
    }
    return config
}

func loadEznagiosConfig() map[string]interface{}{
    // read config file; if not exist create empty file
    configFilePath := setConfigFile()
    configs := readConfigFile(configFilePath)
    loadedFlags := make(map[string]interface{})

    // load json data 
    jsonByte, _ := ioutil.ReadAll(configs)
    json.Unmarshal(jsonByte, &loadedFlags)
    // default flags values
    defaultFlags := make(map[string]interface{})
    defaultFlags["path"] = ""
    defaultFlags["warn"] = false
    defaultFlags["color"] = false
    defaultFlags["verbose"] = false
    defaultFlags["pretty"] = false

    // load default flags from eznagios config file
    if _, set := loadedFlags["path"]; set {
        defaultFlags["path"] = loadedFlags["path"]
    }
    if _, set := loadedFlags["verbose"]; set {
        defaultFlags["verbose"] = loadedFlags["verbose"]
    }
    if _, set := loadedFlags["color"]; set {
        defaultFlags["color"] = loadedFlags["color"]
    }
    if _, set := loadedFlags["warn"]; set {
        defaultFlags["warn"] = loadedFlags["warn"]
    }
    if _, set := loadedFlags["pretty"]; set {
        defaultFlags["pretty"] = loadedFlags["pretty"]
    }
    return defaultFlags
}

// set actual flags, flags that explicitly set in the command line
func setActualFlags(fs *flag.FlagSet) map[string]interface{} {
    bflags := make(map[string]struct{})
    bflags["verbose"]   = struct{}{}
    bflags["color"]     = struct{}{}
    bflags["pretty"]    = struct{}{}
    bflags["warn"]      = struct{}{}
    visited := make(map[string]interface{})
    fs.Visit(func(f *flag.Flag){
        visited[f.Name] = f.Value
    })
    // convert *flag.boolValue into bool
    for fname, fval := range visited {
        if _, ok := bflags[fname]; ok {
            fval, _ = strconv.ParseBool(fmt.Sprintf("%v", fval))
            visited[fname] = fval
        }else {
            fval := fmt.Sprintf("%v", fval)
            visited[fname] = strings.Split(fval, ",")
        }
    }
    return visited
}

// set enabled falgs
func setEnabledFlags(visited map[string]interface{}) map[string]interface{}{
    enabled := make(map[string]interface{})
    defaultFlags := loadEznagiosConfig()

    // check if a flag has been visited
    sval, sd := visited["src"]
    vval, vf := visited["verbose"]
    wval, wf := visited["warn"]
    cval, cf := visited["color"]
    pval, pf := visited["pretty"]

    if sd {
        enabled["path"] = sval
    }else if defaultFlags["path"].(string) != "" {
        enabled["path"] = defaultFlags["path"]
    }else{
        err := errors.New("Please set the default path to nagios configs using 'set' command")
        fmt.Println(&parsingError{err})
        os.Exit(1)
    }

    // optional boolean flags
    if vf && vval.(bool) || !vf && defaultFlags["verbose"].(bool) {
        enabled["verbose"] = true
    }
    if  wf && wval.(bool) || !wf && defaultFlags["warn"].(bool) {
        enabled["warn"] = true
    }
    if pf && pval.(bool) || !pf && defaultFlags["pretty"].(bool) {
        enabled["pretty"] = true
    }
    if cf && cval.(bool) || !cf && defaultFlags["color"].(bool) {
        enabled["color"] = true
    }

    return enabled
}

func main() {
//    hostVal := multiValues{}
    args := []string{}
    excludedDirs := []string{".git", "libexec", "timeperiods.cfg", "servicegroups.cfg"}

    // eznagios commands
    searchCommand   := flag.NewFlagSet ("search", flag.ExitOnError)
    showCommand     := flag.NewFlagSet ("show", flag.ExitOnError)
    deleteCommand   := flag.NewFlagSet ("delete", flag.ExitOnError)
    addCommand      := flag.NewFlagSet ("add", flag.ExitOnError)
    setCommand      := flag.NewFlagSet ("set", flag.ExitOnError)

    // custom usage for each command
    searchCommand.Usage = func(){formatUsage(searchCommand)}
    showCommand.Usage   = func(){formatUsage(showCommand)}
    deleteCommand.Usage = func(){formatUsage(deleteCommand)}
    addCommand.Usage    = func(){formatUsage(addCommand)}
    setCommand.Usage    = func(){formatUsage(setCommand)}

    // associate flags with their corsponding subcommand
    // search command
    searchCommand.String("host", "", "hostname to be searched, Multiple hosts should be separated by comma/space")
    searchCommand.String("src", "", "path to nagios configs directory")
    searchCommand.String("file", "", "file contains list of hosts")
    searchCommand.Bool("verbose", false, "show verbose output")
    searchCommand.Bool("warn", false, "show warning messages")
    searchCommand.Bool("pretty", false, "show tabular format output")
    searchCommand.Bool("color", false, "show colorful output")

    // show command
    showCommand.String("src", "", "path to nagios configs directory")
    showCommand.Bool("verbose", false, "show verbose output")
    showCommand.Bool("warn", false, "show warning messages")

    // set command
    setCommand.String("src", "", "set the default path for nagios config directory")
    setCommand.Bool("color", false, "show colorful output by default")
    setCommand.Bool("verbose", false, "show verbose output by default")
    setCommand.Bool("warn", false, "show warning message by default")

    if len(os.Args) < 2 {
//        fmt.Printf("Expected one of these subcommands %v\n", subCommandList)
        topLevelUsage()
        os.Exit(1)
    }else {
        args = *parseArgs(os.Args)
    }

    switch os.Args[1] {
    case "search":
        searchCommand.Parse(args[2:])
    case "show":
        showCommand.Parse(args[2:])
    case "add":
        addCommand.Parse(args[2:])
    case "delete":
        deleteCommand.Parse(args[2:])
    case "set":
        setCommand.Parse(args[2:])
    default:
        fmt.Println("Error: Unrecognized command")
        os.Exit(1)
    }

    if setCommand.Parsed() {
        // actual flags
        visited := setActualFlags(setCommand)
        eznagiosConfigs := loadEznagiosConfig()
        configFile      := setConfigFile()

        if _, set := visited["src"]; set {
            eznagiosConfigs["path"] = visited["path"]
            fmt.Printf("%vEzNagiosConfig:%v set '%v' as the default path to nagios-configs \n", Green, RST, visited["path"])
        }

        if val, set := visited["color"]; set {
            eznagiosConfigs["color"] = val
            if val.(bool) {
                fmt.Printf("%vEzNagiosConfig:%v set color output as the default output\n", Green, RST)
            }else {
                fmt.Printf("%vEzNagiosConfig:%v unset color output as the default output\n", Green, RST)
            }
        }

        if val, set := visited["verbose"]; set {
            eznagiosConfigs["verbose"] = val
            if val.(bool) {
                fmt.Printf("%vEzNagiosConfig:%v set verbose output as the default output\n", Green, RST)
            }else {
                fmt.Printf("%vEzNagiosConfig:%v unset verbose output as the default output\n", Green, RST)
            }
        }
        if val, set := visited["warn"]; set {
            eznagiosConfigs["warn"] = val
            if val.(bool) {
                fmt.Printf("%vEzNagiosConfig:%v set warning messages to be show by default\n", Green, RST)
            }else {
                fmt.Printf("%vEzNagiosConfig:%v unset warning messages from showing by default\n", Green, RST)
            }
        }
        // Encoding eznagios config as json
        jdata, err := json.MarshalIndent(eznagiosConfigs, "", " "); if err != nil {
            err := errors.New("Failed to update eznagios config file")
            fmt.Println(err)
            os.Exit(1)
        }
        // write configs to a file 
        jfile, err :=  os.Create(configFile); if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        jfile.Write(jdata)
    }

    if searchCommand.Parsed() {
        visited := setActualFlags(searchCommand)
        enabled := setEnabledFlags(visited)

        hval, sh := visited["host"]
        _, sf := visited["file"]
        // required flags
        if !sh && !sf {
            err := errors.New("--host or --file option is required")
            fmt.Println(&parsingError{err})
            os.Exit(1)
        }

        // perform serach
        configFiles := findConfFiles(enabled["path"].(string), ".cfg", excludedDirs)
        rawData, err := readConfFile(configFiles)
        if err != nil {
            panic(fmt.Sprintf("%v", err))
        }
        // parse nagios config file
        objDefs, err := getObjDefs(rawData); if err != nil {
            panic(fmt.Sprintf("%v", err))
        }
        // parse host args
        knownHosts, unknownHosts, noRegex := parseRegex(hval.([]string), &objDefs.hostDefs)
        dictList := []objDict{}
        for _,h := range knownHosts {
            dict := objDict{}
            // search for host object
            host := findHost(&objDefs.hostDefs, &objDefs.hostTempDefs, h)
            // serach hostgroups association
            hostgroups := findHostGroups(&objDefs.hostgroupDefs, &objDefs.hostTempDefs, host)
            // search services association
            services := findServices(&objDefs.serviceDefs, &objDefs.serviceTempDefs, hostgroups, host.GetHostName())
            if services.GetEnabledServiceName() == nil {
                services.svcEnabledName = append(services.svcEnabledName, "Not Found")
            }
            // multiple hosts are stored in a dictionary like format
            dict.hosts = host
            dict.services = services
            dict.hostgroups = hostgroups
            dictList = append(dictList, dict)

            if _, ok := visited["pretty"]; !ok {
                printHostInfo(host.GetHostName(), host.hostDef["address"].ToString(), hostgroups, services)
            }
        }
        if _, ok := visited["pretty"]; ok {
            terminalWidth,_,_ := terminal.GetSize(0)
            printHostInfoPretty(dictList, terminalWidth)
        }
        // print tabular output format
        fmt.Printf("\nNum of hosts: %v\n\n", len(knownHosts))
        //print errors
        for _, v := range unknownHosts {
            err := errors.New("host not found")
            fmt.Println(&NotFoundError{err, "Warn", v})
        }
        for _, v := range noRegex {
            err := errors.New("regex match nothing")
            fmt.Println(&NotFoundError{err, "Warn", v})
        }
    }
    if showCommand.Parsed() {

    }
    if deleteCommand.Parsed() {
        fmt.Println("here")

    }
}

// parseRexec will parse the host args regardless whether the args are regex or not
func parseRegex (s []string, d *defs) ([]string, []string, []string){
    pattern  := regexp.MustCompile(`\{|\[|\*|\^|\(`)
    knownHosts := []string{}            // any host that does exist will be stored here
    unknownHosts := []string{}          // any host that does not exist will be stored here
    reNoMatch := []string{}             // any regex that does not match host obj will be stored here
    for _, val := range s {
        found := false
        if pattern.MatchString(val) {
            for _, def := range *d {
                if def.attrExist("host_name") {
                    for _, hostname := range *def["host_name"] {
                        if match,_ := regexp.MatchString(val, hostname); match {
                            knownHosts = append(knownHosts, hostname)
                            found = true
                            break
                        }
                    }
                }
            }
            if !found {
                reNoMatch = append(reNoMatch, val)
            }
        }else {
            if _, ok := (*d)[val]; ok {
                knownHosts = append(knownHosts, val)
            }else{
                unknownHosts = append(unknownHosts, val)
            }
        }
    }
    // sort slice
    sort.Strings(knownHosts)
    sort.Strings(unknownHosts)
    sort.Strings(reNoMatch)
    return knownHosts, unknownHosts, reNoMatch

}
