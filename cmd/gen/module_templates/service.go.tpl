package {{.ModuleName}}_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/{{.ModuleName}}_module/dtos"
)

type (
	IService interface {
		Get{{.PascalModulePluralName}}(ctx context.Context, paginate *paginator.Pagination) (*paginator.Pagination, error)
		Get{{.PascalModuleName}}(ctx context.Context, id int) (map[string]interface{}, error)
		Create{{.PascalModuleName}}(ctx context.Context, dto *dtos.Create{{.PascalModuleName}}) error
		Update{{.PascalModuleName}}(ctx context.Context, id int, dto *dtos.Update{{.PascalModuleName}}) error
		Delete{{.PascalModuleName}}(ctx context.Context, id int) error
	}
)

func (s Service) Get{{.PascalModulePluralName}}(ctx context.Context, paginate *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "Get{{.PascalModulePluralName}}Service", trace.WithAttributes(attribute.String("service", "Get{{.PascalModulePluralName}}")))

	result, err := s.{{.CamelModuleName}}Repository().Get{{.PascalModuleName}}Paginate(ctx, paginate)

	s.tracer.TraceEnd(childSpan)

	return result, err
}

func (s Service) Get{{.PascalModuleName}}(ctx context.Context, id int) (map[string]interface{}, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "Get{{.PascalModuleName}}Service", trace.WithAttributes(attribute.String("service", "Get{{.PascalModuleName}}")))

	{{.CamelModuleName}}, err := s.{{.CamelModuleName}}Repository().Get{{.PascalModuleName}}ByID(ctx, id)

	s.tracer.TraceEnd(childSpan)

	return map[string]interface{}{
		"data": {{.CamelModuleName}},
	}, err
}

func (s Service) Create{{.PascalModuleName}}(ctx context.Context, dto *dtos.Create{{.PascalModuleName}}) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "Create{{.PascalModuleName}}Service", trace.WithAttributes(attribute.String("service", "Create{{.PascalModuleName}}")))

	{{.CamelModuleName}} := new(models.{{.PascalModuleName}})
  // ...

	err := s.{{.CamelModuleName}}Repository().Create{{.PascalModuleName}}(ctx, {{.CamelModuleName}})

	s.tracer.TraceEnd(childSpan)

	return err
}

func (s Service) Update{{.PascalModuleName}}(ctx context.Context, id int, dto *dtos.Update{{.PascalModuleName}}) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "Update{{.PascalModuleName}}Service", trace.WithAttributes(attribute.String("service", "Update{{.PascalModuleName}}")))

	{{.CamelModuleName}} := new(models.{{.PascalModuleName}})
  // ...

	err := s.{{.CamelModuleName}}Repository().Update{{.PascalModuleName}}(ctx, id, {{.CamelModuleName}})

	s.tracer.TraceEnd(childSpan)

	return err
}

func (s Service) Delete{{.PascalModuleName}}(ctx context.Context, id int) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "Delete{{.PascalModuleName}}Service", trace.WithAttributes(attribute.String("service", "Delete{{.PascalModuleName}}")))

	err := s.{{.CamelModuleName}}Repository().Delete{{.PascalModuleName}}(ctx, id)

	s.tracer.TraceEnd(childSpan)

	return err
}
