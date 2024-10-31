// Código exemplo para o trabaho de sistemas distribuidos (eleicao em anel)
// By Cesar De Rose - 2022

package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type mensagem struct {
	tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmacao da eleicao)
	corpo [4]int // conteudo da mensagem para colocar os ids (usar um tamanho ocmpativel com o numero de processos no anel)
	vencedor int
}

var (
	chans = []chan mensagem{ // vetor de canias para formar o anel de eleicao - chan[0], chan[1] and chan[2] ...
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
	}
	controle = make(chan int)
	wg       sync.WaitGroup // wg is used to wait for the program to finish
)

func ElectionControler(in chan int) {
	defer wg.Done()

	var temp mensagem

	// comandos para o anel iciam aqui

	// mudar o processo 0 - canal de entrada 3 - para falho (defini mensagem tipo 2 pra isto)

	temp.tipo = 3
	chans[3] <- temp

	for {
		control := <-in 

		if(control == 1) {
			falha_eleicao(3, temp)
		}

		if(control == 2) {
			falha_eleicao(0, temp)
		}

		if(control == 3) {
			temp.tipo = 4
			for _, ch := range chans {
				time.Sleep(1 * time.Second)
				ch <- temp
			}
			return
		}

	}
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
	defer wg.Done()
	var bFailed bool = false // todos inciam sem falha
	// variaveis locais que indicam se este processo é o lider e se esta ativo
	var actualLeader int
	actualLeader = leader // indicação do lider veio por parâmatro

	for {
		temp := <-in // ler mensagem
		switch temp.tipo {
		// Eleicao
		case 1:
			{
				time.Sleep(1 * time.Second)
				if !bFailed { // Se eu nao falhei posso participar da eleicao
					if temp.corpo[TaskId-1] != TaskId { // Se ainda nao deu a volta no anel
						temp.corpo[TaskId-1] = TaskId
						fmt.Printf("%2d: ELEIÇÃO: [ %d, %d, %d, %d ]\n", TaskId, temp.corpo[0], temp.corpo[1], temp.corpo[2], temp.corpo[3])
						out <- temp
					} else { // Se deu a volta no anel
						menorValor := math.MaxInt // Inicializa com o maior valor possível
						for _, valor := range temp.corpo {
							if valor < menorValor && valor != 0{
								menorValor = valor
							}
						}
						temp.vencedor = menorValor
						temp.tipo = 3
						temp.corpo[0] = 0
						temp.corpo[1] = 0
						temp.corpo[2] = 0
						temp.corpo[3] = 0
						fmt.Printf("%2d: VENCEDOR DA ELEIÇÃO - [%2d ]\n", TaskId, temp.vencedor)
						out <- temp
						//controle <- 2
					}
					} else {
						fmt.Printf("%2d: Estou falho, não posso pariticpar da eleição\n", TaskId)
						out <- temp
					}
			}

		// Falha
		case 2:
			{
				time.Sleep(1 * time.Second)
				bFailed = true
				fmt.Printf("%2d: Falhei\n", TaskId)
			}

		// Roda Processo
		case 3:
			{
				time.Sleep(1 * time.Second)
				if temp.corpo[TaskId-1] != TaskId { // Se ainda nao deu a volta no anel
					temp.corpo[TaskId-1] = TaskId
					if actualLeader != temp.vencedor && temp.vencedor != 0{
						actualLeader = temp.vencedor
					}
					if !bFailed {
						fmt.Printf("%2d: Rodando, Líder atual: %d\n", TaskId, actualLeader)
					} else {
						fmt.Printf("%2d: Estou Falho\n", TaskId)
					}
					out <- temp
				} else {
					controle <- actualLeader
				}
			}

		case 4:
			{
				fmt.Printf("%2d: Finalizando Execução...\n", TaskId)
				return
			}
		}
	}
}

func falha_eleicao(channel int, temp mensagem) {
	// Nova falha
	temp.tipo = 2
	chans[channel] <- temp

	time.Sleep(2 * time.Second)

	fmt.Printf("[ NOVA ELEIÇÃO ]\n")

	time.Sleep(1 * time.Second)

	// Mensagem de eleicao
	temp.tipo = 1
	chans[(channel + 1) % len(chans)] <- temp
}

func main() {

	wg.Add(5) // Add a count of four, one for each goroutine

	// criar os processo do anel de eleicao

	go ElectionStage(1, chans[3], chans[0], 1) // este é o lider
	go ElectionStage(2, chans[0], chans[1], 1) // não é lider, é o processo 0
	go ElectionStage(3, chans[1], chans[2], 1) // não é lider, é o processo 0
	go ElectionStage(4, chans[2], chans[3], 1) // não é lider, é o processo 0

	fmt.Println("\n   Anel de processos criado")

	// criar o processo controlador

	go ElectionControler(controle)

	fmt.Println("\n   Processo controlador criado\n")

	wg.Wait() // Wait for the goroutines to finish\
}
