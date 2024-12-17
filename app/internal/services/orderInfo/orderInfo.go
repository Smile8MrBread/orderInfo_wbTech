// Service layout
package orderInfo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"orderInfo/app/internal/models"
	"orderInfo/app/internal/storage"
)

type App struct {
	log          *slog.Logger
	addOrder     OrderAdder
	showerData   DataShower
	cashReturner ReturnerCash
}

type OrderAdder interface {
	AddOrder(ctx context.Context, order models.Order) error
}

type DataShower interface {
	ShowData(ctx context.Context, uid string) (models.Order, error)
}

type ReturnerCash interface {
	CashReturner(ctx context.Context) ([]models.Order, error)
}

func New(log *slog.Logger, addOrder OrderAdder, shower DataShower, casher ReturnerCash) *App {
	return &App{log: log, addOrder: addOrder, showerData: shower, cashReturner: casher}
}

func (a *App) AddOrder(ctx context.Context, order models.Order) error {
	const op = "services.orderInfo.AddOrder"
	log := a.log.With(slog.String("op", op))
	log.Info("Add order")

	err := a.addOrder.AddOrder(ctx, order)
	if err != nil {
		log.Error("Failed to add order", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) ShowData(ctx context.Context, uid string) (models.Order, error) {
	const op = "services.orderInfo.ShowData"
	log := a.log.With(slog.String("op", op))
	log.Info("Show data")

	order, err := a.showerData.ShowData(ctx, uid)
	if err != nil {
		if errors.Is(err, storage.ErrUidNotFound) {
			log.Error("Uid not found", slog.String("error", err.Error()))
			return models.Order{}, fmt.Errorf("%s: %w", op, err)
		}

		log.Error("Failed to show data", slog.String("error", err.Error()))
		return models.Order{}, fmt.Errorf("%s: %w", op, err)
	}

	return order, nil
}

func (a *App) CashReturner(ctx context.Context) ([]models.Order, error) {
	const op = "services.orderInfo.CashReturner"
	log := a.log.With(slog.String("op", op))
	log.Info("Return cash")

	orders, err := a.cashReturner.CashReturner(ctx)
	if err != nil {
		log.Error("Failed to return cash", slog.String("error", err.Error()))
		return nil, nil
	}

	return orders, nil
}
