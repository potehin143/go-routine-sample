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
		fmt.Println("Validation error occurred:", Err)
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

	in := make(chan interface{})

	for i := 0; i < writers; i++ {
		wg.Add(1) // Здесь добавляется один ждун
		go write(in, iterCount, i, arrSize)
	}

	go func() {
		var completeRoutinesCounter = 0 // Счетчик завершенных горутин
		for {
			something := <-in

			switch value := something.(type) {
			case Processable:
				if arrSize > 0 {
					min, median, max, Err := processNumbers(value.Numbers())
					if Err != nil {
						fmt.Println("An error occurred:", Err)
					}
					fmt.Println(value.WriterIdentifier(), value.CreatedTime(), min, median, max)
				} else {
					// Этот случай теоретически недостижим, но он предусмотрен на случай
					// если кто-то оменит установренное мною прерывание вычислений с длинной массива 0
					fmt.Println(value.WriterIdentifier(), value.CreatedTime(),
						"no data processing results of empty array")
				}
				break
			case bool:
				if value { // Используем метку завершения для определения момента прекращения работы горутины
					completeRoutinesCounter++
				}
				break
			case error: // Обработка ошибок, возникающих во время генерации
				fmt.Println("Data generation error occurred:", value)
				// Важный момент! Получение ошибки не считается признаком окончания передачи данных
				// Такое решение принято для того, чтобы была только одна точка завершения горутины
				// Кроме того в теории дает сбора и обработки множества ошибок из одной горутины
				break
			default:
				fmt.Println("Unexpected value received:", value)
			}

			if completeRoutinesCounter == writers {
				close(in) // Закрываем канал
				in = nil
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

type Processable interface {
	CreatedTime() time.Time
	WriterIdentifier() string
	Numbers() *[]int // Используем указатель на массим, чтобы не копировать массив при передаче
}

type Task struct {
	createdTime      time.Time
	writerIdentifier string
	numbers          *[]int
}

func (task Task) CreatedTime() time.Time {
	return task.createdTime
}

func (task Task) WriterIdentifier() string {
	return task.writerIdentifier
}

func (task Task) Numbers() *[]int {
	return task.numbers
}

/*
Этот ментод создает задачу
и помещает ее в канал
*/
func write(out chan<- interface{}, num int, index int, arrSize int) {

	id := routineId(index)

	if num >= 0 && arrSize >= 0 {
		for i := 0; i < num; i++ {
			out <- func() Processable { // Здесь я решил использовать обертку из анонимной функции, чтобы
				// в случае изменения интерфейса отследить ошибку компиляции
				return Task{time.Now(),
					id,
					getNumbers(arrSize)}
			}() // Эти круглые скобки нужны, чтобы передавалось значение функции, а не сама функция
		}
	}
	if num < 0 { // Отправка в канал ошибок генерации.
		out <- errors.New("routine " + id + " got illegal num value:" + strconv.Itoa(num))
	}
	if arrSize < 0 {
		out <- errors.New("routine " + id + " got illegal arrSize value:" + strconv.Itoa(arrSize))
	}
	// такая компоновка if else выбрана для того, чтобы была единая точка завершения горутины и снятия блокировки
	//При этом мы можем вообщить обо всех ошибках

	out <- true     // Признак завершения передачи
	defer wg.Done() // После завершения генерации нужного количества задач убираем одного ждуна.
}

func routineId(index int) string {
	return "Routine" + strconv.Itoa(index)
}

/*
Метод динамически создает слайс заданного размера
и наполняет его случайными целыми числами
*/

func getNumbers(size int) *[]int {
	var slice = make([]int, size)
	for i := 0; i < size; i++ {
		slice[i] = rand.Int()
	}
	return &slice
}

/*
Метод выполняет обработку данных и выводит расчитанные параметры
или генерит ошибку, если передается пустой массив,
так как в пустом массиве невозможно вычислить заданные значения
*/
func processNumbers(numbers *[]int) (int, int, int, error) {
	slice := *numbers
	sort.Ints(slice)
	var size = len(slice)
	if size == 0 {
		return 0, 0, 0, errors.New("illegal arrays size 0")
	}
	return slice[0], slice[size/2], slice[size-1], nil
}
