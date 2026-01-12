package domain

import (
	customerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
	fieldom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	investordom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	managerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
)

type Project struct {
	ID        int64                  // primary key
	Name      string                 // project name
	Customer  customerdom.Customer   // loaded client
	Managers  []managerdom.Manager   // many-to-many relation
	Investors []investordom.Investor // pivot relation with extra field
	Fields    []fieldom.Field        // child fields
}
