package main

import (
	"database/sql"  // me permite trabajar con bases de datos
	"encoding/json" // convierte datos a JSON
	"fmt"           // imprime mensajes
	"net/http"      // permite crear un servidor web
	"strings"
	"time" // deja trabajar con fechas

	// permite manipular cadenas de texto
	_ "github.com/go-sql-driver/mysql" // el traductor entre Go y MySQL
	"github.com/gorilla/mux"           // ayuda a manejar los endpoints
)

// Consumo representa un registro de la tabla consumos
type Consumo struct {
	MeterId      int     `json:"meter_id"`      // número del medidor
	ActiveEnergy float64 `json:"active_energy"` // energía consumida
	Date         string  `json:"date"`          // fecha del consumo
}

// RespuestaConsumo es para mostrar el formato final que se enviara
type RespuestaConsumo struct {
	Period    []string    `json:"period"`     // lista de fechas
	DataGraph []DataGraph `json:"data_graph"` // lista de medidores
}

// DataGraph tiene la informacion de cada medidor
type DataGraph struct {
	MeterId            int       `json:"meter_id"` 
	Address            string    `json:"address"`  
	Active             []float64 `json:"active"`   
	ReactiveInductive  []float64 `json:"reactive_inductive"`
	ReactiveCapacitive []float64 `json:"reactive_capacitive"`
	Exported           []float64 `json:"exported"`
}

// conexion es la conexión a la base de datos
var conexion *sql.DB

func main() {
	// Aquí se le digo a Go cómo conectarse a MySQL
	// usuario:contraseña@dirección/nombre_base_de_datos
	var err error
	conexion, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/bia-consumos")

	// Si hubo un error al conectar, se mostarrá
	if err != nil {
		fmt.Println("Error al conectar:", err)
		return
	}

	// Cerrar la conexión cuando el programa termine
	defer conexion.Close()

	fmt.Println("¡Conexión exitosa con la base de datos!")

	// Creo el enrutador, es como el recepcionista que dirige cada pregunta a la puerta correcta
	enrutador := mux.NewRouter()

	// Registra  la primera puerta (endpoint)
	enrutador.HandleFunc("/consumption", obtenerConsumos).Methods("GET")

	// Inicia el servidor en el puerto 8080
	fmt.Println("Servidor iniciado en http://localhost:8080")
	http.ListenAndServe(":8080", enrutador)
}

// formatearFecha convierte una fecha a formato "JUN 1"
func formatearFecha(fecha time.Time) string {
	// Meses en español
	meses := map[string]string{
		"Jan": "ENE", "Feb": "FEB", "Mar": "MAR",
		"Apr": "ABR", "May": "MAY", "Jun": "JUN",
		"Jul": "JUL", "Aug": "AGO", "Sep": "SEP",
		"Oct": "OCT", "Nov": "NOV", "Dec": "DIC",
	}
	return meses[fecha.Format("Jan")] + " " + fecha.Format("2")
}

// formatearMes convierte una fecha a formato "JUN 2023"
func formatearMes(fecha time.Time) string {
	meses := map[string]string{
		"Jan": "ENE", "Feb": "FEB", "Mar": "MAR",
		"Apr": "ABR", "May": "MAY", "Jun": "JUN",
		"Jul": "JUL", "Aug": "AGO", "Sep": "SEP",
		"Oct": "OCT", "Nov": "NOV", "Dec": "DIC",
	}
	return meses[fecha.Format("Jan")] + " " + fecha.Format("2006")
}

// obtenerConsumoPorRango suma los consumos entre dos fechas
func obtenerConsumoPorRango(meterId string, inicio string, fin string) float64 {
	// Consultamos la suma de consumos entre dos fechas
	fila := conexion.QueryRow(
		"SELECT COALESCE(SUM(active_energy), 0) FROM consumos WHERE meter_id = ? AND date BETWEEN ? AND ?",
		meterId, inicio, fin,
	)

	// Guardamos el resultado
	var total float64
	fila.Scan(&total)
	return total
}

