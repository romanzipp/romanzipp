<h3>
    
```go
​
package main

import (
  "fmt"
  "encoding/json"
)

type Person struct {
  name   string
  status []string
  stack  Stack
}

type Stack struct {
  languages string[]
}

func main() {

  me := &Person {
    name:   "Roman Zipp"
    status: []string{ "Web Developer", "Student for Business Informatics" }
    stack:  Stack{
      languages: []string{ "PHP", "JS", "Ruby", "Go" }
    }
  }

  data, _ := json.Marshal(*me)
  
  fmt.Println(string(data))
}
​
```
</h3>
