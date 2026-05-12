package policy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/platform/auditlog"
	"hydroponic-backend/internal/platform/di"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes registers all policy module routes.
func RegisterRoutes(r *gin.RouterGroup, deps di.Deps) {
	h := NewHandler(deps.MySQL)

	// Start policy auto-scheduler (event-driven + timer-scan)
	NewScheduler(deps.MySQL, deps.EventHub, deps.MQTT, deps.Log).Start()

	pol := r.Group("/policies")
	// ControlPolicy CRUD - Admin/Operator for writes, all roles for reads
	pol.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), withPolicyAudit(deps.MySQL, "CREATE_RULE", false, h.CreatePolicy))
	pol.POST("/full", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), withPolicyAudit(deps.MySQL, "CREATE_RULE", false, h.CreatePolicyWithNested))
	pol.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListPolicies)
	pol.GET("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetPolicy)
	pol.PUT("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), withPolicyAudit(deps.MySQL, "UPDATE_RULE", true, h.UpdatePolicy))
	pol.DELETE("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), withPolicyAudit(deps.MySQL, "DELETE_RULE", true, h.DeletePolicy))

	// Publish and Archive actions
	pol.POST("/:id/publish", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.PublishPolicy)
	pol.POST("/:id/archive", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.ArchivePolicy)

	// Policy manual execution
	pol.POST("/:id/execute", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.ExecutePolicy)
	pol.GET("/:id/executions", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListExecutionsByPolicy)

	// PolicyCondition nested resources
	pol.GET("/:id/conditions", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListConditions)
	pol.POST("/:id/conditions", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateCondition)
	pol.GET("/:id/conditions/:conditionId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetCondition)
	pol.PUT("/:id/conditions/:conditionId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdateCondition)
	pol.DELETE("/:id/conditions/:conditionId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.DeleteCondition)

	// PolicyTarget nested resources
	pol.GET("/:id/targets", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListTargets)
	pol.POST("/:id/targets", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateTarget)
	pol.GET("/:id/targets/:targetId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetTarget)
	pol.PUT("/:id/targets/:targetId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdateTarget)
	pol.DELETE("/:id/targets/:targetId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.DeleteTarget)

	// PolicyExecution - global list and detail
	r.GET("/policy-executions", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListExecutions)
	r.GET("/policy-executions/:executionId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetExecution)
}

type auditCaptureWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *auditCaptureWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func withPolicyAudit(db *gorm.DB, action string, usePathID bool, next gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqBody, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewReader(reqBody))

		writer := &auditCaptureWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = writer
		next(c)

		if c.Writer.Status() != http.StatusOK {
			return
		}

		targetID := extractPolicyTargetID(c, writer.body.Bytes(), usePathID)
		if targetID == nil {
			return
		}

		_ = auditlog.WriteEntry(db, auditlog.Entry{
			UserID:     currentUserID(c),
			Action:     action,
			TargetType: "RULE",
			TargetID:   targetID,
			Detail:     decodeJSONBody(reqBody),
			RequestID:  c.GetString("request_id"),
		})
	}
}

func extractPolicyTargetID(c *gin.Context, responseBody []byte, usePathID bool) *uint64 {
	if usePathID {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			return nil
		}
		return &id
	}

	var envelope struct {
		Data struct {
			ID uint64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(responseBody, &envelope); err != nil || envelope.Data.ID == 0 {
		return nil
	}
	return &envelope.Data.ID
}

func decodeJSONBody(body []byte) interface{} {
	if len(body) == 0 {
		return nil
	}
	var payload interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return string(body)
	}
	return payload
}
