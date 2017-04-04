package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/clczx/learning-go/w2v/core"
	"os"
	_ "reflect"
	"strconv"
	"strings"
    "sort"
)

func loadVectorFile(inputFile string) ([]*core.Vector, map[string]int) {
	f, err := os.Open(inputFile)

	if err != nil {
		fmt.Println("file open error")
		return nil, nil
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	line := scanner.Text()
	items := strings.Split(line, " ")
	vocabSize, _ := strconv.ParseInt(items[0], 10, 0)
	vectLength, _ := strconv.ParseInt(items[1], 10, 0)

	wordMap := make(map[string]int)
	syn0Vect := make([]*core.Vector, vocabSize)

	fmt.Println(vocabSize, vectLength)
	k := 0
	for scanner.Scan() {
		line = scanner.Text()
		items = strings.Split(line, "\t")
		word := items[0]
		wordMap[word] = k
		syn0Vect[k] = core.NewVector(int(vectLength))
		for i := 0; i < int(vectLength); i++ {
			val, err := strconv.ParseFloat(items[i+1], 64)
			if err == nil {
				syn0Vect[k].Values[i] = val
			}
		}
		syn0Vect[k].Norm()
		k += 1
	}
	fmt.Println(len(syn0Vect))
	return syn0Vect, wordMap
}

func main() {

	binFile := flag.String("input", "vector_result", "word vector file")

	flag.Parse()

	syn0Vect, wordMap := loadVectorFile(*binFile)

	scanner := bufio.NewScanner(os.Stdin)
	vocabSize := len(syn0Vect)
	for {
		fmt.Print("please enter the words: ")
		scanner.Scan()
		word := scanner.Text()
		if word == "exit" {
			break
		}
		if idx, Ok := wordMap[word]; Ok {
			fmt.Println("\t word position in vocabulary : ", idx)
			distList := make(core.DisPairList, vocabSize)
			for word, k := range wordMap {
				val := core.VecDotProduct(syn0Vect[k], syn0Vect[idx])
                //fmt.Println(syn0Vect[k], syn0Vect[idx])
                //fmt.Println(word, k, val)
				distList[k] = core.DisPair{word, val}
			}

			sort.Sort(sort.Reverse(distList))
			displyNum := 20
			for i := 1; i < displyNum; i++ {
				fmt.Printf("\t%s\t%f\n", distList[i].Key, distList[i].Value)
			}
		} else {
			fmt.Println("\t word not found in vocabulary")
		}
	}
}