// obtenerConsumos es la función que responde cuando alguien
// llama a la puerta /consumption
func obtenerConsumos(w http.ResponseWriter, r *http.Request) {
	// r es la petición que llega
	// Query() nos da todos los parámetros de la URL
	// Leemos los ids y los separamos por coma
	meterId := r.URL.Query().Get("meters_ids")
	listaIds := strings.Split(meterId, ",")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	kindPeriod := r.URL.Query().Get("kind_period")

	// Imprimo los parámetros para verificar que llegaron bien
	fmt.Println("Medidores:", listaIds)
	fmt.Println("Fecha inicio:", startDate)
	fmt.Println("Fecha fin:", endDate)
	fmt.Println("Período:", kindPeriod)

	// Lista respuesta final
	var periodos []string
	var dataGraph []DataGraph

	// Recorre cada medidor de la lista
	for _, id := range listaIds {

		// Listas de consumos para el medidor
		var activos []float64
		var consumos []Consumo
		calcularPeriodos := len(periodos) == 0

		if kindPeriod == "weekly" {
			// Convertir las fechas de texto a fechas reales
			inicio, _ := time.Parse("2006-01-02", startDate)
			fin, _ := time.Parse("2006-01-02", endDate)

			// Recorrer semana por semana
			for inicio.Before(fin) || inicio.Equal(fin) {
				// El fin de esta semana es 6 días después
				finSemana := inicio.AddDate(0, 0, 6)

				// Si el fin de semana pasa del end_date usamos end_date
				if finSemana.After(fin) {
					finSemana = fin
				}

				// Solo agrega los períodos para el primer medidor
				if calcularPeriodos {
					periodos = append(periodos, formatearFecha(inicio)+" - "+formatearFecha(finSemana))
				}

				// Consulta el consumo de esa semana
				total := obtenerConsumoPorRango(id, inicio.Format("2006-01-02"), finSemana.Format("2006-01-02"))
				activos = append(activos, total)

				// Avanzamos 7 días
				inicio = inicio.AddDate(0, 0, 7)
			}
		} else {
			// Para daily y monthly se usa consulta SQL
			var consulta string
			if kindPeriod == "daily" {
				consulta = "SELECT meter_id, SUM(active_energy) as active_energy, DATE(date) as date FROM consumos WHERE meter_id = ? AND date BETWEEN ? AND ? GROUP BY meter_id, DATE(date) ORDER BY date"
			} else if kindPeriod == "monthly" {
				consulta = "SELECT meter_id, SUM(active_energy) as active_energy, DATE(date) as date FROM consumos WHERE meter_id = ? AND date BETWEEN ? AND ? GROUP BY meter_id, DATE_FORMAT(date,'%Y-%m') ORDER BY date"
			}

			registros, err := conexion.Query(consulta, id, startDate, endDate)
			if err != nil {
				fmt.Println("Error al consultar:", err)
				return
			}
			defer registros.Close()

			// acá voy guardando cada dato que me trae la base de datos
                for registros.Next() {
				var consumo Consumo
				registros.Scan(&consumo.MeterId, &consumo.ActiveEnergy, &consumo.Date)
				consumos = append(consumos, consumo)
			}

			//Se recorre los consumos y espera las fechas de los valores
			for _, consumo := range consumos {
				fecha, _ := time.Parse("2006-01-02", consumo.Date)
				if calcularPeriodos {
					if kindPeriod == "monthly" {
						periodos = append(periodos, formatearMes(fecha))
					} else {
						periodos = append(periodos, formatearFecha(fecha))
					}
				}
				activos = append(activos, consumo.ActiveEnergy)
			}
		}

		// Convertimos el id a número
		metaId := 0
		fmt.Sscan(id, &metaId)

		// Creamos listas de ceros para los campos que no tenemos
		ceros := make([]float64, len(activos))

		// Se pone un medidor al data_graph
		dataGraph = append(dataGraph, DataGraph{
			MeterId:            metaId,
			Address:            "Dirección mock",
			Active:             activos,
			ReactiveInductive:  ceros,
			ReactiveCapacitive: ceros,
			Exported:           ceros,
		})
	}

	// Respuesta final
	respuesta := RespuestaConsumo{
		Period:    periodos,
		DataGraph: dataGraph,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}
