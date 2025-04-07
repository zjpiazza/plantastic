Sure! Here's a detailed `README.md` file for your Plantastic application, based on all the context we've discussed so far. This file provides an overview of the application, how to set it up, and how to use it.

---

![Logo](./logo.png)

# 🌱 Plantastic API

Plantastic is a RESTful API designed to help users manage their gardens, garden beds, plants, and tasks. It provides endpoints for creating, updating, and managing gardens, beds, and tasks, as well as tracking environmental data. Built with using the Gin webframework for Golang and GORM as the ORM.

---

## Features

- **Garden Management**: Create, update, and delete gardens.
- **Garden Bed Management**: Manage garden beds within a specific garden.
- **Task Management**: Create and manage tasks for a garden or specific garden beds.
- **Environmental Data**: Track environmental data like temperature, humidity, and soil moisture (optional).
- **Authentication Middleware**: Protect your API endpoints with an example authentication middleware.
- **Database Integration**: Uses GORM for database interactions with support for PostgreSQL (or other databases).

---

## Project Structure

```
plantastic/
├── cmd/
│   └── api/
│       └── main.go          # Entry point for the application
├── internal/
│   ├── handlers/            # HTTP handlers for each resource
│   │   ├── garden_handlers.go
│   │   ├── bed_handlers.go
│   │   └── task_handlers.go
│   ├── models/              # Data models for the application
│   │   ├── garden.go
│   │   ├── bed.go
│   │   └── task.go
│   ├── routes/              # API routes
│   │   └── routes.go
│   ├── middleware/          # Middleware (e.g., authentication)
│   │   └── auth.go
│   └── storage/             # Database interaction layer
│       ├── garden_storage.go
│       ├── bed_storage.go
│       └── task_storage.go
├── go.mod                   # Go module file
├── go.sum                   # Dependency lock file
└── README.md                # Project documentation
```

---

## Requirements

- **Go**: Version 1.18 or higher
- **PostgreSQL**: (or another database supported by GORM)
- **Dependencies**:
  - [Gin](https://github.com/gin-gonic/gin) for web framework
  - [GORM](https://gorm.io/) for ORM
  - [GORM PostgreSQL Driver](https://gorm.io/docs/connecting_to_the_database.html#PostgreSQL)

---

## Installation

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/yourusername/plantastic.git
   cd plantastic
   ```

2. **Initialize Go Modules**:

   ```bash
   go mod tidy
   ```

3. **Install Dependencies**:

   ```bash
   go get github.com/gorilla/mux
   go get gorm.io/gorm
   go get gorm.io/driver/postgres
   ```

4. **Set Up the Database**:

   Create a PostgreSQL database and set the `DATABASE_URL` environment variable:

   ```bash
   export DATABASE_URL="host=localhost user=postgres password=postgres dbname=plantastic port=5432 sslmode=disable"
   ```

   Replace the placeholders with your actual database credentials.

5. **Run the Application**:

   For development:

   ```bash
   go run ./cmd/api
   ```

   To build and run:

   ```bash
   go build ./cmd/api
   ./api
   ```

---

## API Endpoints

### Gardens

- **GET /gardens**: List all gardens.
- **POST /gardens**: Create a new garden.
- **GET /gardens/{garden_id}**: Retrieve a specific garden.
- **PUT /gardens/{garden_id}**: Update a garden.
- **DELETE /gardens/{garden_id}**: Delete a garden.

### Garden Beds

- **GET /gardens/{garden_id}/beds**: List all beds in a specific garden.
- **POST /gardens/{garden_id}/beds**: Create a new bed in a garden.
- **GET /beds/{bed_id}**: Retrieve a specific bed.
- **PUT /beds/{bed_id}**: Update a bed.
- **DELETE /beds/{bed_id}**: Delete a bed.

### Tasks

- **GET /gardens/{garden_id}/tasks**: List all tasks for a garden.
- **POST /gardens/{garden_id}/tasks**: Create a new task for a garden.
- **GET /tasks/{task_id}**: Retrieve a specific task.
- **PUT /tasks/{task_id}**: Update a task.
- **DELETE /tasks/{task_id}**: Delete a task.

---

## Example Request and Response

### Create a Garden

**Request**:

```http
POST /gardens
Content-Type: application/json

{
  "name": "My Backyard Garden",
  "location": "45.523, -122.676",
  "description": "A small vegetable and herb garden"
}
```

**Response**:

```http
HTTP/1.1 201 Created
Content-Type: application/json

{
  "id": "g123",
  "name": "My Backyard Garden",
  "location": "45.523, -122.676",
  "description": "A small vegetable and herb garden",
  "created_at": "2024-01-26T10:00:00Z",
  "updated_at": "2024-01-26T10:00:00Z"
}
```

---

## Database

The application uses GORM for database interactions. The following models are defined:

### Garden

```go
type Garden struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Beds        []Bed     `gorm:"foreignKey:GardenID"`
}
```

### Bed

```go
type Bed struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	GardenID  string    `gorm:"foreignKey:GardenID" json:"garden_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Size      string    `json:"size"`
	SoilType  string    `json:"soil_type"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Tasks     []Task    `gorm:"foreignKey:BedID"`
}
```

### Task

```go
type Task struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	GardenID    string    `json:"garden_id"`
	BedID       *string   `json:"garden_bed_id"`
	Description string    `json:"description"`
	DueDate     string    `json:"due_date"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
```

---

## Testing

TODO

---

## Future Enhancements

- [ ] Add support for user accounts and authentication (e.g., JWT).
- [ ] Implement environmental data tracking (e.g., temperature, humidity).
- [ ] Add support for recurring tasks.

---

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.

---

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

---