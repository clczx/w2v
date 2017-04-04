package core


type Pair struct {
    Key string
    Value int
}

type PairList []Pair

func (p PairList) Len() int { return len(p) }

func (p PairList) Less(i, j int) bool {
    return p[i].Value < p[j].Value
}

//func (p PairList) Less(i, j int) bool {
//    v := reflect.ValueOf(p[i].Value)
//
//    if v.Kind() == reflect.Int {
//        return p[i].Value.(int) < p[j].Value.(int)
//    } else if v.Kind() == reflect.Float64 {
//        return p[i].Value.(float64) < p[j].Value.(float64)
//    } else {
//        fmt.Println("type mismatch")
//    }
//    return false
//}

func (p PairList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }
