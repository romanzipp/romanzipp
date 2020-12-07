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

[![](https://img.shields.io/twitter/follow/romanzipp?label=Twitter&style=social)](https://twitter.com/romanzipp)
[![](https://img.shields.io/github/followers/romanzipp?label=Github&style=social)](https://github.com/romanzipp)
[![](https://img.shields.io/website?label=ich.wtf&up_message=up&url=https%3A%2F%2Fich.wtf)](https://ich.wtf)
