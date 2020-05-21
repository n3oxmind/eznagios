package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"golang.org/x/crypto/ssh/terminal"
    "errors"
)

type multiValues []string
type boolFlag []string
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

// divid long line of usage into multiple line of width = n
func splitLongLine(s string, n int) (usage *[]string) {
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

// parse usage text based on terminal size
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
    parts := splitLongLine(usageText, width-actionWidth-rightEdgeGap)
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

func topLevelUsage() {
    maxFlagLen  := 6
    cmdSearch   := flag.Flag{Name:"search", Usage:"find services and hostgroups that belong to a specific host"}
    cmdShow     := flag.Flag{Name:"show", Usage:"show Nagios object definition"}
    cmdDelete   := flag.Flag{Name:"delete", Usage:"delete Nagios object definition/association"}
    fmt.Fprintf(os.Stderr, "Usage: %v <subcommand> --help for help on a specific subcommand\n", os.Args[0])
    fmt.Fprintf(os.Stderr, "\nsubcommands:\n")
    fmt.Fprintf(os.Stderr, "%v", *prettyUsage(cmdSearch, maxFlagLen, ""))
    fmt.Fprintf(os.Stderr, "%v", *prettyUsage(cmdShow, maxFlagLen, ""))
    fmt.Fprintf(os.Stderr, "%v", *prettyUsage(cmdDelete, maxFlagLen, ""))
}
func main() {
    bFlags := boolFlag{}         // slice to store enabled boolean flags only
    hostVal := multiValues{}
    args := []string{}
    excludedDirs := []string{".git", "libexec", "timeperiods.cfg", "servicegroups.cfg"}
//    hostname := ""
    // eznagios subcommands
    searchCommand   := flag.NewFlagSet ("search", flag.ExitOnError)
    showCommand     := flag.NewFlagSet ("show", flag.ExitOnError)
    deleteCommand   := flag.NewFlagSet ("delete", flag.ExitOnError)
    addCommand      := flag.NewFlagSet ("add", flag.ExitOnError)

    // custom usage for each subcommand
    searchCommand.Usage = func(){formatUsage(searchCommand)}
    showCommand.Usage   = func(){formatUsage(showCommand)}
    deleteCommand.Usage = func(){formatUsage(deleteCommand)}
    addCommand.Usage    = func(){formatUsage(addCommand)}

    // associate flags with their corsponding subcommand
    // search subcommand
    searchCommand.Var(&hostVal, "host", "hostname to be searched, Multiple hosts should be separated by comma/space")
    searchSourceDir     := searchCommand.String("src", "", "path to nagios configs directory")
    searchVerboseFlag   := searchCommand.Bool("verbose", false, "show verbose output")
    searchWarnFlag      := searchCommand.Bool("warn", false, "show warning messages")
    searchPrettyFlag    := searchCommand.Bool("pretty", false, "show tabular format output")
    searchColorFlag     := searchCommand.Bool("color", false, "show colorful output")
    searchFile          := searchCommand.String("file", "", "file contains list of hosts")

    // show subcommand
    showSourceDir     := showCommand.String("src", "", "path to nagios configs directory")
    showVerboseFlag   := showCommand.Bool("verbose", false, "show verbose output")
    showWarnFlag      := showCommand.Bool("warn", false, "show warning messages")

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
    default:
        fmt.Println("Error: Unrecognized subcommand")
    }

    if searchCommand.Parsed() {
        // required flags
        if len(hostVal) == 0 && *searchFile == "" {
            err := errors.New("--host or --file option is required")
            fmt.Println(&parsingError{err})
            os.Exit(1)
        }
        if *searchSourceDir == "" {
            err := errors.New("Please set up naigos configs path")
            fmt.Println(&parsingError{err})
            os.Exit(1)
        }
        // optional flags
        if *searchVerboseFlag {
            bFlags = append(bFlags, "verbose")
        }
        if *searchWarnFlag {
            bFlags = append(bFlags, "warn")
        }
        if *searchPrettyFlag {
            bFlags = append(bFlags, "pretty")
        }
        if *searchColorFlag {
            bFlags = append(bFlags, "color")
        }
        // perform serach
        configFiles := findConfFiles(*searchSourceDir, ".cfg", excludedDirs)
        rawData, err := readConfFile(configFiles)
        if err != nil {
            panic(fmt.Sprintf("%v", err))
        }
        // parse nagios config file
        objDefs, err := getObjDefs(rawData); if err != nil {
            panic(fmt.Sprintf("%v", err))
        }
        // search for host object
        host := findHost(&objDefs.hostDefs, &objDefs.hostTempDefs, hostVal[0])
        if host.GetHostName() == "" {
            err := errors.New("Host does not exist")
            fmt.Println(&DefNotFoundError{err})
        }
        // serach hostgroups association
        hostgroups := findHostGroups(&objDefs.hostgroupDefs, &objDefs.hostTempDefs, host)
        // search services association
        services := findServices(&objDefs.serviceDefs, &objDefs.serviceTempDefs, hostgroups, host.GetHostName())
        if services.GetEnabledServiceName() == nil {
            services.svcEnabledName = append(services.svcEnabledName, "Not Found")
        }
        // print host details
        printHostInfo(host.GetHostName(), hostgroups, services)


    }
    if showCommand.Parsed() {
        fmt.Println(*showSourceDir)
        fmt.Println(*showVerboseFlag)
        fmt.Println(*showWarnFlag)

    }
    if deleteCommand.Parsed() {
        fmt.Println("here")

    }
}
