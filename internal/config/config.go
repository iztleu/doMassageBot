package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Configuration struct {
	TelegramBotToken string `json:"TelegramBotToken"`
	UpdateTimeout    int    `json:"UpdateTimeout"`
	DbConfig         Db     `json:"db"`
}
type Db struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	User      string `json:"user"`
	Password  string `json:"pass"`
	DbName    string `json:"dbname"`
	IdleConns int    `json:"idle_conns"`
	OpenConns int    `json:"open_conns"`
}

func LoadConfiguration(path string) (Configuration, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var configs Configuration
	json.Unmarshal(byteValue, &configs)
	return configs, err
}
