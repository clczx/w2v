package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/clczx/learning-go/w2v/core"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	TABLE_SIZE = 1e8
)

type vocabWord struct {
	cnt int
}

func initUnigramTable(wordCount core.PairList) []int {
	vocabSize := len(wordCount)
	sum := 0.0
    pow := 0.75
	for i := 0; i < vocabSize; i++ {
		sum += math.Pow(float64(wordCount[i].Value), pow)
	}

	table := make([]int, TABLE_SIZE)
	i := 0
	interv := float64(wordCount[i].Value) / float64(sum)

	for j := 0; j < TABLE_SIZE; j++ {
		table[j] = i
		if float64(j)/float64(TABLE_SIZE) > interv {
			i += 1
			interv += math.Pow(float64(wordCount[i].Value), pow) / float64(sum)
		}
		if i >= vocabSize {
			i = vocabSize - 1
		}
	}
	return table
}

func parseParams() map[string]interface{} {
	params := make(map[string]interface{})
	train := flag.String("train", "train.txt", "path of train file")
	outfile := flag.String("output", "vector_result", "path of output file")
	vectorSize := flag.Int("size", 100, "size of vector")
	windows := flag.Int("windows", 5, "size of windows")
	minCount := flag.Int("min-count", 5, "discard words less than <int>, the default is 5")
	iter := flag.Int("iter", 5, "the iteration time of training, default 5")
	negative := flag.Int("negative", 5, "the number of negative sampling")
	alpha := flag.Float64("alpha", 0.05, "Set the starting learning rate; default is 0.025 for skip-gram and 0.05 for CBOW")
	sample := flag.Float64("sample", 0.001, "set threshold for the occurrence of word, randomly down-sampling highly appear words")
	cbow := flag.Bool("cbow", true, "is set CBOW method")

	flag.Parse()

	params["train"] = *train
	params["outfile"] = *outfile
	params["vectorSize"] = int(*vectorSize)
	params["windows"] = int(*windows)
	params["negative"] = int(*negative)
	params["alpha"] = float64(*alpha)
	params["sample"] = float64(*sample)
	params["cbow"] = bool(*cbow)
	params["minCount"] = int(*minCount)
	params["iter"] = int(*iter)
	return params
}

func negativeSampling(table []int) int {
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)
	num := r.Intn(TABLE_SIZE)
	return table[num]
}

func loadTrainData(trainFile string) (map[string]int, int) {
	wordCount := make(map[string]int)

	f, err := os.Open(trainFile)
	if err != nil {
		fmt.Println("file open error")
		return wordCount, 0
	}
	defer f.Close()

	var line string

	scanner := bufio.NewScanner(f)

	trainsCount := 0
	for scanner.Scan() {
		line = scanner.Text()
		words := strings.Split(line, " ")

		for _, word := range words {
			if word != " " && word != "" {
				if count, ok := wordCount[word]; !ok {
					wordCount[word] = 1
				} else {
					wordCount[word] = count + 1
				}
				trainsCount += 1
			}
		}
	}
	return wordCount, trainsCount
}

func sortVocaWords(wordCountMap map[string]int, minCount int) (core.PairList, int) {
	vocabSize := len(wordCountMap)
	wordList := make(core.PairList, 0, vocabSize)
	for word, count := range wordCountMap {
		if count >= minCount {
			wordList = append(wordList, core.Pair{word, count})
		}
	}
	sort.Sort(sort.Reverse(wordList))
	return wordList, vocabSize
}

func savaModel(wordList core.PairList, syn0Vec []*core.Vector, outfile string) {
	f, err := os.Create(outfile)
	if err != nil {
		fmt.Printf("save Model error %s\n", err)
		return
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	vocabSize := len(wordList)
	vectLength := syn0Vec[0].Length

	w.WriteString(strconv.FormatInt(int64(vocabSize), 10))
	w.WriteString(" ")
	w.WriteString(strconv.FormatInt(int64(vectLength), 10))
	w.WriteString("\n")

	for idx, pair := range wordList {
		w.WriteString(pair.Key)
		for i := 0; i < vectLength; i++ {
			s := strconv.FormatFloat(syn0Vec[idx].Values[i], 'f', -1, 64)
			w.WriteString("\t" + s)
		}
		w.WriteString("\n")
	}
	w.Flush()
}

func initNet(vocabSize int, vecLength int) ([]*core.Vector, []*core.Vector) {
	syn0Vec := make([]*core.Vector, vocabSize)
	for i := 0; i < vocabSize; i++ {
		syn0Vec[i] = core.NewVector(vecLength)
	}

	syn1neg := make([]*core.Vector, vocabSize)
	for i := 0; i < vocabSize; i++ {
		syn1neg[i] = core.NewVector(vecLength)
	}

	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)
	for i := 0; i < vocabSize; i++ {
		for j := 0; j < vecLength; j++ {
			syn0Vec[i].Values[j] = (float64(r.Intn(65537))/float64(65536) - 0.5) / float64(vecLength)
		}
	}
	return syn0Vec, syn1neg
}

