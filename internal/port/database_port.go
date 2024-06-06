package port

import (
	"github.com/abhilashdk2016/my-grpc-go-server/internal/adapter/database"
	"github.com/google/uuid"
)

type DummyDatabasePort interface {
	Save(data *database.DummyOrm) (uuid.UUID, error)
	GetByUuid(uuid *uuid.UUID) (database.DummyOrm, error)
}
