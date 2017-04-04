package core

type DisPair struct {
    Key string
    Value float64
}

type DisPairList []DisPair

func (p DisPairList) Len() int { return len(p) }

func (p DisPairList) Less(i, j int) bool {
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

func (p DisPairList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }
