# Microservicio de Consumos de Energía - BIA Proyecto Entrevista

Este microservicio permite consultar los consumos de energía de los medidores de los clientes, agrupados por día, semana o mes.

## Tecnologías utilizadas
- Golang
- MySQL
- XAMPP

## Requisitos previos
- Tener instalado Go 1.21 o superior
- Tener instalado XAMPP con MySQL corriendo en el puerto 3306

## Cómo instalar y correr el proyecto

### 1. Clonar el repositorio
git clone https://github.com/laurasofia-18/bia-consumos.git

### 2. Entrar a la carpeta
cd bia-consumos

### 3. Instalar dependencias
go mod tidy

### 4. Configurar la base de datos
- Abrir XAMPP y iniciar MySQL
- Abrir phpMyAdmin en http://localhost/phpmyadmin
- Crear una base de datos llamada bia-consumos
- Importar el archivo test_bia.csv en la tabla consumos

### 5. Correr el proyecto
go run main.go

El servidor quedará corriendo en http://localhost:8080

## Endpoints disponibles

### Consumo diario
GET http://localhost:8080/consumption?meters_ids=1&start_date=2023-06-01&end_date=2023-06-10&kind_period=daily

Respuesta esperada:
{
  "period": ["JUN 1", "JUN 2", "JUN 3"],
  "data_graph": [
    {
      "meter_id": 1,
      "address": "Dirección mock",
      "active": [139088.88, 139560.04, 140171.94],
      "reactive_inductive": [0, 0, 0],
      "reactive_capacitive": [0, 0, 0],
      "exported": [0, 0, 0]
    }
  ]
}

### Consumo semanal
GET http://localhost:8080/consumption?meters_ids=1&start_date=2023-06-01&end_date=2023-06-26&kind_period=weekly

Respuesta:
{
  "period": ["JUN 1 - JUN 7", "JUN 8 - JUN 14", "JUN 15 - JUN 21", "JUN 22 - JUN 26"],
  "data_graph": [
    {
      "meter_id": 1,
      "address": "Dirección mock",
      "active": [843164.27, 864838.94, 885274.53, 603936.45],
      "reactive_inductive": [0, 0, 0, 0],
      "reactive_capacitive": [0, 0, 0, 0],
      "exported": [0, 0, 0, 0]
    }
  ]
}

### Consumo mensual
GET http://localhost:8080/consumption?meters_ids=1&start_date=2023-06-01&end_date=2023-07-10&kind_period=monthly

Respuesta:
{
  "period": ["JUN 2023", "JUL 2023"],
  "data_graph": [
    {
      "meter_id": 1,
      "address": "Dirección mock",
      "active": [4401580.21, 789036.97],
      "reactive_inductive": [0, 0],
      "reactive_capacitive": [0, 0],
      "exported": [0, 0]
    }
  ]
}

### Múltiples medidores
GET http://localhost:8080/consumption?meters_ids=1,2&start_date=2023-06-01&end_date=2023-06-10&kind_period=daily

## Correr los tests
go test

## Autor
Laura Sofia Aguirre Urrego
Aprendiz SENA