package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// Consumo representa un registro de la tabla consumos
type Consumo struct {
	MeterId      int     `json:"meter_id"`
	ActiveEnergy float64 `json:"active_energy"`
	Date         string  `json:"date"`
}

// RespuestaConsumo es el formato final que se enviará
type RespuestaConsumo struct {
	Period    []string    `json:"period"`
	DataGraph []DataGraph `json:"data_graph"`
}

// DataGraph tiene la información de cada medidor
type DataGraph struct {
	MeterId            int       `json:"meter_id"`
	Address            string    `json:"address"`
	Active             []float64 `json:"active"`
	ReactiveInductive  []float64 `json:"reactive_inductive"`
	ReactiveCapacitive []float64 `json:"reactive_capacitive"`
	Exported           []float64 `json:"exported"`
}

var conexion *sql.DB

func main() {
	var err error
	conexion, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/bia-consumos")
	if err != nil {
		fmt.Println("Error al conectar:", err)
		return
	}
	defer conexion.Close()

	fmt.Println("¡Conexión exitosa con la base de datos!")

	enrutador := mux.NewRouter()
	enrutador.HandleFunc("/consumption", obtenerConsumos).Methods("GET")

	fmt.Println("Servidor iniciado en http://localhost:8080")
	http.ListenAndServe(":8080", enrutador)
}

func formatearFecha(fecha time.Time) string {
	meses := map[string]string{
		"Jan": "ENE", "Feb": "FEB", "Mar": "MAR",
		"Apr": "ABR", "May": "MAY", "Jun": "JUN",
		"Jul": "JUL", "Aug": "AGO", "Sep": "SEP",
		"Oct": "OCT", "Nov": "NOV", "Dec": "DIC",
	}
	return meses[fecha.Format("Jan")] + " " + fecha.Format("2")
}

func formatearMes(fecha time.Time) string {
	meses := map[string]string{
		"Jan": "ENE", "Feb": "FEB", "Mar": "MAR",
		"Apr": "ABR", "May": "MAY", "Jun": "JUN",
		"Jul": "JUL", "Aug": "AGO", "Sep": "SEP",
		"Oct": "OCT", "Nov": "NOV", "Dec": "DIC",
	}
	return meses[fecha.Format("Jan")] + " " + fecha.Format("2006")
}

func obtenerConsumoPorRango(meterId string, inicio string, fin string) float64 {
	fila := conexion.QueryRow(
		"SELECT COALESCE(SUM(active_energy), 0) FROM consumos WHERE meter_id = ? AND date BETWEEN ? AND ?",
		meterId, inicio, fin,
	)
	var total float64
	fila.Scan(&total)
	return total
}

func consultarDireccion(meterId string) string {
	direcciones := map[string]string{
		"1": "Calle 13 # 41-50",
		"2": "Carrera 64 # 2-02",
		"3": "Carrera 14 # 5-05",
	}
	address, existe := direcciones[meterId]
	if !existe {
		return "Dirección no encontrada"
	}
	return address
}

func obtenerConsumos(w http.ResponseWriter, r *http.Request) {
	meterId := r.URL.Query().Get("meters_ids")
	listaIds := strings.Split(meterId, ",")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	kindPeriod := r.URL.Query().Get("kind_period")

	fmt.Println("Medidores:", listaIds)
	fmt.Println("Fecha inicio:", startDate)
	fmt.Println("Fecha fin:", endDate)
	fmt.Println("Período:", kindPeriod)

	var periodos []string
	var dataGraph []DataGraph

	for _, id := range listaIds {
		var activos []float64
		var consumos []Consumo
		calcularPeriodos := len(periodos) == 0

		if kindPeriod == "weekly" {
			inicio, _ := time.Parse("2006-01-02", startDate)
			fin, _ := time.Parse("2006-01-02", endDate)

			for inicio.Before(fin) || inicio.Equal(fin) {
				finSemana := inicio.AddDate(0, 0, 6)
				if finSemana.After(fin) {
					finSemana = fin
				}
				if calcularPeriodos {
					periodos = append(periodos, formatearFecha(inicio)+" - "+formatearFecha(finSemana))
				}
				total := obtenerConsumoPorRango(id, inicio.Format("2006-01-02"), finSemana.Format("2006-01-02"))
				activos = append(activos, total)
				inicio = inicio.AddDate(0, 0, 7)
			}
		} else {
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

			for registros.Next() {
				var consumo Consumo
				registros.Scan(&consumo.MeterId, &consumo.ActiveEnergy, &consumo.Date)
				consumos = append(consumos, consumo)
			}
			registros.Close()

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

		metaId := 0
		fmt.Sscan(id, &metaId)
		ceros := make([]float64, len(activos))

		dataGraph = append(dataGraph, DataGraph{
			MeterId:            metaId,
			Address:            consultarDireccion(id),
			Active:             activos,
			ReactiveInductive:  ceros,
			ReactiveCapacitive: ceros,
			Exported:           ceros,
		})
	}

	respuesta := RespuestaConsumo{
		Period:    periodos,
		DataGraph: dataGraph,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}
