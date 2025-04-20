package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type SystemMetrics struct {
	RAMUsage  float64 `json:"ram_usage"`
	CPUUsage  float64 `json:"cpu_usage"`
	Timestamp int64   `json:"timestamp"`
}

func getCPULoad() (float64, error) {
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}
	if len(percentages) > 0 {
		return percentages[0], nil
	}
	return 0, fmt.Errorf("no CPU data available")
}

func monitorSystem(ctx context.Context, conn *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Получаем метрики RAM
			vmStat, _ := mem.VirtualMemory()
			ramPercent := vmStat.UsedPercent

			// Получаем метрики CPU
			cpuPercent, err := getCPULoad()
			if err != nil {
				fmt.Println("CPU monitoring error:", err)
				cpuPercent = 0
			}

			// Создаем структуру с метриками
			metrics := SystemMetrics{
				RAMUsage:  ramPercent,
				CPUUsage:  cpuPercent,
				Timestamp: time.Now().Unix(),
			}

			// Отправляем данные
			jsonData, _ := json.Marshal(metrics)
			err = conn.WriteMessage(websocket.TextMessage, jsonData)
			if err != nil {
				fmt.Println("Write error:", err)
				return
			}

			time.Sleep(1 * time.Second)
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	monitorSystem(ctx, conn)
}

func servePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	server := &http.Server{
		Addr: ":8080",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(w, r, ctx)
	})
	http.HandleFunc("/", servePage)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Println("Сервер запущен: http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Ошибка сервера: %v\n", err)
		}
	}()

	<-stop
	fmt.Println("\nПолучен сигнал завершения...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Ошибка при завершении сервера: %v\n", err)
	}

	cancel()
	fmt.Println("Сервер успешно остановлен")
}
