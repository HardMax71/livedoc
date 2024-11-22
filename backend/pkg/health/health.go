package health

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type HealthChecker struct {
	healthpb.UnimplementedHealthServer
	db *pgxpool.Pool
}

func NewHealthChecker(db *pgxpool.Pool) *HealthChecker {
	return &HealthChecker{db: db}
}

func (h *HealthChecker) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		return &healthpb.HealthCheckResponse{
			Status: healthpb.HealthCheckResponse_NOT_SERVING,
		}, nil
	}

	return &healthpb.HealthCheckResponse{
		Status: healthpb.HealthCheckResponse_SERVING,
	}, nil
}

func (h *HealthChecker) Watch(req *healthpb.HealthCheckRequest, server healthpb.Health_WatchServer) error {
	return nil
}
