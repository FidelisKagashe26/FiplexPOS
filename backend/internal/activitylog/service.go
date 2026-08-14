package activitylog

import (
	"POS-fiplex/internal/activitylog/repository"
	"POS-fiplex/internal/common"

	"POS-fiplex/pkg/logger"
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type IActivityService interface {
	Log(ctx context.Context, userID uuid.UUID, action repository.LogActionType, entityType repository.LogEntityType, entityID string, details map[string]interface{})
	GetActivityLogs(ctx context.Context, req GetActivityLogsRequest) (*ActivityLogListResponse, error)
}

type ActivityService struct {
	repo repository.Querier
	log  logger.ILogger
}

func NewActivityService(repo repository.Querier, log logger.ILogger) IActivityService {
	return &ActivityService{
		repo: repo,
		log:  log,
	}
}

func (s *ActivityService) Log(ctx context.Context, userID uuid.UUID, action repository.LogActionType, entityType repository.LogEntityType, entityID string, details map[string]interface{}) {
	var detailsJSON []byte
	var err error

	if details != nil {
		detailsJSON, err = json.Marshal(details)
		if err != nil {
			s.log.Errorf("Log | Failed to marshal activity log details: %v", err)
		}
	}

	// Capture the shop before the goroutine — the request context may be recycled by then.
	shopParam := common.ShopParamFromContext(ctx)

	go func() {
		_, err := s.repo.CreateActivityLog(ctx, repository.CreateActivityLogParams{
			UserID:     pgtype.UUID{Bytes: userID, Valid: true},
			ActionType: repository.LogActionType(action),
			EntityType: repository.LogEntityType(entityType),
			EntityID:   entityID,
			Details:    detailsJSON,
			ShopID:     shopParam,
		})
		if err != nil {
			s.log.Errorf("Log | Failed to create activity log: %v", err)
		}
	}()
}

func (s *ActivityService) GetActivityLogs(ctx context.Context, req GetActivityLogsRequest) (*ActivityLogListResponse, error) {
	page := 1
	if req.Page != nil {
		page = *req.Page
	}
	limit := 10
	if req.Limit != nil {
		limit = *req.Limit
	}

	shopParam := common.ShopParamFromContext(ctx)

	arg := repository.GetActivityLogsParams{
		Limit:  int32(limit),
		Offset: int32((page - 1) * limit),
		ShopID: shopParam,
	}

	countArg := repository.CountActivityLogsParams{ShopID: shopParam}

	if req.ActionType != nil {
		actionType := repository.LogActionType(*req.ActionType)
		arg.ActionType = &actionType
		countArg.ActionType = &actionType
	}

	if req.EntityType != nil {
		entityType := repository.LogEntityType(*req.EntityType)
		arg.EntityType = &entityType
		countArg.EntityType = &entityType
	}

	if req.UserID != nil {
		uid, err := uuid.Parse(*req.UserID)
		if err == nil {
			arg.UserID = pgtype.UUID{Bytes: uid, Valid: true}
			countArg.UserID = pgtype.UUID{Bytes: uid, Valid: true}
		}
	}

	if req.StartDate != nil {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err == nil {
			arg.StartDate = pgtype.Timestamptz{Time: t, Valid: true}
			countArg.StartDate = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}

	if req.EndDate != nil {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err == nil {
			t = t.Add(24 * time.Hour).Add(-1 * time.Second)
			arg.EndDate = pgtype.Timestamptz{Time: t, Valid: true}
			countArg.EndDate = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}

	logs, err := s.repo.GetActivityLogs(ctx, arg)
	if err != nil {
		s.log.Errorf("GetActivityLogs | Failed to fetch logs: %v", err)
		return nil, err
	}

	totalItems, err := s.repo.CountActivityLogs(ctx, countArg)
	if err != nil {
		s.log.Errorf("GetActivityLogs | Failed to count logs: %v", err)
		return nil, err
	}

	response := &ActivityLogListResponse{
		Logs:       make([]ActivityLogResponse, 0),
		TotalItems: totalItems,
		Page:       page,
		Limit:      limit,
		TotalPages: int(math.Ceil(float64(totalItems) / float64(limit))),
	}

	for _, log := range logs {
		var details map[string]interface{}
		if len(log.Details) > 0 {
			_ = json.Unmarshal(log.Details, &details)
		}

		var userName string
		if log.UserName != nil {
			userName = *log.UserName
		}

		response.Logs = append(response.Logs, ActivityLogResponse{
			ID:         log.ID,
			UserID:     log.UserID.Bytes,
			UserName:   userName,
			ActionType: repository.LogActionType(log.ActionType),
			EntityType: repository.LogEntityType(log.EntityType),
			EntityID:   log.EntityID,
			Details:    details,
			CreatedAt:  log.CreatedAt.Time,
		})
	}

	return response, nil
}
