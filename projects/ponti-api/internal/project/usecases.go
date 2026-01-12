package project

import (
	"context"
	"fmt"
	"log"

	customer "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer"
	customerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
	field "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field"
	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	investor "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor"
	investordom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot"
	manager "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager"
	managerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

type useCases struct {
	repo     Repository
	customer customer.UseCases
	manager  manager.UseCases
	investor investor.UseCases
	field    field.UseCases
	lot      lot.UseCases
}

func NewUseCases(
	repo Repository,
	cu customer.UseCases,
	ma manager.UseCases,
	in investor.UseCases,
	fu field.UseCases,
	lo lot.UseCases,
) UseCases {
	return &useCases{
		repo:     repo,
		customer: cu,
		manager:  ma,
		investor: in,
		field:    fu,
		lot:      lo,
	}
}

func (u *useCases) CreateProject(ctx context.Context, p *domain.Project) (int64, error) {
	// Track created entity IDs for compensation
	var createdMgrs, createdInvs, createdFields []int64

	// 1) Customer
	if p.Customer.ID == 0 {
		custID, err := u.customer.CreateCustomer(ctx, &customerdom.Customer{Name: p.Customer.Name})
		if err != nil {
			return 0, fmt.Errorf("create customer: %w", err)
		}
		p.Customer.ID = custID
	}

	// 2) Managers
	for i := range p.Managers {
		m := &p.Managers[i]
		if m.ID == 0 {
			id, err := u.manager.CreateManager(ctx, &managerdom.Manager{Name: m.Name})
			if err != nil {
				// rollback managers
				for _, mgrID := range createdMgrs {
					if delErr := u.manager.DeleteManager(ctx, mgrID); delErr != nil {
						log.Printf("rollback manager %d failed: %v", mgrID, delErr)
					}
				}
				// rollback customer
				_ = u.customer.DeleteCustomer(ctx, p.Customer.ID)
				return 0, fmt.Errorf("create manager %q: %w", m.Name, err)
			}
			createdMgrs = append(createdMgrs, id)
			m.ID = id
		}
	}

	// 3) Investors
	for i := range p.Investors {
		inv := &p.Investors[i]
		if inv.ID == 0 {
			id, err := u.investor.CreateInvestor(ctx, &investordom.Investor{Name: inv.Name, Percentage: inv.Percentage})
			if err != nil {
				// rollback investors
				for _, invID := range createdInvs {
					_ = u.investor.DeleteInvestor(ctx, invID)
				}
				// rollback managers
				for _, mgrID := range createdMgrs {
					_ = u.manager.DeleteManager(ctx, mgrID)
				}
				// rollback customer
				_ = u.customer.DeleteCustomer(ctx, p.Customer.ID)
				return 0, fmt.Errorf("create investor %q: %w", inv.Name, err)
			}
			createdInvs = append(createdInvs, id)
			inv.ID = id
		}
	}

	// 4) Fields (CreateField handles nested lots)
	for i := range p.Fields {
		f := &p.Fields[i]
		fid, err := u.field.CreateField(ctx, f)
		if err != nil {
			// rollback fields
			for _, fldID := range createdFields {
				_ = u.field.DeleteField(ctx, fldID)
			}
			// rollback investors
			for _, invID := range createdInvs {
				_ = u.investor.DeleteInvestor(ctx, invID)
			}
			// rollback managers
			for _, mgrID := range createdMgrs {
				_ = u.manager.DeleteManager(ctx, mgrID)
			}
			// rollback customer
			_ = u.customer.DeleteCustomer(ctx, p.Customer.ID)
			return 0, fmt.Errorf("create field %q: %w", f.Name, err)
		}
		createdFields = append(createdFields, fid)
		f.ID = fid
	}

	// 5) Persist project and pivot tables (handled transactionally in repository)
	projID, err := u.repo.CreateProject(ctx, p)
	if err != nil {
		// full rollback: project, fields, investors, managers, customer
		_ = u.repo.DeleteProject(ctx, projID)
		for _, fldID := range createdFields {
			_ = u.field.DeleteField(ctx, fldID)
		}
		for _, invID := range createdInvs {
			_ = u.investor.DeleteInvestor(ctx, invID)
		}
		for _, mgrID := range createdMgrs {
			_ = u.manager.DeleteManager(ctx, mgrID)
		}
		_ = u.customer.DeleteCustomer(ctx, p.Customer.ID)
		return 0, fmt.Errorf("create project: %w", err)
	}

	return projID, nil
}

func (u *useCases) GetProject(ctx context.Context, id int64) (*domain.Project, error) {
	proj, err := u.repo.GetProject(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := u.enrichProject(ctx, proj); err != nil {
		return nil, err
	}
	return proj, nil
}

func (u *useCases) ListProjects(ctx context.Context) ([]domain.Project, error) {
	list, err := u.repo.ListProjects(ctx)
	if err != nil {
		return nil, err
	}
	for i := range list {
		if err := u.enrichProject(ctx, &list[i]); err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (u *useCases) ListProjectsByCustomerID(ctx context.Context, customerID int64) ([]domain.Project, error) {
	list, err := u.repo.ListProjectsByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}
	for i := range list {
		if err := u.enrichProject(ctx, &list[i]); err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (u *useCases) UpdateProject(ctx context.Context, p *domain.Project) error {
	return u.repo.UpdateProject(ctx, p)
}

func (u *useCases) DeleteProject(ctx context.Context, id int64) error {
	return u.repo.DeleteProject(ctx, id)
}

// helpers
func (u *useCases) enrichProject(ctx context.Context, p *domain.Project) error {
	// Customer
	cust, err := u.customer.GetCustomer(ctx, p.Customer.ID)
	if err != nil {
		return fmt.Errorf("fetch customer %d: %w", p.Customer.ID, err)
	}
	p.Customer = *cust

	// Managers
	var mgrs []managerdom.Manager
	for _, m := range p.Managers {
		man, err := u.manager.GetManager(ctx, m.ID)
		if err != nil {
			return fmt.Errorf("fetch manager %d: %w", m.ID, err)
		}
		mgrs = append(mgrs, *man)
	}
	p.Managers = mgrs

	// Investors
	var invs []investordom.Investor
	for _, inv := range p.Investors {
		i, err := u.investor.GetInvestor(ctx, inv.ID)
		if err != nil {
			return fmt.Errorf("fetch investor %d: %w", inv.ID, err)
		}
		invs = append(invs, *i)
	}
	p.Investors = invs

	// Fields (incluye nested Lots)
	var flds []fielddom.Field
	for _, f := range p.Fields {
		fld, err := u.field.GetField(ctx, f.ID)
		if err != nil {
			return fmt.Errorf("fetch field %d: %w", f.ID, err)
		}
		flds = append(flds, *fld)
	}
	p.Fields = flds

	return nil
}
