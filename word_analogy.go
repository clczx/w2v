package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/clczx/learning-go/w2v/core"
	"os"
	_ "reflect"
	"sort"
	"strconv"
	"strings"
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
		text := scanner.Text()
		if text == "exit" {
			break
		}
		words := strings.Split(text, " ")
		wordPos := make([]int, 0, 10)
		for _, word := range words {
			if idx, Ok := wordMap[word]; Ok {
				fmt.Println("\t word position in vocabulary : ", idx)
				wordPos = append(wordPos, idx)
			}
		}
		if len(wordPos) != 3 {
			fmt.Println("Please enter only three words")
			continue
		}
		wordVec := core.DeepCopy(syn0Vect[wordPos[1]])
		wordVec.Increment(syn0Vect[wordPos[0]], -1)
		wordVec.Sum(syn0Vect[wordPos[2]])

		distList := make(core.DisPairList, vocabSize)
		for word, k := range wordMap {
			val := core.VecDotProduct(syn0Vect[k], wordVec)
			//fmt.Println(syn0Vect[k], syn0Vect[idx])
			//fmt.Println(word, k, val)
			distList[k] = core.DisPair{word, val}
		}

		sort.Sort(sort.Reverse(distList))
		displyNum := 20
		for i := 1; i < displyNum; i++ {
			fmt.Printf("\t%s\t%f\n", distList[i].Key, distList[i].Value)
		}
	}
}
