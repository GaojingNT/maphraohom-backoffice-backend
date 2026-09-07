# Generator

## Module (CRUD)

The framework with generate new module directory. then inside will have route, dtos, controller, service and repository files.

### Module folder structure

```text
📁 src/modules
├── 📁 team_module
│   ├── 📁 dtos
│   │   ├── 📄 dtos.go
│   ├── 📄 team_controller.go
│   ├── 📄 team_module.go
│   ├── 📄 team_repository.go
│   ├── 📄 team_service.go
```

### Generate module command

```bash
go run cmd/generate.go module <module_name>

# Example
go run cmd/generate.go module team
```

### Register module routes

Don't forgot to register module route at `src/routes/http.go` at the `HTTPRoutes` method.

```go
func HTTPRoutes(s *http_server.HttpServer) {
  ...
  // Register routes
  ...
  team_module.NewModule().Routes(api, middleware)
}
```

---

## Model (GORM - declaring models)

The framework with generate new model file.

### Model folder structure

```text
📁 src
├── 📄 models
│   ├── 📄 team_model.go
```

### Generate model command

```bash
go run cmd/generate.go model <model_name>

# Example
go run cmd/generate.go model team
```

### Sample model file

```go
package models

type Team struct {
 BaseModel
}
```
