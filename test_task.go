package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"
)

var wg sync.WaitGroup // Объект для синхронизации горутин

func main() {

	writersPtr := flag.Int("writers", 0, "an int")
	iterCountPrt := flag.Int("iter-count", 0, "an int")
	arrSizePrt := flag.Int("arr-size", 0, "an int")

	flag.Parse() // здесь считываются параметры приложения

	var writers = *writersPtr
	var iterCount = *iterCountPrt
	var arrSize = *arrSizePrt

	Err := validate(writers, iterCount, arrSize)
	if Err != nil {
		fmt.Println(fmt.Println("Validation error occurred:", Err))
		return
	}

	if writers == 0 || iterCount == 0 || arrSize == 0 {
		fmt.Println("Execution is cancelled because its redundant") //Вычисления не имеют смысла
		return                                                      //Поэтому разумно вообще их не запускать.
	}
	fmt.Println("Started with params: ")

	fmt.Println("writers: ", writers)
	fmt.Println("iter-count: ", iterCount)
	fmt.Println("arr-size: ", arrSize)

	in := make(chan Task)

	for i := 0; i < writers; i++ {
		wg.Add(1) // Здесь добавляется один ждун
		go write(in, iterCount, i, arrSize)
	}

	go func() {
		var completeRoutinesCounter = 0 // Счетчик завершенных горутин
		var task Task
		for {
			task = <-in
			if task.IsLast { // Используем метку последней задачи для опреления момента заверщения работы горутины
				completeRoutinesCounter++
			}

			if arrSize > 0 {
				min, median, max, Err := processNumbers(task.Numbers)
				if Err != nil {
					fmt.Println("An error occurred:", Err)
				}
				fmt.Println(task.WriterIdentifier, task.CreatedTime, min, median, max)
			} else {
				// Этот случай теоретически недостижим, но он предусмотрен на случай
				// если кто-то оменит установренное мною прерывание вычислений с длинной массива 0
				fmt.Println(task.WriterIdentifier, task.CreatedTime, "no data processing results of empty array")
			}

			if completeRoutinesCounter == writers {
				return
			}
		}
	}()

	wg.Wait() // Ожидаем завершения всех горутин
}

/*
Метод преднязначен для валидайции входных данных
В целях упрощенеия проверка выполняется до обнаружения первой ошибки
*/
func validate(writers int, iterCount int, arrSize int) error {
	if writers < 0 {
		return errors.New("illegal writers value:" + strconv.Itoa(writers))
	}
	if iterCount < 0 {
		return errors.New("illegal iterCount value:" + strconv.Itoa(iterCount))
	}
	if arrSize < 0 {
		return errors.New("illegal arrSize value:" + strconv.Itoa(arrSize))
	}
	return nil
}

type Task struct {
	CreatedTime      time.Time
	WriterIdentifier string
	Numbers          []int
	IsLast           bool // Это признак того, что задача является последней от данной горутины
}

/*
Этот ментод создает задачу
и помещает ее в канал
*/
func write(out chan<- Task, num int, index int, arrSize int) {
	for i := 0; i < num; i++ {
		out <- Task{time.Now(),
			"Routine" + strconv.Itoa(index),
			getNumbers(arrSize),
			i == num-1}
	}
	defer wg.Done() // После завершения генерации нужного колическтва задач убираем одного ждуна.
}

/*
Метод динамически создает слайс заданного размера
и наполняет его случайными целыми числами
*/

func getNumbers(size int) []int {
	var slice = make([]int, size)
	for i := 0; i < size; i++ {
		slice[i] = rand.Int()
	}
	return slice
}

/*
Метод выполняет обработку данных и выводит расчитанные параметры
или генерит ошибку, если передается пустой массив,
так как в пустом массиве невозможно вычислить заданные значения
*/
func processNumbers(numbers []int) (int, int, int, error) {
	sort.Ints(numbers)
	var size = len(numbers)
	if size == 0 {
		return 0, 0, 0, errors.New("illegal arrays size 0")
	}
	return numbers[0], numbers[size/2], numbers[size-1], nil
}
