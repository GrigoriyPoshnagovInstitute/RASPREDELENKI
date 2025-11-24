package workflows

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// шеф выполняет когда юзер удален

func UserDeletedWorkflow(ctx workflow.Context, userID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting UserDeletedWorkflow", "UserID", userID)

	err := workflow.Sleep(ctx, 30*time.Second)
	if err != nil {
		return err
	}

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})

	logger.Info("Executing Hard Delete", "UserID", userID)
	err = workflow.ExecuteActivity(ctx, userServiceActivities.HardDeleteUser, userID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to hard delete user", "Error", err)
		return err
	}

	return nil
}
