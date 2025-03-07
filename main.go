package main

import (
  "os"
  "fmt"
  "strings"
  "strconv"
)

// Registers
  // TAR - Temp A Reg
  // TBR
  // TCR
  // TDR
  // GAI - General A Index
  // GBI
  // GAP - General A Pointer
  // GBP

var registers map[string]int = map[string]int {
  "tar": 0,
  "tbr": 0,
  "tcr": 0,
  "tdr": 0,

  "gai": 0,
  "gbi": 0,
  "ret": 0,
}

var flags map[string]bool = map[string]bool {
  "je": false,
  "jne": false,
  "jg": false,
  "jl": false,
  "jge": false,
  "jle": false,
}
// Maybe make like registers
var gap, gbp string

type stackType []int
var stack stackType
var labels map[string]int = make(map[string]int)
var vars   map[string]int = make(map[string]int)
var arrays map[string][]int = make(map[string][]int)

var mainLine int = 0
var lineNum int = 0

var input []string

func keyInMap(mapToSearch map[string]int, val string) bool {
  _, ok := mapToSearch[val]
  return ok
}

func keyInMapArray(mapToSearch map[string][]int, val string) bool {
  _, ok := mapToSearch[val]
  return ok
}

func (s stackType) Push(v int) stackType {
    return append(s, v)
}

func (s stackType) Pop() (stackType, int) {
    l := len(s)

    if l == 0 {
      fmt.Printf("Cannot pop from empty stack: %s", input[lineNum])
      os.Exit(1)
    }

    return  s[:l-1], s[l-1]
}

func argument(arg string) int {
  if keyInMap(registers, arg) {return registers[arg]}

  num, err := strconv.Atoi(arg)

  if err == nil {
    return num
  }

  return vars[arg]
}

func arrayArgument(arg string) []int {
  if keyInMapArray(arrays, arg) {return arrays[arg]} 
  if arg == "gap" {return arrays[gap]} 
  if arg == "gbp" {return arrays[gbp]}
  
  fmt.Printf("Invalid array or register: %s", arg)
  os.Exit(1)
  return []int{1}
}

func main() {
  if len(os.Args) < 2 {
    fmt.Printf("Use: %s file.fkasm\n", os.Args[0])
    os.Exit(1)
  }

  data, err := os.ReadFile(os.Args[1])

  if err != nil {
    fmt.Printf("Cannot read from %s", os.Args[1])
  }

  unCheckedinput := strings.Split(string(data), "\n") 


  // If line blank or comment, leave out
  // If line begins with label or main, take note
  for _, line := range unCheckedinput {
    trimmedLine := strings.TrimSpace(line)
    if trimmedLine != "" && string(trimmedLine[0]) != ";" {
      splitLine := strings.Split(trimmedLine, " ")
      if splitLine[0] == "label" {
        labels[splitLine[1]] = len(input)
      } else if splitLine[0] == "main" {
        mainLine = len(input)
      }
      input = append(input, trimmedLine)
    }
  } 
  
  for lineNum = mainLine + 1; lineNum < len(input); lineNum++ {
    tokens := strings.Split(input[lineNum], " ")
    //fmt.Println(tokens) // uncomment to see stack, debug
   switch tokens[0] {
    case "goto":
      registers["ret"] = lineNum
      
      switch tokens[1] {
      case "gap":
        lineNum = labels[gap]
      case "gbp":
        lineNum = labels[gbp]
      default:
        if keyInMap(labels, tokens[1]) {
          lineNum = labels[tokens[1]]
        } else {
          lineNum = argument(tokens[1])
        }
      }
    case "ret":
      lineNum = registers["ret"]
    case "int":
      vars[tokens[1]] = argument(tokens[2])
    case "str":
      var finalString []int 
      isSlashed := false
      
      for _, char := range(strings.Join(tokens[2:], " ")[1:]) {
        if char == '\\' {
          isSlashed = true
          continue
        }
        if char != '"' && !isSlashed {
          finalString = append(finalString, int(char))
        } else {
          break
        }
      }
      arrays[tokens[1]] = finalString
    case "exit":
      errCode, err := strconv.Atoi(tokens[1])
      if err != nil {
        fmt.Printf("On line %d exit without int\n", lineNum)
        os.Exit(1)
      }
      os.Exit(errCode)
    case "print":
      for _, x := range(tokens[1:]) {
        fmt.Print(argument(x)) 
      }
    case "printchar":
      for _, x := range(tokens[1:]) {
        fmt.Print(string(rune(argument(x))))
      }
    case "mov":
      // Refactor for gap and gbp
      if keyInMap(registers, tokens[1]) {
        registers[tokens[1]] = argument(tokens[2])
      } else if keyInMap(vars, tokens[1]) {
        vars[tokens[1]] = argument(tokens[2])
      }
    case "movarr":
      if keyInMapArray(arrays, tokens[1]) {
        arrays[tokens[1]] = arrayArgument(tokens[2])
      } else if tokens[1] == "gap" {
        var finalString string
        for _, x := range(arrayArgument(tokens[2])) {
          finalString += string(rune(x))
        }
        gap = finalString
      } else if tokens[1] == "gbp" {
        var finalString string
        for _, x := range(arrayArgument(tokens[2])) {
          finalString += string(rune(x))
        }
        gbp = finalString
      }
    case "push":
      for _, x := range(tokens[1:]) {
        stack = stack.Push(argument(x))
      }
    case "pop":
      for _, x := range(tokens[1:]) {
        var popVal int
        stack, popVal = stack.Pop()

        if keyInMap(registers, x) {
          registers[x] = popVal
        } else {
          vars[x] = popVal
        }
      }
     
    case "cmp":
      arg1 := argument(tokens[1])
      arg2 := argument(tokens[2])
      
      equals := arg1 == arg2
      greater := arg1 > arg2
      
      flags["je"]  = equals
      flags["jg"]  = greater
      flags["jge"] = equals || greater
      flags["jl"]  = !greater && !equals
      flags["jle"] = !greater
      flags["jne"] = !equals
    case "add":
      addNum := argument(tokens[2])
      if keyInMap(registers, tokens[1]) {
        registers[tokens[1]] += addNum
      } else if keyInMap(vars, tokens[1]) {
        vars[tokens[1]] += addNum
      }
    case "sub":
      subNum := argument(tokens[2])
      if keyInMap(registers, tokens[1]) {
        registers[tokens[1]] -= subNum
      } else if keyInMap(vars, tokens[1]) {
        vars[tokens[1]] -= subNum
      }
    case "len":
      registers["tar"] = len(arrayArgument(tokens[1]))
    case "index":
      registers["tar"] = arrayArgument(tokens[1])[argument(tokens[2])]
    default:
      boolVal, ok := flags[tokens[0]]
      
      if ok && boolVal {
        token, err := strconv.Atoi(tokens[1])
        registers["ret"] = lineNum 
      
        if err != nil {
          lineNum = labels[tokens[1]]
        } else {
          lineNum = token - 1 
        }

      }
    }
  }
  

  
}