func sigmod(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

func CBOW(syn0Vec, syn1neg []*core.Vector, b, j, senLen, vecLength, negative int, alpha float64, table, sentence []int) {
	neu1 := core.NewVector(vecLength)
	neu1e := core.NewVector(vecLength)
	contextCount := 0

	word := sentence[j]

	for k := 0; k < 2*b+1; k++ {
		pos := j - b + k
		if pos == j || pos < 0 || pos >= senLen {
			continue
		}
		wordIdx := sentence[pos]
		neu1.Sum(syn0Vec[wordIdx])
		contextCount += 1
	}

	if contextCount > 0 {
		neu1.Multiply(1.0 / float64(contextCount))
	}

	var label, target int
	for k := 0; k <= negative; k++ {
		if k == 0 {
			target = word
			label = 1
		} else {
			target = negativeSampling(table)
			if target == word {
				continue
			}
			label = 0
		}

		q := sigmod(core.VecDotProduct(neu1, syn1neg[target]))
		g := alpha * (float64(label) - q)
		neu1e.Increment(syn1neg[target], g)
		syn1neg[target].Increment(neu1, g)
	}

	for k := 0; k < 2*b+1; k++ {
		pos := j - b + k
		if pos == j || pos < 0 || pos >= senLen {
			continue
		}
		wordIdx := sentence[pos]
		syn0Vec[wordIdx].Sum(neu1e)
	}

}

func skipGram(syn0Vec, syn1neg []*core.Vector, b, j, senLen, vecLength, negative int, alpha float64, table, sentence []int) {
	word := sentence[j]

	for k := 0; k < 2*b+1; k++ {
		pos := j - b + k
		if pos == j || pos < 0 || pos >= senLen {
			continue
		}
		u := sentence[pos]

		neu1e := core.NewVector(vecLength)

		var label, target int
		for d := 0; d < negative; d++ {
			if d == 0 {
				target = word
				label = 1
			} else {
				target = negativeSampling(table)
				if target == word {
					continue
				}
				label = 0
			}

			q := sigmod(core.VecDotProduct(syn0Vec[u], syn1neg[target]))
			g := alpha * (float64(label) - q)
			neu1e.Increment(syn1neg[target], g)
			syn1neg[target].Increment(syn0Vec[u], g)

		}
		syn0Vec[u].Sum(neu1e)
	}
}

func trainModel(trainFile string, vocabSize, trainsCount int, params map[string]interface{}, wordIdxMap map[string]int,
	wordList core.PairList, table []int) {
	vecLength := params["vectorSize"].(int)
	windows := params["windows"].(int)
	negative := params["negative"].(int)
	startAlpha := params["alpha"].(float64)
	sample := params["sample"].(float64)
	outputFile := params["outfile"].(string)
	cbow := params["cbow"].(bool)
	iter := params["iter"].(int)

	syn0Vec, syn1neg := initNet(vocabSize, vecLength)
	alpha := startAlpha
	localIter := iter

	fmt.Println(len(syn0Vec), len(syn1neg))

	f, err := os.Open(trainFile)
	if err != nil {
		fmt.Println("file open error")
		return
	}
	defer f.Close()

	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)

	count := 0

	for localIter > 0 {
		var line string

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line = scanner.Text()
			words := strings.Split(line, " ")

			sentence := make([]int, 0, 1000)
			for _, word := range words {
				if word != " " && word != "" {
					idx, ok := wordIdxMap[word]
					if !ok {
						continue
					}
					count += 1
					cnt := wordList[idx].Value
					if count % 10000 == 0 {
						fmt.Printf("\ralpha:%f, process : %0.2f%%, local itera : %d", alpha,
							float64(count*100)/float64(trainsCount*iter+1), iter-localIter+1)
						alpha = startAlpha * (1.0 - float64(count)/float64(trainsCount*iter+1))
						if alpha < 0.0001*startAlpha {
							alpha = 0.0001 * startAlpha
						}
					}
					if sample > 0 {
                        // down sampling
						ran := (math.Sqrt(float64(cnt)/(float64(trainsCount)*sample) + 1)) * (sample * float64(trainsCount) / float64(cnt))
						if ran < float64(r.Intn(65537))/float64(65536) {
							continue
						}
					}
					sentence = append(sentence, idx)
				}
			}

			senLen := len(sentence)

			for j := 0; j < senLen; j++ {
				b := r.Intn(windows) + 1
				if cbow {
					//CBOW
					CBOW(syn0Vec, syn1neg, b, j, senLen, vecLength, negative, alpha, table, sentence)
				} else {
					// skip-gram
					skipGram(syn0Vec, syn1neg, b, j, senLen, vecLength, negative, alpha, table, sentence)
				}
			}
		}

		localIter -= 1
		_, err := f.Seek(0, 0)
		if err != nil {
			log.Fatal("seek file error!")
		}
	}

	savaModel(wordList, syn0Vec, outputFile)
}

func main() {
	params := parseParams()
	//	for key, val := range params {
	//		fmt.Println(key, val)
	//	}

	trainFile := params["train"].(string)
	minCount := params["minCount"].(int)

	wordCountMap, trainsCount := loadTrainData(trainFile)
	wordList, vocabSize := sortVocaWords(wordCountMap, minCount)

	wordIdxMap := make(map[string]int)
	for idx, item := range wordList {
		wordIdxMap[item.Key] = idx
		//		fmt.Println(item.Key, item.Value)
	}
	table := initUnigramTable(wordList)
	//	for j := 0; j < 10; j++ {
	//		fmt.Println(negativeSampling(table))
	//	}
	//fmt.Println(table[0])
	fmt.Println("the total words is ", trainsCount)

	trainModel(trainFile, vocabSize, trainsCount, params, wordIdxMap, wordList, table)
}
