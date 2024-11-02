// Código exemplo para o trabaho de sistemas distribuidos (eleicao em anel)
// By Cesar De Rose - 2022

package main

import (
	"fmt"
	"sync"
	"time"
)

type mensagem struct {
	tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmacao da eleicao)
	corpo [5]int // conteudo da mensagem para colocar os ids (usar um tamanho ocmpativel com o numero de processos no anel)
	vencedor int // Campo criado para passar apenas o vencedor da eleicao
}

var (
	chans = []chan mensagem{ // vetor de canais para formar o anel de eleicao - chan[0], chan[1] and chan[2] ...
		make(chan mensagem),
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
	// Variaveis para controlar o reativar e desativar para teste
	var active_4 bool = true
	var active_5 bool = true

	// comandos para o anel iciam aqui
	temp.tipo = 3 // Mensagem para iniciar os processos
	chans[4] <- temp

	for {
		control := <-in 
		if(control == 5) {
			if active_5 {
				falha_eleicao(3, temp)
				active_5 = false
			} else {
				control = 1
			}

		}
		if(control == 4) {
			if active_4 {
				falha_eleicao(2, temp)
				active_4 = false
			} else {
				reativa_eleicao(3, temp)
			}
		}
		if(control == 3) {
			reativa_eleicao(2, temp)
		}
		if(control == 1) {
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
						fmt.Printf("%2d: ELEIÇÃO: [ %d, %d, %d, %d, %d]\n", TaskId, temp.corpo[0], temp.corpo[1], temp.corpo[2], temp.corpo[3], temp.corpo[4])
						out <- temp
					} else { // Se deu a volta no anel
						// Vencedor é o maior id
						maior := temp.corpo[0]
						for _, valor := range temp.corpo {
							if valor > maior {
								maior = valor
							}
						}
						temp.vencedor = maior
						temp.tipo = 3
						temp.corpo[0] = 0
						temp.corpo[1] = 0
						temp.corpo[2] = 0
						temp.corpo[3] = 0
						temp.corpo[4] = 0
						fmt.Printf("%2d: VENCEDOR DA ELEIÇÃO - [%2d ]\n", TaskId, temp.vencedor)
						// Volta a rodar
						out <- temp
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
					// Atualiza lider
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
		// Termina execução do processo
		case 4:
			{
				fmt.Printf("%2d: Finalizando Execução...\n", TaskId)
				return
			}
		// Reativa
		case 5:
			{
				time.Sleep(1 * time.Second)
				bFailed = false
				fmt.Printf("%2d: Fui Reativado\n", TaskId)
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

func reativa_eleicao(channel int, temp mensagem) {
	// Reativa
	temp.tipo = 5
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
	go ElectionStage(1, chans[4], chans[0], 5) // não é lider, é o processo 5
	go ElectionStage(2, chans[0], chans[1], 5) // não é lider, é o processo 5
	go ElectionStage(3, chans[1], chans[2], 5) // não é lider, é o processo 5
	go ElectionStage(4, chans[2], chans[3], 5) // não é lider, é o processo 5
	go ElectionStage(5, chans[3], chans[4], 5) // este é o lider

	fmt.Println("\n   Anel de processos criado")

	// criar o processo controlador
	go ElectionControler(controle)

	fmt.Println("\n   Processo controlador criado\n")

	wg.Wait() // Wait for the goroutines to finish\
}
