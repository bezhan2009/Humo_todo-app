package main

import (
	todoapp "Humo_todo-app"
	"Humo_todo-app/db"
	"Humo_todo-app/pkg/handler"
	"Humo_todo-app/pkg/repository"
	"Humo_todo-app/pkg/service"
	"context"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func Start() {
	/***************** Инициализация config-ов *****************/
	if err := initConfigs(); err != nil {
		log.Fatalf("Error while initializing configs. Error is: %s", err.Error())
	}
	/**********************************************************/

	/***************** Инициализация базы данных *****************/
	database, err := repository.NewSqliteDB(viper.GetString("db.dbname"))
	if err != nil {
		log.Fatalf("Error while opening DB. Error is: %s", err.Error())
	}
	db.Init(database)
	/**********************************************************/

	/***************** Внедрение зависимостей *****************/
	repos := repository.NewRepository(database)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)
	/**********************************************************/

	/***************** Starting App *****************/
	MainServer := new(todoapp.Server)
	go func() {
		if err := MainServer.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
			log.Fatalf("Error while running http server. Error is %s", err.Error())
		}
	}()
	fmt.Println("TodoApp Started its work")
	fmt.Printf("Server is listening port: %s\n", viper.GetString("port"))
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	/**********************************************************/

	/***************** Shutting App Down *****************/
	fmt.Println("TodoApp Shutting Down")
	if err := MainServer.Shutdown(context.Background()); err != nil {
		log.Fatalf("error while shutting server down. Error is: %s", err.Error())
	}
	if err := database.Close(); err != nil {
		log.Fatalf("error while closing DB. Error is: %s", err.Error())
	}
	/**********************************************************/
}

func main() {
	Start()
}

// initConfigs Функция инициализации config-ов
func initConfigs() error {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs" // default relative path
	}
	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
