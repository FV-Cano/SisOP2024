package cfg

import (
	"encoding/json"
	"os"
)

/*	ConfigInit:

	Hay que pasarle una interfaz como parámetro para que el decoder pueda
	decodificar el json con el formato de la interfaz especificada.
	Si no lo haces, decodifica con un map[string]interface{}.
*/
func ConfigInit(filePath string, config interface{}) error {
	configFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&config); err != nil {
		return err
	}

	return nil
}
