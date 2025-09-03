# Sales API - Taller Go (UNSL)

Este proyecto implementa una **API de ventas** en Go como ejercicio final del Taller de Go (UNSL).  
La API extiende un CRUD de usuarios existente y permite gestionar ventas con validaciones, estados y filtros.

---

## 🚀 Funcionalidades

- **Crear una venta** (`POST /sales`)  
  - Recibe `user_id` y `amount`.  
  - Valida existencia de usuario mediante `GET /users/:id`.  
  - Genera un `UUID` único, estado aleatorio (`pending`, `approved`, `rejected`) y timestamps.  

- **Actualizar una venta** (`PATCH /sales/:id`)  
  - Permite actualizar solo el estado si está en `pending`.  
  - Transiciones válidas: `pending → approved` o `pending → rejected`.  

- **Buscar ventas** (`GET /sales?user_id={id}&status={status}`)  
  - Devuelve todas las ventas de un usuario.  
  - Soporta filtro opcional por estado.  
  - Incluye metadatos: cantidad por estado y monto total.  

---

## 🛠️ Tecnologías utilizadas

- [Go](https://go.dev/)  
- [Gin](https://github.com/gin-gonic/gin) - Framework web  
- [Resty](https://github.com/go-resty/resty) - Cliente HTTP  
- [UUID](https://github.com/google/uuid) - Identificadores únicos  
- [Zap](https://github.com/uber-go/zap) - Logger  

---

## 📦 Instalación y ejecución

1. Clonar el repositorio:
   git clone https://github.com/usuario/sales-api-go.git
   cd sales-api-go

2. Instalar dependencias:
   go mod tidy

3. Ejecutar el servidor:
   go run main.go

El servidor correrá en:
👉 `http://localhost:8080`

---

## 🧪 Tests

El proyecto incluye **tests unitarios y de integración**:

* Unitario: creación de venta con `user_id` inexistente.
* Integración: flujo completo `POST → PATCH → GET`.

Ejecutar:

go test ./...

---

## 📖 Ejemplos de uso

### Crear una venta

curl -X POST http://localhost:8080/sales \
  -H "Content-Type: application/json" \
  -d '{"user_id": "123", "amount": 1500}'

### Actualizar estado

curl -X PATCH http://localhost:8080/sales/{id} \
  -H "Content-Type: application/json" \
  -d '{"status": "approved"}'

### Buscar ventas

curl "http://localhost:8080/sales?user_id=123&status=approved"

---

## 📌 Notas

* El almacenamiento de ventas es en memoria.
* Respuestas HTTP siguen buenas prácticas (201, 200, 400, 404, 409, 500).
