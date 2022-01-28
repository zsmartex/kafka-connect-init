package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/google/go-cmp/cmp"
)

type Connector struct {
	Name   string          `json:"name"`
	Config ConnectorConfig `json:"config"`
}

type ConnectorConfig map[string]string

var KafkaConnectHost string
var Client *resty.Client

func main() {
	KafkaConnectHost = os.Getenv("KAFKA_CONNECT_HOST")
	Client = resty.New()

	files, err := ioutil.ReadDir("connectors")
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		file_name := f.Name()

		buf, _ := os.ReadFile(fmt.Sprintf("connectors/%s", file_name))

		var connector Connector
		json.Unmarshal(buf, &connector)

		current_connector_config, err := GetConnectorConfig(connector.Name)
		if err != nil {
			log.Printf("Creating connector %s", connector.Name)
			if err := CreateConnector(connector); err != nil {
				log.Println(err)
			}
			continue
		}

		delete(current_connector_config, "name")

		if !cmp.Equal(current_connector_config, connector.Config) {
			log.Printf("Updating connector %s", connector.Name)
			UpdateConnector(connector.Name, connector.Config)
			continue
		}
	}
}

func GetConnectorConfig(name string) (ConnectorConfig, error) {
	res, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&ConnectorConfig{}).
		Get(fmt.Sprintf("%s/connectors/%s/config", KafkaConnectHost, name))

	if err != nil {
		return nil, err
	}

	connector_config := res.Result().(*ConnectorConfig)

	if (*connector_config)["connector.class"] == "" {
		return nil, errors.New("Connector not found")
	}

	return *connector_config, nil
}

func CreateConnector(connector Connector) error {
	res, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(connector).
		Post(fmt.Sprintf("%s/connectors", KafkaConnectHost))

	log.Println(string(res.Body()))

	return err
}

func UpdateConnector(name string, config ConnectorConfig) (Connector, error) {
	res, err := Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(config).
		SetResult(&Connector{}).
		Put(fmt.Sprintf("%s/connectors/%s/config", KafkaConnectHost, name))

	if err != nil {
		return Connector{}, err
	}

	log.Println(string(res.Body()))

	connector := res.Result().(*Connector)

	return *connector, nil
}
